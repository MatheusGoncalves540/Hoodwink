package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EngineVersion acompanha game/ENGINE_VERSION (a v1 está congelada em
// server/replay/engines/v1). O make engine-bump atualiza este valor e o
// make engine-check falha se ele divergir do arquivo.
const EngineVersion = 1

// InitReplayRecorder liga o finalizador de replay ao WithRoom. O envio HTTP fica no
// services (não no redisFuncs) para não criar ciclo de import com o backend.
func InitReplayRecorder(backend *BackendService) {
	redisFuncs.ReplayFinalize = func(ctx context.Context, rdb *redis.Client, room *rooms.Room) {
		if err := FinalizeReplay(ctx, backend, rdb, room); err != nil {
			utils.LogError("falha ao finalizar replay: " + err.Error())
		}
	}
}

// FinalizeReplay monta o .hwr, envia ao backend e só então limpa a lista. Se o envio
// falhar, a lista permanece (TTL de 6h) e o replay não é gerado.
func FinalizeReplay(ctx context.Context, backend *BackendService, rdb *redis.Client, room *rooms.Room) error {
	hwr, err := buildReplay(ctx, rdb, room)
	if err != nil {
		return err
	}

	postCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// player_ids é indexado pelo Player.Index do cabeçalho (ordem lexicográfica),
	// que é o índice usado pelo vencedor e pelo backend.
	playerIDs := playerIDsByReplayIndex(room)
	upload := struct {
		Data      string   `json:"data"`
		RoomID    string   `json:"room_id"`
		PlayerIDs []string `json:"player_ids"`
	}{
		Data:      base64.StdEncoding.EncodeToString(hwr),
		RoomID:    room.ID,
		PlayerIDs: playerIDs,
	}
	body, err := json.Marshal(upload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(postCtx, http.MethodPost, strings.TrimRight(backend.url, "/")+"/internal/replays", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", backend.apiKey)

	resp, err := backend.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("backend respondeu %d ao enviar replay", resp.StatusCode)
	}

	// Só limpa depois do envio confirmado.
	return redisFuncs.ClearReplayList(ctx, rdb, room.MatchID)
}

// buildReplay lê as entradas do Redis e monta o .hwr (sem assinatura).
func buildReplay(ctx context.Context, rdb *redis.Client, room *rooms.Room) ([]byte, error) {
	recorded, err := redisFuncs.ReadReplayEntries(ctx, rdb, room.MatchID)
	if err != nil {
		return nil, err
	}
	if len(recorded) == 0 {
		return nil, fmt.Errorf("replay sem entradas")
	}

	matchID, err := uuid.Parse(room.MatchID)
	if err != nil {
		return nil, fmt.Errorf("matchId inválido: %w", err)
	}

	r := &format.Replay{
		FormatVersion: format.FormatVersion,
		EngineVersion: EngineVersion,
		MatchID:       matchID,
		Seed:          [16]byte(room.Seed),
		StartedAtMs:   room.StartTime.UnixMilli(),
		Mode:          string(room.Rules),
	}

	// Jogadores na ordem de MoveTime (para o runner reconstruir MoveTime denso),
	// mas o índice de cada um é a ordem lexicográfica dos UUIDs: é a mesma ordem
	// de SortedPlayerIDs usada pelos laços do motor e pelos payloads/ator.
	roster := playerIDsByMoveTime(room)
	idToIndex := playerIndexMap(room)
	r.Players = make([]format.Player, 0, len(roster))
	for _, id := range roster {
		r.Players = append(r.Players, format.Player{
			Index: idToIndex[id],
			Name:  truncateName(room.Players[id].Name),
		})
	}

	// Entradas.
	prevAt := r.StartedAtMs
	prevConnected := map[string]bool{}
	for _, e := range recorded {
		delta := e.AtMs - prevAt
		if delta < 0 {
			delta = 0
		}
		prevAt = e.AtMs

		entry := format.Entry{
			DeltaMs: delta,
			Type:    entryType(e.Type),
			Player:  format.NoPlayer,
			Payload: rewritePlayerIDs(e.Payload, idToIndex),
		}
		if e.Player != "" {
			if idx, ok := idToIndex[e.Player]; ok {
				entry.Player = idx
			}
		}
		if connectedChanged(prevConnected, e.Connected) {
			entry.Connected = connectedBitmap(idToIndex, len(room.Players), e.Connected)
		}
		prevConnected = connectedSet(e.Connected)

		r.Entries = append(r.Entries, entry)
	}

	// Entrada final (fim).
	winner := uint8(format.NoPlayer)
	if room.WinnerID != "" {
		if idx, ok := idToIndex[room.WinnerID]; ok {
			winner = idx
		}
	}
	r.Entries = append(r.Entries, format.Entry{
		DeltaMs: 0,
		Type:    format.EntryEnd,
		Player:  winner,
	})

	// Checkpoints: a cada 20 entradas e no fim.
	r.Checkpoints = buildCheckpoints(recorded, len(r.Entries), redisFuncs.ReplayCheckpoint(room))

	return format.Encode(r, nil)
}

var replayEntryTypes = map[string]format.EntryType{
	"start":      format.EntryStart,
	"play":       format.EntryPlay,
	"tick":       format.EntryTick,
	"connect":    format.EntryConnect,
	"disconnect": format.EntryDisconnect,
	"join":       format.EntryJoin,
	"leave":      format.EntryLeave,
}

func entryType(s string) format.EntryType {
	if t, ok := replayEntryTypes[s]; ok {
		return t
	}
	return format.EntryPlay
}

// truncateName corta o nome em 32 bytes sem partir um rune UTF-8.
func truncateName(name string) string {
	if name == "" {
		return "?"
	}
	if len(name) <= format.MaxPlayerNameBytes {
		return name
	}
	cut := name[:format.MaxPlayerNameBytes]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	if cut == "" {
		return "?"
	}
	return cut
}

func connectedChanged(prev map[string]bool, curr []string) bool {
	if len(prev) != len(curr) {
		return true
	}
	for _, id := range curr {
		if !prev[id] {
			return true
		}
	}
	return false
}

func connectedSet(ids []string) map[string]bool {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// rewritePlayerIDs troca qualquer string igual a um UUID de jogador pelo
// índice correspondente ("0".."N-1"). O runner reconstrói a sala com esses ids.
func rewritePlayerIDs(payload []byte, idToIndex map[string]uint8) []byte {
	if len(payload) == 0 {
		return payload
	}
	var v any
	if err := json.Unmarshal(payload, &v); err != nil {
		return payload
	}
	rewritePlayerIDsValue(v, idToIndex)
	out, err := json.Marshal(v)
	if err != nil {
		return payload
	}
	return out
}

func rewritePlayerIDsValue(v any, idToIndex map[string]uint8) {
	stride := func(val any) (string, bool) {
		s, ok := val.(string)
		if !ok {
			return "", false
		}
		idx, ok := idToIndex[s]
		if !ok {
			return "", false
		}
		return strconv.Itoa(int(idx)), true
	}

	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if idx, ok := stride(val); ok {
				x[k] = idx
				continue
			}
			rewritePlayerIDsValue(val, idToIndex)
		}
	case []any:
		for i, val := range x {
			if idx, ok := stride(val); ok {
				x[i] = idx
				continue
			}
			rewritePlayerIDsValue(val, idToIndex)
		}
	}
}

// playerIndexMap devolve o índice de replay de cada jogador na ordem
// lexicográfica dos ids (a mesma de SortedPlayerIDs no motor).
func playerIndexMap(room *rooms.Room) map[string]uint8 {
	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	m := make(map[string]uint8, len(ids))
	for i, id := range ids {
		m[id] = uint8(i)
	}
	return m
}

// playerIDsByReplayIndex devolve os ids na ordem do índice de replay
// (lexicográfica), casando com format.Player.Index e com o vencedor.
func playerIDsByReplayIndex(room *rooms.Room) []string {
	idToIndex := playerIndexMap(room)
	ids := make([]string, len(room.Players))
	for id, idx := range idToIndex {
		ids[idx] = id
	}
	return ids
}

// playerIDsByMoveTime devolve os ids em ordem de MoveTime (que nunca repete).
func playerIDsByMoveTime(room *rooms.Room) []string {
	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := room.Players[ids[i]], room.Players[ids[j]]
		if a.MoveTime != b.MoveTime {
			return a.MoveTime < b.MoveTime
		}
		return ids[i] < ids[j]
	})
	return ids
}

// connectedBitmap monta o bitmap de conectados na ordem dos índices de replay.
func connectedBitmap(idToIndex map[string]uint8, playerCount int, ids []string) []byte {
	bitmapLen := (playerCount + 7) / 8
	bm := make([]byte, bitmapLen)
	for _, id := range ids {
		if idx, ok := idToIndex[id]; ok {
			bm[idx/8] |= 1 << (idx % 8)
		}
	}
	return bm
}

func buildCheckpoints(recorded []redisFuncs.ReplayEntry, n int, finalCP [8]byte) []format.Checkpoint {
	count := (n + format.CheckpointInterval - 1) / format.CheckpointInterval
	cps := make([]format.Checkpoint, 0, count)
	for i := 0; i < count; i++ {
		pos := (i + 1) * format.CheckpointInterval
		if pos >= n {
			pos = n
		}
		if pos == n {
			cps = append(cps, format.Checkpoint(finalCP))
			continue
		}
		// pos-1 é o índice 0-based da entrada gravada correspondente.
		cps = append(cps, format.Checkpoint(recorded[pos-1].CP))
	}
	return cps
}
