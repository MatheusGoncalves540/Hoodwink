package redisFuncs

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// canonicalReplayRoom devolve uma cópia da sala contendo apenas o que o replay
// reproduz. O hash não pode depender de dados de exibição nem dos UUIDs reais:
//
//   - id do jogador vira o índice na ordem lexicográfica, que é a mesma ordem de
//     SortedPlayerIDs usada pelos laços do motor (e a usada nos payloads/ator);
//   - MoveTime vira a ordem densa (0..N-1), preservando todas as comparações;
//   - NextMoveTime, nome/senha/ID da sala, chat, maxPlayers e created são
//     normalizados;
//   - tempos são truncados em milissegundos, a precisão que o .hwr carrega.
func canonicalReplayRoom(room *rooms.Room) *rooms.Room {
	clone := room.Clone()
	clone.ID = clone.MatchID
	clone.Name = ""
	clone.Password = ""
	clone.MaxPlayers = 0
	clone.CustomMatch = false
	clone.Chat = nil
	clone.NonReplayable = false

	ids := make([]string, 0, len(clone.Players))
	for id := range clone.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	idToIndex := make(map[string]string, len(ids))
	for i, id := range ids {
		idToIndex[id] = strconv.Itoa(i)
	}

	moveOrder := append([]string(nil), ids...)
	sort.Slice(moveOrder, func(i, j int) bool {
		a, b := clone.Players[moveOrder[i]], clone.Players[moveOrder[j]]
		if a.MoveTime != b.MoveTime {
			return a.MoveTime < b.MoveTime
		}
		return moveOrder[i] < moveOrder[j]
	})
	moveRank := make(map[string]int, len(ids))
	for i, id := range moveOrder {
		moveRank[id] = i
	}

	players := make(map[string]*players.Player, len(ids))
	for id, p := range clone.Players {
		p.Name = ""
		p.Id = idToIndex[id]
		p.MoveTime = moveRank[id]
		players[idToIndex[id]] = p
	}
	clone.Players = players
	clone.NextMoveTime = len(ids)

	clone.StartTime = truncateReplayTime(clone.StartTime)
	clone.Created = clone.StartTime
	if clone.GameEvent != nil {
		clone.GameEvent.ExpiresAt = truncateReplayTime(clone.GameEvent.ExpiresAt)
	}

	return clone
}

// canonicalReplayJSON serializa a sala canônica e troca qualquer string igual a
// um UUID de jogador pelo índice correspondente. Isso cobre currentPlayer,
// winnerId, GameEvent, efeitos e eventos de apresentação.
func canonicalReplayJSON(room *rooms.Room) ([]byte, error) {
	clone := canonicalReplayRoom(room)
	data, err := json.Marshal(clone)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	idToIndex := make(map[string]string, len(ids))
	for i, id := range ids {
		idToIndex[id] = strconv.Itoa(i)
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	rewriteReplayStrings(v, idToIndex)
	return json.Marshal(v)
}

// rewriteReplayStrings troca, recursivamente, valores string pelos índices.
func rewriteReplayStrings(v any, idToIndex map[string]string) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if s, ok := val.(string); ok {
				if idx, ok := idToIndex[s]; ok {
					x[k] = idx
					continue
				}
			}
			rewriteReplayStrings(val, idToIndex)
		}
	case []any:
		for i, val := range x {
			if s, ok := val.(string); ok {
				if idx, ok := idToIndex[s]; ok {
					x[i] = idx
					continue
				}
			}
			rewriteReplayStrings(val, idToIndex)
		}
	}
}

func truncateReplayTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.Truncate(time.Millisecond)
}
