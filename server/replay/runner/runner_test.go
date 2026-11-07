package runner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	roomsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
)

// syntheticEntries monta uma sequência determinística: início + rodadas de
// TAKE_COINS (jogadores 0..3) intercaladas com ticks. Jogadas fora do turno são
// recusadas pelo motor de forma determinística, então a sequência é estável.
func syntheticEntries() []format.Entry {
	entries := []format.Entry{{DeltaMs: 0, Type: format.EntryStart, Player: format.NoPlayer}}
	delta := int64(800)
	for range 5 {
		for i := 0; i < 4; i++ {
			entries = append(entries, format.Entry{
				DeltaMs: delta,
				Type:    format.EntryPlay,
				Player:  uint8(i),
				Payload: []byte(`{"type":"TAKE_COINS"}`),
			})
			delta += 300
			entries = append(entries, format.Entry{DeltaMs: delta, Type: format.EntryTick, Player: format.NoPlayer})
			delta += 200
		}
	}
	return entries
}

// buildSyntheticReplay monta um replay v1 sintético e calcula os checkpoints
// dirigindo a sala com a mesma lógica do Runner (applyEntry). Devolve também os
// hashes canônicos após cada entrada, para comparação.
func buildSyntheticReplay(t *testing.T) (*format.Replay, []string) {
	t.Helper()

	replay := &format.Replay{
		EngineVersion: 1,
		MatchID:       [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		Seed:          [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "ClassicRules",
		Players: []format.Player{
			{Index: 0, Name: "ana"},
			{Index: 1, Name: "bia"},
			{Index: 2, Name: "cris"},
			{Index: 3, Name: "davi"},
		},
		Entries: syntheticEntries(),
	}

	r, err := New(replay, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	var checkpoints []format.Checkpoint
	hashes := make([]string, 0, len(replay.Entries))

	for r.pos < len(replay.Entries) {
		if _, err := r.applyEntry(replay.Entries[r.pos]); err != nil {
			t.Fatalf("aplicar entrada %d: %v", r.pos, err)
		}
		r.pos++

		h, err := canonicalHash(r.room)
		if err != nil {
			t.Fatalf("hash no passo %d: %v", r.pos, err)
		}
		hashes = append(hashes, h)

		if r.nextCheckpointIdx < len(r.checkpointPositions) &&
			r.pos == r.checkpointPositions[r.nextCheckpointIdx] {
			c, err := checkpointOf(r.room)
			if err != nil {
				t.Fatalf("checkpoint no passo %d: %v", r.pos, err)
			}
			checkpoints = append(checkpoints, c)
			r.nextCheckpointIdx++
		}
	}

	replay.Checkpoints = checkpoints
	return replay, hashes
}

func TestSyntheticReplayChegaAoMesmoHash(t *testing.T) {
	replay, wantHashes := buildSyntheticReplay(t)

	// Ida e volta no formato, como um arquivo .hwr real.
	data, err := format.Encode(replay, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := format.Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	r, err := New(decoded, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	var gotHashes []string
	for r.pos < len(decoded.Entries) {
		if err := r.Step(); err != nil {
			t.Fatalf("Step %d: %v", r.pos, err)
		}
		gotHashes = append(gotHashes, r.Hash())
	}

	if len(gotHashes) != len(wantHashes) {
		t.Fatalf("%d hashes, quer %d", len(gotHashes), len(wantHashes))
	}
	for i := range wantHashes {
		if gotHashes[i] != wantHashes[i] {
			t.Fatalf("hash divergiu no passo %d", i+1)
		}
	}
}

func TestVersaoDesconhecida(t *testing.T) {
	replay, _ := buildSyntheticReplay(t)
	replay.EngineVersion = 99

	_, err := New(replay, Options{})
	if err == nil || !strings.Contains(err.Error(), "desconhecida") {
		t.Fatalf("esperava erro de versão desconhecida, veio %v", err)
	}
}

func TestCheckpointAdulteradoAcusaOPasso(t *testing.T) {
	replay, _ := buildSyntheticReplay(t)

	// Adultera o primeiro checkpoint do arquivo.
	replay.Checkpoints[0][0] ^= 0xFF

	r, err := New(replay, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	var div *Divergence
	for r.pos < len(replay.Entries) {
		err := r.Step()
		if err != nil {
			if !errors.As(err, &div) {
				t.Fatalf("esperava Divergence, veio %v", err)
			}
			break
		}
	}
	if div == nil {
		t.Fatal("esperava divergência de checkpoint")
	}

	want := r.checkpointPositions[0]
	if div.Step != want {
		t.Fatalf("divergiu no passo %d, quer %d", div.Step, want)
	}
}

func TestPanicViraErro(t *testing.T) {
	replay, _ := buildSyntheticReplay(t)

	r, err := New(replay, Options{Sink: panicSink{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	// A primeira entrada (start) chama Broadcast via StartGameEffect, que panica.
	if err := r.Step(); err == nil {
		t.Fatal("esperava erro de panic, veio nil")
	}
}

func TestSeekVoltaAtras(t *testing.T) {
	replay, _ := buildSyntheticReplay(t)

	r, err := New(replay, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	total := len(replay.Entries)
	for r.pos < total {
		if err := r.Step(); err != nil {
			t.Fatalf("Step %d: %v", r.pos, err)
		}
	}
	finalHash := r.Hash()

	back := total - 30
	if back < 0 {
		back = 0
	}
	if err := r.Seek(back); err != nil {
		t.Fatalf("Seek(%d): %v", back, err)
	}
	if r.pos != back {
		t.Fatalf("posição após Seek = %d, quer %d", r.pos, back)
	}

	for r.pos < total {
		if err := r.Step(); err != nil {
			t.Fatalf("Step após Seek %d: %v", r.pos, err)
		}
	}
	if r.Hash() != finalHash {
		t.Fatal("hash após voltar 30 passos e andar de novo divergiu")
	}
}

func TestJogadaRecusadaNaoMudaEstado(t *testing.T) {
	replay := &format.Replay{
		EngineVersion: 1,
		MatchID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Seed:          [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "ClassicRules",
		Players: []format.Player{
			{Index: 0, Name: "ana"},
			{Index: 1, Name: "bia"},
		},
		Entries: []format.Entry{
			{DeltaMs: 0, Type: format.EntryStart, Player: format.NoPlayer},
			// Fora do turno: o atual é "0", então "1" é recusado.
			{DeltaMs: 1000, Type: format.EntryPlay, Player: 1, Payload: []byte(`{"type":"TAKE_COINS"}`)},
		},
	}

	r, err := New(replay, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer r.Close()

	if _, err := r.applyEntry(replay.Entries[0]); err != nil {
		t.Fatalf("start: %v", err)
	}
	r.pos = 1
	before, err := canonicalHash(r.room)
	if err != nil {
		t.Fatalf("hash antes: %v", err)
	}

	applied, err := r.applyEntry(replay.Entries[1])
	if err != nil {
		t.Fatalf("jogada: %v", err)
	}
	if applied {
		t.Fatal("jogada fora do turno deveria ser recusada")
	}
	after, err := canonicalHash(r.room)
	if err != nil {
		t.Fatalf("hash depois: %v", err)
	}
	if before != after {
		t.Fatal("jogada recusada mudou o estado")
	}
}

// panicSink panica no Broadcast, para exercitar a recuperação de panic do Runner.
type panicSink struct{}

func (panicSink) Broadcast(context.Context, *roomsv1.Room, *roomsv1.Room, []string) error {
	panic("boom")
}

func (panicSink) RegistryTimeout(context.Context, *roomsv1.Room, time.Time) {}
func (panicSink) ClearTimeout(context.Context, *roomsv1.Room)               {}
