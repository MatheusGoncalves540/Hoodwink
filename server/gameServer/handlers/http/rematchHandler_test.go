package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func rematchTestHandler(t *testing.T) (*HTTPHandler, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	gameRules, err := rules.LoadRules("../../game/rules/modes/ClassicRules.yaml")
	if err != nil {
		t.Fatalf("carregar regras: %v", err)
	}
	reg := rules.NewRegistry()
	reg.Register(string(structs.ClassicRules), gameRules)

	return &HTTPHandler{RoomService: services.NewRoomService(rdb), RulesRegistry: reg}, rdb
}

func seedFinishedRoom(t *testing.T, rdb *redis.Client, over bool) *rooms.Room {
	t.Helper()
	old := &rooms.Room{
		ID:         "OLDRM1",
		Name:       "Quinta dos mentirosos",
		MaxPlayers: 6,
		GameOver:   over,
		Players: map[string]*players.Player{
			"p1": {Id: "p1", Name: "Ana"},
			"p2": {Id: "p2", Name: "Bia"},
		},
	}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := rdb.Set(context.Background(), "room:"+old.ID, data, time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	return old
}

func rematchRequest(roomID, playerID string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/rematch/"+roomID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("RoomId", roomID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, contextKeys.PlayerContextKey, &endpointStructures.ClaimsBackend{Id: playerID})
	return req.WithContext(ctx)
}

// TestRematchConcurrentCallsCriamUmaSalaSo cobre o "blindado": todos clicam ao
// mesmo tempo e recebem a mesma sala, sem duplicar.
func TestRematchConcurrentCallsCriamUmaSalaSo(t *testing.T) {
	handler, rdb := rematchTestHandler(t)
	seedFinishedRoom(t, rdb, true)

	const calls = 8
	var wg sync.WaitGroup
	ids := make([]string, calls)
	statuses := make([]int, calls)
	for i := 0; i < calls; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rr := httptest.NewRecorder()
			handler.RematchRoom(rdb)(rr, rematchRequest("OLDRM1", "p1"))
			statuses[i] = rr.Code
			var body endpointStructures.CreateRoomResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err == nil {
				ids[i] = body.RoomId
			}
		}(i)
	}
	wg.Wait()

	for i := range statuses {
		if statuses[i] != http.StatusOK {
			t.Fatalf("chamada %d: status %d, quer 200", i, statuses[i])
		}
		if ids[i] == "" || ids[i] != ids[0] {
			t.Fatalf("chamada %d: sala %q, quer a mesma %q", i, ids[i], ids[0])
		}
	}

	// a sala antiga continua; exatamente uma sala nova foi criada
	keys, err := rdb.Keys(context.Background(), "room:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("salas no Redis = %v, quer a antiga e uma nova", keys)
	}
}

func TestRematchRecusaQuemNaoJogou(t *testing.T) {
	handler, rdb := rematchTestHandler(t)
	seedFinishedRoom(t, rdb, true)

	rr := httptest.NewRecorder()
	handler.RematchRoom(rdb)(rr, rematchRequest("OLDRM1", "intruso"))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, quer 403", rr.Code)
	}
}

func TestRematchRecusaPartidaEmAndamento(t *testing.T) {
	handler, rdb := rematchTestHandler(t)
	seedFinishedRoom(t, rdb, false)

	rr := httptest.NewRecorder()
	handler.RematchRoom(rdb)(rr, rematchRequest("OLDRM1", "p1"))
	if rr.Code != http.StatusConflict {
		t.Fatalf("status %d, quer 409", rr.Code)
	}
}
