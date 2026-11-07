package redisFuncs

import (
	"context"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestPublishRoomUpdatePublicoTemChatPrivadoNao(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pubsub := client.Subscribe(ctx, "room:R1:broadcast")
	defer pubsub.Close()
	if _, err := pubsub.Receive(ctx); err != nil {
		t.Fatalf("assinar canal: %v", err)
	}

	room := &rooms.Room{
		ID:         "R1",
		MatchID:    "match-1",
		Name:       "Sala de Ana",
		MaxPlayers: 2,
		Players: map[string]*players.Player{
			"p1": {Id: "p1", Name: "Ana", Cards: []structs.Card{{Name: structs.CardAssassin, Index: 0}}},
			"p2": {Id: "p2", Name: "Bia", Cards: []structs.Card{{Name: structs.CardGuardian, Index: 0}}},
		},
		Deck: []structs.CardName{structs.CardAssassin, structs.CardGuardian, structs.CardKamikaze},
		Chat: []structs.ChatMessage{{PlayerName: "Ana", Timestamp: time.Unix(0, 0).UTC(), Message: "oi"}},
	}

	if err := PublishRoomUpdate(ctx, client, room, true, nil); err != nil {
		t.Fatalf("PublishRoomUpdate: %v", err)
	}

	// 1 mensagem pública + 1 privada por jogador.
	public := receiveBroadcast(t, ctx, pubsub)
	if !public.ToAll {
		t.Fatalf("primeira mensagem deveria ser pública, veio toAll=false")
	}
	publicData := decodeData(t, public)
	assertChatLen(t, publicData, 1)
	if _, ok := publicData["deck"].(float64); !ok {
		t.Fatalf("deck público = %T, quer número (contagem, sem baralho)", publicData["deck"])
	}

	privateByPlayer := map[string]map[string]any{}
	for i := 0; i < len(room.Players); i++ {
		msg := receiveBroadcast(t, ctx, pubsub)
		if msg.ToAll {
			t.Fatal("mensagem privada veio com toAll=true")
		}
		if len(msg.ConfidencialPlayerIds) != 1 {
			t.Fatalf("ids confidenciais = %v, quer exatamente 1", msg.ConfidencialPlayerIds)
		}
		data := decodeData(t, msg)
		assertChatLen(t, data, 0)
		if _, ok := data["deck"].(float64); !ok {
			t.Fatalf("deck privado = %T, quer número (contagem, sem baralho)", data["deck"])
		}
		privateByPlayer[msg.ConfidencialPlayerIds[0]] = data
	}

	if len(privateByPlayer) != 2 {
		t.Fatalf("mensagens privadas = %d jogadores, quer 2", len(privateByPlayer))
	}
	for playerId := range room.Players {
		data, ok := privateByPlayer[playerId]
		if !ok {
			t.Fatalf("sem mensagem privada para %s", playerId)
		}
		cards := data["players"].(map[string]any)[playerId].(map[string]any)["cards"].([]any)
		card := cards[0].(map[string]any)
		if card["name"] == nil {
			t.Fatalf("carta privada de %s sem nome: %v", playerId, card)
		}
	}
}

func receiveBroadcast(t *testing.T, ctx context.Context, pubsub *redis.PubSub) *structs.BroadcastMessage {
	t.Helper()
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("receber mensagem: %v", err)
	}
	parsed, err := structs.ParseBroadcastMessage([]byte(msg.Payload))
	if err != nil {
		t.Fatalf("parse broadcast: %v", err)
	}
	return parsed
}

func decodeData(t *testing.T, msg *structs.BroadcastMessage) map[string]any {
	t.Helper()
	data, ok := msg.Data.(map[string]any)
	if !ok {
		t.Fatalf("data = %T, quer objeto JSON", msg.Data)
	}
	return data
}

func assertChatLen(t *testing.T, data map[string]any, want int) {
	t.Helper()
	chat, _ := data["chat"].([]any)
	if len(chat) != want {
		t.Fatalf("chat = %v, quer %d mensagem(ns)", data["chat"], want)
	}
}
