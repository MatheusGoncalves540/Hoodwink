package rooms_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func TestSeedEscondidaDeLogEPublico(t *testing.T) {
	seed := rooms.Seed{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	hexSeed := hex.EncodeToString(seed[:])

	if got := seed.String(); got != "[oculto]" {
		t.Fatalf("String() = %q, quer [oculto]", got)
	}
	if got := seed.GoString(); got != "[oculto]" {
		t.Fatalf("GoString() = %q, quer [oculto]", got)
	}
	if got := seed.LogValue().String(); got != "[oculto]" {
		t.Fatalf("LogValue() = %q, quer [oculto]", got)
	}

	room := &rooms.Room{ID: "r1", MatchID: "m1", Seed: seed, Rng: rooms.NewRngState(seed)}

	// fmt %+v e %#v não podem vazar a semente (usam String/GoString).
	if got := fmt.Sprintf("%+v", room); strings.Contains(got, hexSeed) {
		t.Fatalf("%%+v vazou a semente: %s", got)
	}
	if got := fmt.Sprintf("%#v", room); strings.Contains(got, hexSeed) {
		t.Fatalf("%%#v vazou a semente: %s", got)
	}

	// O JSON público (cliente) não pode conter semente nem rng; matchId é público de propósito.
	pub, err := json.Marshal(room.GetPublicRoomForUpdates())
	if err != nil {
		t.Fatalf("serializar sala pública: %v", err)
	}
	pubStr := string(pub)
	if strings.Contains(pubStr, hexSeed) {
		t.Fatalf("JSON público vazou a semente: %s", pubStr)
	}
	for _, forbidden := range []string{"seed", "rng"} {
		if strings.Contains(pubStr, forbidden) {
			t.Fatalf("JSON público contém %q: %s", forbidden, pubStr)
		}
	}

	// O JSON da sala (Redis) preserva a semente para o re-seed.
	full, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("serializar sala: %v", err)
	}
	if !strings.Contains(string(full), hexSeed) {
		t.Fatalf("JSON da sala deveria conter a semente (para o Redis), mas não contém")
	}
}

func TestRngStateSobreviveAoClone(t *testing.T) {
	seed := rooms.Seed{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16}
	room := &rooms.Room{Seed: seed, Rng: rooms.NewRngState(seed), Players: map[string]*players.Player{}}

	// Avança o gerador antes do clone.
	for range 3 {
		room.Rand().IntN(1000)
	}

	clone := room.Clone()

	// Depois do clone, os dois continuam da mesma posição.
	for range 50 {
		a := room.Rand().IntN(1 << 30)
		b := clone.Rand().IntN(1 << 30)
		if a != b {
			t.Fatalf("rng divergiu após Clone: original=%d clone=%d", a, b)
		}
	}
}

func TestSortedPlayerIDsOrdena(t *testing.T) {
	room := &rooms.Room{Players: map[string]*players.Player{
		"bia": {}, "ana": {}, "cris": {}, "davi": {},
	}}
	got := room.SortedPlayerIDs()
	want := []string{"ana", "bia", "cris", "davi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SortedPlayerIDs = %v, quer %v", got, want)
	}
}

func TestRoomNowFixaEClear(t *testing.T) {
	room := &rooms.Room{}
	fixed := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	room.SetNow(fixed)
	if got := room.Now(); !got.Equal(fixed) {
		t.Fatalf("Now() = %v, quer %v (relógio fixo)", got, fixed)
	}

	room.ClearNow()
	if got := room.Now(); !got.After(fixed) {
		t.Fatalf("Now() após ClearNow = %v, deveria ser depois de %v", got, fixed)
	}
}

func TestNewSeedUnica(t *testing.T) {
	a := rooms.NewSeed()
	b := rooms.NewSeed()
	if a == b {
		t.Fatal("duas sementes de crypto/rand não deveriam ser iguais")
	}
}
