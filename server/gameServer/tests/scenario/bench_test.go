package scenario

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// go test -bench . -benchmem ./tests/scenario
//
// Custos do caminho quente do servidor, sem Redis (a rede fica fora da conta):
//   - ParsePlayerPlay: cada mensagem de cada jogador;
//   - ContestRound: declarar, contestar, escolher a carta e resolver (inclui os round-trips JSON = ler/gravar a sala);
//   - Broadcast10: montar e serializar a versão de cada um dos 10 jogadores, o que acontece a cada mudança.

func BenchmarkParsePlayerPlay(b *testing.B) {
	raw := []byte(`{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":1,"targetAllyCardIndex":0}}`)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := structs.ParsePlayerPlay(raw, "ana"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContestRound(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		tb := newTable(b, seat("pop", 7, structs.CardInvestor, structs.CardRebel), seat("por", 2, structs.CardKamikaze, structs.CardGravedigger))
		tb.mustPlay("pop", `{"type":"INVESTOR"}`)
		tb.mustPlay("por", `{"type":"CONTEST"}`)
		tb.tickUntil(structs.EventContestPenalty)
		tb.mustPlay("pop", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":1}}`)
		tb.settle()
	}
}

func BenchmarkBroadcast10(b *testing.B) {
	seats := make([]seatSpec, 10)
	for i := range seats {
		seats[i] = seat(fmt.Sprint("p", i), 5, structs.CardAssassin, structs.CardGuardian)
	}
	tb := newTable(b, seats...)
	b.ReportAllocs()
	for b.Loop() {
		public := tb.room.GetPublicRoomForUpdates()
		for _, p := range tb.room.Players {
			private := public.Clone()
			private.Players[p.Id] = p.GetPrivatePlayerForPublicUpdates()
			if _, err := json.Marshal(private); err != nil {
				b.Fatal(err)
			}
		}
	}
}
