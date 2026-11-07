package redisFuncs

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// Cause identifica por que a sala está sendo alterada. Cada caminho de mudança
// passa pelo WithRoom com a causa certa (o T8 grava a entrada com este valor).
type Cause string

const (
	CausePlay       Cause = "play"
	CauseTick       Cause = "tick"
	CauseConnect    Cause = "connect"
	CauseDisconnect Cause = "disconnect"
	CauseJoin       Cause = "join"
	CauseLeave      Cause = "leave"
	CauseStart      Cause = "start"
)

// roomMutexes serializa, dentro desta instância, tudo que lê-modifica-grava uma sala.
// O lock do Redis é por instância (reentrante), então sozinho não separava a jogada de um
// jogador do tick do gameProcessor: os dois carregavam a sala, e o último a salvar apagava
// o outro (ex.: uma contestação feita no fim da janela sumia).
// ponytail: um mutex por sala, nunca removido (bytes por sala criada); limpar ao apagar a sala se pesar.
var roomMutexes sync.Map

// WithRoom carrega a versão mais nova da sala já com o mutex dela e o lock distribuído,
// preenche Connected a partir do Redis e roda fn. O lock é adquirido ANTES do LoadRoom
// para não ler um estado velho gravado por outro pod, renovado em segundo plano e solto
// ao final (mesmo em panic, via defer).
//
// fn muta a sala em memória e devolve aplicado (true quando algo mudou). Só então a
// sala é persistida; tick que não rodou (fn devolve false) não grava nada.
func WithRoom(ctx context.Context, rdb *redis.Client, roomID string, cause Cause, fn func(*rooms.Room) bool, replayPlay ...ReplayPlayHint) error {
	m, _ := roomMutexes.LoadOrStore(roomID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	instanceID := utils.GetInstanceID()
	ttl := time.Duration(utils.MustEnvInt("ROOM_LOCK_TTL_MS", 3000)) * time.Millisecond

	lock, err := AcquireRoomLock(ctx, rdb, roomID, instanceID, ttl)
	if err != nil {
		return err
	}
	if lock == nil {
		lock, err = waitRoomLock(ctx, rdb, roomID, instanceID, ttl)
		if err != nil {
			return err
		}
	}
	defer lock.Release()

	room, err := LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}
	room.AttachLock(lock)

	// Captura o relógio UMA vez antes do fn: durante a chamada a sala usa esse valor
	// fixo (determinismo). O T8 passa a gravar o mesmo valor na entrada do replay.
	now := time.Now().UTC()
	room.SetNow(now)
	defer room.ClearNow()

	// Preenche o conjunto Connected a partir do registro no Redis antes do fn; o motor
	// nunca consulta o Redis para isso (lê room.Connected).
	if err := fillConnected(ctx, rdb, room); err != nil {
		return err
	}

	wasStarted := !room.StartTime.IsZero()

	aplicado := fn(room)

	// Gravação de replay: a entrada vai no MESMO MULTI do save (nunca depois).
	// Tick que não rodou e NonReplayable não gravam nada; jogada recusada grava
	// mesmo assim (o replay aplica num clone e descarta).
	var hint *ReplayPlayHint
	if len(replayPlay) > 0 {
		hint = &replayPlay[0]
	}
	entry, recordEntry := replayEntryToRecord(cause, aplicado, room, now, wasStarted, hint)
	switch {
	case recordEntry && aplicado:
		if err := SaveRoomAndReplay(ctx, rdb, room, entry); err != nil {
			return err
		}
	case recordEntry:
		// Jogada recusada: grava só a entrada, sem persistir o estado.
		if err := appendReplayEntry(ctx, rdb, room.MatchID, entry); err != nil {
			return err
		}
	case aplicado:
		if err := SaveRoom(ctx, rdb, room); err != nil {
			return err
		}
	}

	// Fim de jogo: monta o .hwr e envia ao backend, fora do lock (goroutine).
	if room.GameOver && !room.NonReplayable && ReplayEnabled() && ReplayFinalize != nil {
		go ReplayFinalize(context.Background(), rdb, room.Clone())
	}

	return nil
}

// fillConnected preenche room.Connected com os ids (ordenados) dos jogadores da sala
// que estão registrados no Redis nesta sala. Usa MGET (1 round-trip para todos os
// jogadores + 1 para as instâncias) em vez de um GET por jogador.
func fillConnected(ctx context.Context, rdb *redis.Client, room *rooms.Room) error {
	if len(room.Players) == 0 {
		room.SetConnected(nil)
		return nil
	}

	playerIds := make([]string, 0, len(room.Players))
	playerKeys := make([]string, 0, len(room.Players))
	for playerId := range room.Players {
		playerIds = append(playerIds, playerId)
		playerKeys = append(playerKeys, "player:"+playerId+":room")
	}

	// Registro de todos os jogadores de uma vez (o formato é "roomId:instanceId").
	values, err := rdb.MGet(ctx, playerKeys...).Result()
	if err != nil {
		return err
	}

	type registration struct {
		roomId     string
		instanceId string
	}
	registrations := make(map[string]registration, len(values))
	instanceIds := make([]string, 0, len(values))
	seenInstances := make(map[string]struct{}, len(values))
	invalid := make([]string, 0)

	for i, value := range values {
		if value == nil {
			continue // jogador não está em nenhuma sala
		}
		raw, ok := value.(string)
		if !ok {
			invalid = append(invalid, playerIds[i])
			continue
		}

		parts := strings.Split(raw, ":")
		if len(parts) != 2 {
			// Formato antigo ou inválido: remove o registro, como o caminho por jogador.
			invalid = append(invalid, playerIds[i])
			continue
		}

		registrations[playerIds[i]] = registration{roomId: parts[0], instanceId: parts[1]}
		if _, seen := seenInstances[parts[1]]; !seen {
			seenInstances[parts[1]] = struct{}{}
			instanceIds = append(instanceIds, parts[1])
		}
	}

	// Liveness das instâncias: uma chave por instância, não por jogador.
	alive := make(map[string]bool, len(instanceIds))
	if len(instanceIds) > 0 {
		aliveKeys := make([]string, len(instanceIds))
		for i, instanceId := range instanceIds {
			aliveKeys[i] = "instance:" + instanceId + ":alive"
		}
		aliveValues, err := rdb.MGet(ctx, aliveKeys...).Result()
		if err != nil {
			return err
		}
		for i, value := range aliveValues {
			alive[instanceIds[i]] = value != nil
		}
	}

	connected := make([]string, 0, len(room.Players))
	for _, playerId := range playerIds {
		reg, ok := registrations[playerId]
		if !ok {
			continue
		}
		if !alive[reg.instanceId] {
			// Instância morreu: remove o registro do jogador.
			invalid = append(invalid, playerId)
			continue
		}
		if reg.roomId == room.ID {
			connected = append(connected, playerId)
		}
	}

	if len(invalid) > 0 {
		pipe := rdb.Pipeline()
		for _, playerId := range invalid {
			pipe.Del(ctx, "player:"+playerId+":room")
		}
		_, _ = pipe.Exec(ctx)
	}

	sort.Strings(connected)
	room.SetConnected(connected)
	return nil
}

// waitRoomLock tenta de novo a cada 50ms por até 2s enquanto a sala está com outro pod.
func waitRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string, ttl time.Duration) (*RoomLock, error) {
	const (
		retryInterval = 50 * time.Millisecond
		maxWait       = 2 * time.Second
	)

	start := time.Now()
	loggedSlow := false

	for {
		lock, err := AcquireRoomLock(ctx, rdb, roomID, instanceID, ttl)
		if err != nil {
			return nil, err
		}
		if lock != nil {
			return lock, nil
		}

		if time.Since(start) > 500*time.Millisecond && !loggedSlow {
			utils.LogDebug("esperando lock da sala " + roomID)
			loggedSlow = true
		}
		if time.Since(start) >= maxWait {
			return nil, fmt.Errorf("não foi possível adquirir lock da sala %s (em uso por outro pod)", roomID)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}
