package redisFuncs

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

const (
	replayListTTL   = 6 * time.Hour
	replayKeyPrefix = "replay:"
)

// ReplayEnabled informa se a gravação de replay está ligada (padrão true).
// REPLAY_ENABLED=false desliga a gravação sem deploy de código.
func ReplayEnabled() bool {
	return os.Getenv("REPLAY_ENABLED") != "false"
}

// ReplayListKey é a chave da lista de entradas do replay de uma partida no Redis.
func ReplayListKey(matchID string) string {
	return replayKeyPrefix + matchID
}

// ReplayFinalize é chamado (em goroutine) quando a partida termina, para montar o
// .hwr e enviá-lo ao backend. É injetado pelo pacote services (que cuida do HTTP),
// evitando ciclo de import redisFuncs <-> services.
var ReplayFinalize func(ctx context.Context, rdb *redis.Client, room *rooms.Room)

// ReplayPlayHint captura o ator e o payload JSON original de uma jogada para a
// entrada de replay. É passada ao WithRoom pelo caminho de jogada.
type ReplayPlayHint struct {
	Actor   string
	Payload []byte
}

// ReplayEntry é uma entrada de replay gravada no Redis. Não contém a semente: o
// .hwr é montado só no fim, a partir do estado final da sala.
type ReplayEntry struct {
	AtMs      int64    `json:"at"`
	Type      string   `json:"type"`
	Applied   bool     `json:"applied,omitempty"`
	Player    string   `json:"player,omitempty"`
	Connected []string `json:"connected,omitempty"`
	Payload   []byte   `json:"payload,omitempty"`
	CP        [8]byte  `json:"cp"`
}

// ReplayCheckpoint devolve os primeiros 8 bytes do sha256 do estado canônico
// (JSON) da sala. O runner usa o mesmo cálculo para conferir divergências.
// A canonicalização (IDs de jogador, MoveTime, campos de exibição e precisão
// dos tempos) fica em replayCanonical.go para os dois lados baterem.
func ReplayCheckpoint(room *rooms.Room) [8]byte {
	data, err := canonicalReplayJSON(room)
	if err != nil {
		return [8]byte{}
	}
	sum := sha256.Sum256(data)
	var out [8]byte
	copy(out[:], sum[:8])
	return out
}

// replayEntryToRecord monta a entrada a gravar, ou ok=false quando nada é gravado
// (jogo ainda não começou, tick que não rodou, NonReplayable ou replay desligado).
func replayEntryToRecord(cause Cause, aplicado bool, room *rooms.Room, now time.Time, wasStarted bool, hint *ReplayPlayHint) ([]byte, bool) {
	if !ReplayEnabled() || room.NonReplayable {
		return nil, false
	}

	isStarted := !room.StartTime.IsZero()

	// Entrada 0 = início (transição de não-iniciado para iniciado).
	if !wasStarted && isStarted {
		return marshalReplayEntry(ReplayEntry{
			AtMs:      now.UnixMilli(),
			Type:      "start",
			Connected: append([]string(nil), room.Connected...),
			CP:        ReplayCheckpoint(room),
		})
	}

	// Antes do início nada é gravado.
	if !isStarted {
		return nil, false
	}

	var actor string
	var payload []byte
	if hint != nil {
		actor = hint.Actor
		payload = hint.Payload
	}

	switch cause {
	case CausePlay:
		// Jogada é gravada mesmo quando recusada (o replay aplica num clone e descarta).
		return marshalReplayEntry(ReplayEntry{
			AtMs:      now.UnixMilli(),
			Type:      string(cause),
			Applied:   aplicado,
			Player:    actor,
			Connected: append([]string(nil), room.Connected...),
			Payload:   payload,
			CP:        ReplayCheckpoint(room),
		})
	case CauseTick:
		// Tick que não rodou (fn false) não grava nada.
		if !aplicado {
			return nil, false
		}
		return marshalReplayEntry(ReplayEntry{
			AtMs:      now.UnixMilli(),
			Type:      string(cause),
			Connected: append([]string(nil), room.Connected...),
			CP:        ReplayCheckpoint(room),
		})
	default:
		// connect/disconnect/join/leave/start: só mudanças aplicadas.
		if !aplicado {
			return nil, false
		}
		return marshalReplayEntry(ReplayEntry{
			AtMs:      now.UnixMilli(),
			Type:      string(cause),
			Player:    actor,
			Connected: append([]string(nil), room.Connected...),
			CP:        ReplayCheckpoint(room),
		})
	}
}

func marshalReplayEntry(e ReplayEntry) ([]byte, bool) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, false
	}
	return data, true
}

// appendReplayEntry grava a entrada de uma jogada recusada (sem SET room).
func appendReplayEntry(ctx context.Context, rdb *redis.Client, matchID string, entry []byte) error {
	if rdb == nil {
		return nil
	}
	replayKey := ReplayListKey(matchID)
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.RPush(ctx, replayKey, entry)
		pipe.Expire(ctx, replayKey, replayListTTL)
		return nil
	})
	return err
}

// ReadReplayEntries lê todas as entradas de replay de uma partida, em ordem.
func ReadReplayEntries(ctx context.Context, rdb *redis.Client, matchID string) ([]ReplayEntry, error) {
	vals, err := rdb.LRange(ctx, ReplayListKey(matchID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]ReplayEntry, 0, len(vals))
	for _, v := range vals {
		var e ReplayEntry
		if err := json.Unmarshal([]byte(v), &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ClearReplayList remove a lista de entradas após o envio do .hwr.
func ClearReplayList(ctx context.Context, rdb *redis.Client, matchID string) error {
	return rdb.Del(ctx, ReplayListKey(matchID)).Err()
}
