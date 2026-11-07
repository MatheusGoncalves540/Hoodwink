package scenario

import (
	"fmt"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// Fórmulas das regras v2 (plans/02-regras-v2.md; client/web/src/game/pricing.ts diz o mesmo à mesa).
// Escritas aqui de novo, de propósito: se o servidor mudar o cálculo, o teste acusa a divergência.
func clamp(v, lo, hi int) int { return max(lo, min(hi, v)) }

var taxes = []int{-5, -2, 0, 1, 3, 5}

// Declara a carta no turno de ana e confere quanto saiu do bolso.
func TestPricesPaidOnDeclaration(t *testing.T) {
	cases := []struct {
		name  string
		card  structs.CardName
		raw   string
		price func(tax int) int
	}{
		{"Assassino 3+t, mín 2", structs.CardAssassin, `{"type":"ASSASSIN","payload":{"targetPlayer":"bia","targetCardIndex":0}}`, func(t int) int { return max(2, 3+t) }},
		{"Clarividente 4+t, mín 2", structs.CardClairvoyant, `{"type":"CLAIRVOYANT","payload":{"targetPlayer":"bia","targetCardIndex":0,"showToAllPlayers":false}}`, func(t int) int { return max(2, 4+t) }},
		{"Guardiã no turno 6+t, mín 5", structs.CardGuardian, `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`, func(t int) int { return max(5, 6+t) }},
		{"Matar 6 + t/2 (t>0) ou 6+t, mín 3", structs.CardRebel, `{"type":"KILL_CARD","payload":{"targetPlayer":"bia","targetCardIndex":0}}`, func(t int) int {
			if t > 0 {
				return 6 + t/2
			}
			return max(3, 6+t)
		}},
		{"Kamikaze é de graça", structs.CardKamikaze, `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":0}}`, func(int) int { return 0 }},
	}
	for _, c := range cases {
		for _, tax := range taxes {
			t.Run(fmt.Sprintf("%s/taxa %d", c.name, tax), func(t *testing.T) {
				tb := newTable(t, seat("ana", 15, c.card, structs.CardRebel), seat("bia", 0, structs.CardRebel, structs.CardRebel))
				tb.room.Tax = tax
				tb.mustPlay("ana", c.raw)
				tb.wantCoins("ana", 15-c.price(tax))
			})
		}
	}
}

func TestPricesGains(t *testing.T) {
	for _, tax := range taxes {
		t.Run(fmt.Sprintf("Trilionário 3+t, mín 1/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("ana", 0, structs.CardTrillionaire, structs.CardRebel), seat("bia", 0, structs.CardRebel, structs.CardRebel))
			tb.room.Tax = tax
			tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
			tb.settle()
			tb.wantCoins("ana", max(1, 3+tax))
		})

		t.Run(fmt.Sprintf("Trapaceiro tira clamp(2+t,2,5) e fica com a metade para cima/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("ana", 0, structs.CardTrickster, structs.CardRebel), seat("bia", 10, structs.CardRebel, structs.CardRebel))
			tb.room.Tax = tax
			tb.mustPlay("ana", `{"type":"TRICKSTER","payload":{"targetPlayer":"bia"}}`)
			tb.settle()
			x := clamp(2+tax, 2, 5)
			tb.wantCoins("bia", 10-x)
			tb.wantCoins("ana", (x+1)/2)
		})

		t.Run(fmt.Sprintf("Pegar moeda 1+t, taxa negativa não reduz/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("ana", 2, structs.CardRebel, structs.CardRebel), seat("bia", 0, structs.CardRebel, structs.CardRebel))
			tb.room.Tax = tax
			tb.mustPlay("ana", `{"type":"TAKE_COINS"}`)
			tb.settle()
			tb.wantCoins("ana", 2+max(1, 1+tax))
		})
	}
}

func TestPricesReactions(t *testing.T) {
	for _, tax := range taxes {
		t.Run(fmt.Sprintf("Guardiã na hora do ataque 5+t, mín 4/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("bia", 15, structs.CardAssassin, structs.CardRebel), seat("ana", 15, structs.CardGuardian, structs.CardRebel))
			tb.room.Tax = tax
			tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
			tb.tickUntil(structs.EventCardKilled)
			tb.mustPlay("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
			tb.wantCoins("ana", 15-max(4, 5+tax))
		})

		t.Run(fmt.Sprintf("Guardiã contra Clarividente 4+t, mín 2/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("bia", 15, structs.CardClairvoyant, structs.CardRebel), seat("ana", 15, structs.CardGuardian, structs.CardRebel))
			tb.room.Tax = tax
			tb.mustPlay("bia", `{"type":"CLAIRVOYANT","payload":{"targetPlayer":"ana","targetCardIndex":0,"showToAllPlayers":true}}`)
			tb.mustPlay("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
			tb.wantCoins("ana", 15-max(2, 4+tax))
		})
	}
}

func TestInvestorCapFollowsTax(t *testing.T) {
	for _, tax := range taxes {
		t.Run(fmt.Sprintf("teto clamp(4+t,2,6)/taxa %d", tax), func(t *testing.T) {
			tb := newTable(t, seat("ana", 20, structs.CardInvestor, structs.CardRebel), seat("bia", 0, structs.CardRebel, structs.CardRebel))
			tb.room.Tax = tax
			limit := clamp(4+tax, 2, 6)
			for range limit {
				tb.p("ana").AddInvestment(6)
			}
			tb.mustReject("ana", `{"type":"INVESTOR"}`)
			tb.p("ana").Investments = tb.p("ana").Investments[1:]
			tb.mustPlay("ana", `{"type":"INVESTOR"}`)
			tb.wantCoins("ana", 19)
		})
	}
}

func TestCannotPayWhatYouDoNotHave(t *testing.T) {
	tb := newTable(t, seat("ana", 2, structs.CardAssassin, structs.CardRebel), seat("bia", 0, structs.CardRebel, structs.CardRebel))
	tb.mustReject("ana", `{"type":"ASSASSIN","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
	tb.mustReject("ana", `{"type":"KILL_CARD","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
	tb.p("ana").Coins = 0
	tb.mustReject("ana", `{"type":"INVESTOR"}`)
	tb.wantCoins("ana", 0)
}
