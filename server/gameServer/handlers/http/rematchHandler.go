package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

const (
	// rematchMappingTTL guarda a sala da revanche enquanto os jogadores clicam.
	rematchMappingTTL = 5 * time.Minute
	// rematchWait é quanto um pedido espera a revanche que outro pedido está criando.
	rematchWait = 5 * time.Second
)

func rematchKey(roomID string) string { return "rematch:" + roomID }

// RematchRoom devolve a sala da revanche de uma partida encerrada. A primeira
// pessoa que chama cria a sala; as outras recebem a mesma, mesmo em paralelo
// (o mapeamento é criado sob o lock da sala).
func (h *HTTPHandler) RematchRoom(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(contextKeys.PlayerContextKey).(*endpointStructures.ClaimsBackend)
		if !ok || claims == nil {
			utils.SendError(w, "dados de player inválidos no token", http.StatusUnauthorized)
			return
		}

		oldID := strings.TrimSpace(chi.URLParam(r, "RoomId"))
		old, err := redisFuncs.LoadRoom(r.Context(), rdb, oldID)
		if err != nil {
			utils.SendError(w, "Sala não encontrada", http.StatusNotFound)
			return
		}
		if _, err := old.GetPlayer(claims.Id); err != nil {
			utils.SendError(w, "Você não jogou nesta sala", http.StatusForbidden)
			return
		}
		if !old.GameOver {
			utils.SendError(w, "A partida ainda está em andamento", http.StatusConflict)
			return
		}

		newID, err := rematchRoom(r, rdb, h, old)
		if err != nil {
			utils.LogError(fmt.Errorf("rematch da sala %s: %w", old.ID, err))
			utils.SendError(w, "Não foi possível preparar a revanche", http.StatusInternalServerError)
			return
		}

		utils.SendJSON(w, http.StatusOK, endpointStructures.CreateRoomResponse{
			RoomId: newID,
			Msg:    "Revanche pronta.",
		})
	}
}

func rematchRoom(r *http.Request, rdb *redis.Client, h *HTTPHandler, old *rooms.Room) (string, error) {
	ctx := r.Context()
	key := rematchKey(old.ID)
	if id := readyRematch(ctx, rdb, key); id != "" {
		return id, nil
	}

	lock, err := redisFuncs.AcquireRoomLock(ctx, rdb, key, utils.GetInstanceID(), 5*time.Second)
	if err != nil {
		return "", err
	}
	if lock == nil {
		// Outra requisição está criando a sala agora: espera a chave aparecer.
		return waitRematch(ctx, rdb, key)
	}
	defer func() { _ = lock.Release() }()

	// Reconfere sob o lock: outra pessoa pode ter criado antes de a chave aparecer.
	if id := readyRematch(ctx, rdb, key); id != "" {
		return id, nil
	}

	old.Name = strings.TrimSpace(old.Name)
	if len([]rune(old.Name)) < 3 {
		old.Name = "Revanche"
	}
	if old.MaxPlayers < 2 || old.MaxPlayers > 10 {
		old.MaxPlayers = 10
	}

	newRoom, err := h.RoomService.CreateNewRoom(r, h.RulesRegistry, endpointStructures.CreateRoomRequest{
		RoomName:   old.Name,
		Password:   old.Password,
		MaxPlayers: old.MaxPlayers,
	}, true)
	if err != nil {
		return "", err
	}

	// A sala nasce com TTL curto (tempo de abrir o WebSocket). A revanche pode
	// esperar alguns segundos até o segundo jogador clicar.
	if err := rdb.Expire(ctx, "room:"+newRoom.ID, rematchMappingTTL).Err(); err != nil {
		return "", err
	}
	if err := rdb.Set(ctx, key, newRoom.ID, rematchMappingTTL).Err(); err != nil {
		return "", err
	}
	return newRoom.ID, nil
}

// readyRematch devolve a sala da revanche já criada. Se a sala não existe mais,
// descarta o mapeamento para a próxima chamada criar outra.
func readyRematch(ctx context.Context, rdb *redis.Client, key string) string {
	id, err := rdb.Get(ctx, key).Result()
	if err != nil || id == "" {
		return ""
	}
	exists, err := rdb.Exists(ctx, "room:"+id).Result()
	if err != nil || exists == 0 {
		rdb.Del(ctx, key)
		return ""
	}
	return id
}

func waitRematch(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	deadline := time.Now().Add(rematchWait)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
		if id := readyRematch(ctx, rdb, key); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("outra requisição ainda está criando a revanche")
}
