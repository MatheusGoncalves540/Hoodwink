package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/MatheusGoncalves540/Hoodwink-replay/runner"
)

// buildReplay monta um replay v1 sintético e calcula os checkpoints dirigindo um
// Runner. O hash vem do próprio runner (canonicalHash, que normaliza os campos
// fora do jogo), então o teste não duplica a normalização. Divergências são
// esperadas enquanto os checkpoints ainda são zero; o estado final fica igual
// porque o runner não descarta a sala ao divergir.
func buildReplay(t *testing.T) (*format.Replay, string) {
	t.Helper()

	entries := []format.Entry{{DeltaMs: 0, Type: format.EntryStart, Player: format.NoPlayer}}
	delta := int64(400)
	for i := 0; i < 15; i++ {
		entries = append(entries, format.Entry{
			DeltaMs: delta,
			Type:    format.EntryPlay,
			Player:  uint8(i % 2),
			Payload: []byte(`{"type":"TAKE_COINS"}`),
		})
		delta += 120
		entries = append(entries, format.Entry{DeltaMs: delta, Type: format.EntryTick, Player: format.NoPlayer})
		delta += 80
	}

	replay := &format.Replay{
		EngineVersion: 1,
		MatchID:       [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		Seed:          [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "ClassicRules",
		Players: []format.Player{
			{Index: 0, Name: "ana"},
			{Index: 1, Name: "bia"},
		},
		Entries: entries,
	}

	positions := checkpointPositionsFor(len(entries))
	replay.Checkpoints = make([]format.Checkpoint, len(positions))

	r, err := runner.New(replay, runner.Options{})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	defer r.Close()

	next := 0
	for range entries {
		_ = r.Step() // divergência de checkpoint é esperada nesta primeira passada
		if next < len(positions) && r.Position() == positions[next] {
			hash := r.Hash()
			if len(hash) < 2*format.CheckpointSize {
				t.Fatalf("hash canônico curto no passo %d: %q", r.Position(), hash)
			}
			b, err := hex.DecodeString(hash[:2*format.CheckpointSize])
			if err != nil {
				t.Fatalf("hash canônico inválido no passo %d: %v", r.Position(), err)
			}
			copy(replay.Checkpoints[next][:], b)
			next++
		}
	}
	if next != len(positions) {
		t.Fatalf("faltaram %d checkpoints", len(positions)-next)
	}

	return replay, r.Hash()
}

// checkpointPositionsFor replica format.CheckpointInterval (pacote format não
// exporta a lista de posições, só a constante).
func checkpointPositionsFor(n int) []int {
	positions := make([]int, 0, n/format.CheckpointInterval+1)
	for p := format.CheckpointInterval; p < n; p += format.CheckpointInterval {
		positions = append(positions, p)
	}
	return append(positions, n)
}

func TestSessionPercorreReplay(t *testing.T) {
	replay, wantFinal := buildReplay(t)
	data, err := format.Encode(replay, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	s := &session{}
	if !s.load(data) {
		t.Fatalf("load: %s", s.err())
	}

	steps := 0
	for s.step() {
		steps++
	}
	if steps != len(replay.Entries) {
		t.Fatalf("%d passos, quer %d", steps, len(replay.Entries))
	}
	if s.err() != "" {
		t.Fatalf("erro ao andar até o fim: %s", s.err())
	}
	if got := s.hash(); got != wantFinal {
		t.Fatalf("hash final %s, quer %s", got, wantFinal)
	}
	if s.state() == "" {
		t.Fatal("estado vazio depois do fim")
	}
	if info := s.info(); !strings.Contains(info, `"done":true`) {
		t.Fatalf("info sem done: %s", info)
	}
}

func TestSessionSeekVoltaAtras(t *testing.T) {
	replay, wantFinal := buildReplay(t)
	data, err := format.Encode(replay, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	s := &session{}
	if !s.load(data) {
		t.Fatalf("load: %s", s.err())
	}
	if !s.seek(len(replay.Entries)) {
		t.Fatalf("seek: %s", s.err())
	}
	if !strings.Contains(s.info(), `"position":`+strconv.Itoa(len(replay.Entries))) {
		t.Fatalf("posição errada: %s", s.info())
	}

	back := len(replay.Entries) - 5
	if !s.seek(back) {
		t.Fatalf("seek(%d): %s", back, s.err())
	}
	for s.step() {
	}
	if got := s.hash(); got != wantFinal {
		t.Fatalf("hash após seek %s, quer %s", got, wantFinal)
	}
}

func TestSessionRejeitaBytesInvalidos(t *testing.T) {
	s := &session{}
	if s.load([]byte("nao-e-um-hwr")) {
		t.Fatal("load aceitou lixo")
	}
	if s.err() == "" {
		t.Fatal("esperava mensagem de erro")
	}
	if s.state() != "" || s.hash() != "" {
		t.Fatal("estado/hash não deveriam existir depois de load inválido")
	}
}

func TestSessionKidEVerificacaoDaAssinatura(t *testing.T) {
	replay, _ := buildReplay(t)
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	data, err := format.Encode(replay, format.NewEd25519Signer("replay-v1", priv))
	if err != nil {
		t.Fatalf("Encode assinado: %v", err)
	}

	s := &session{}
	if !s.load(data) {
		t.Fatalf("load: %s", s.err())
	}
	if got := s.kid(); got != "replay-v1" {
		t.Fatalf("kid = %q, quer replay-v1", got)
	}
	if err := s.verify(pub); err != nil {
		t.Fatalf("verify com a chave certa: %v", err)
	}

	// Chave pública errada e bytes de tamanho inválido não podem passar.
	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey (outra): %v", err)
	}
	if err := s.verify(otherPub); !errors.Is(err, format.ErrInvalidSignature) {
		t.Fatalf("verify com outra chave = %v, quer ErrInvalidSignature", err)
	}
	if err := s.verify([]byte{1, 2, 3}); err == nil {
		t.Fatal("verify aceitou chave de 3 bytes")
	}

	// 1 byte trocado na assinatura não muda o decode (o load é cego à assinatura),
	// mas verify tem que recusar.
	tampered := append([]byte(nil), data...)
	tampered[len(tampered)-1] ^= 0xFF
	sTampered := &session{}
	if !sTampered.load(tampered) {
		t.Fatalf("load do assinado adulterado: %s", sTampered.err())
	}
	if err := sTampered.verify(pub); !errors.Is(err, format.ErrInvalidSignature) {
		t.Fatalf("verify da assinatura trocada = %v, quer ErrInvalidSignature", err)
	}
}

func TestSessionSemAssinaturaNaoVerifica(t *testing.T) {
	replay, _ := buildReplay(t)
	data, err := format.Encode(replay, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	s := &session{}
	if err := s.verify(make([]byte, ed25519.PublicKeySize)); err == nil {
		t.Fatal("verify antes do load deveria falhar")
	}
	if !s.load(data) {
		t.Fatalf("load: %s", s.err())
	}
	if got := s.kid(); got != "" {
		t.Fatalf("kid = %q, quer vazio", got)
	}
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if err := s.verify(pub); !errors.Is(err, format.ErrInvalid) {
		t.Fatalf("verify sem assinatura = %v, quer ErrInvalid", err)
	}
}

func TestSessionRecusaVersaoDiferenteDoBuild(t *testing.T) {
	replay, _ := buildReplay(t)
	data, err := format.Encode(replay, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	old := buildEngineVersion
	buildEngineVersion = "2"
	defer func() { buildEngineVersion = old }()

	s := &session{}
	if s.load(data) {
		t.Fatal("load aceitou replay v1 num motor v2")
	}
	if !strings.Contains(s.err(), "versão 1") {
		t.Fatalf("mensagem sem a versão do replay: %s", s.err())
	}

	buildEngineVersion = "1"
	if !s.load(data) {
		t.Fatalf("load recusou a versão certa: %s", s.err())
	}
}
