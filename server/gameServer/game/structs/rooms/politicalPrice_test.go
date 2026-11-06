package rooms

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

func TestPoliticalPriceDoublesAndHalvesAfterThreeTurns(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()
	room.DoubledCardValues[structs.PlayPoliticalCard] = structs.DoubledCardValues{
		TimesValueDoubled:   0,
		RoundsUntilDecrease: 0,
	}

	value, err := room.GetCardValue(registry, structs.PlayPoliticalCard, 0)
	if err != nil {
		t.Fatal(err)
	}
	if value != 2 {
		t.Fatalf("preço inicial = %d, quer 2", value)
	}

	// Cada uso dobra o preço, até o máximo de 8
	for i := 0; i < 3; i++ {
		room.MarkCardValueAsDoubled(registry, structs.PlayPoliticalCard)
		value, _ = room.GetCardValue(registry, structs.PlayPoliticalCard, 0)
	}

	if value != 8 {
		t.Fatalf("preço após 3 usos = %d, quer 8", value)
	}

	// No turno do uso o contador não decrementa
	room.DecreaseDoubledCardValuesRounds(registry)

	// Após 3 turnos sem uso, o preço cai pela metade
	for i := 0; i < 3; i++ {
		room.DecreaseDoubledCardValuesRounds(registry)
	}

	value, _ = room.GetCardValue(registry, structs.PlayPoliticalCard, 0)
	if value != 4 {
		t.Fatalf("preço após 3 turnos sem uso = %d, quer 4", value)
	}
}
