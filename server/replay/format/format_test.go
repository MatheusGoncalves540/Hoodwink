package format

import (
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"
)

// validReplay monta um replay sintético válido com nPlayers jogadores.
func validReplay(nPlayers int) *Replay {
	players := make([]Player, nPlayers)
	for i := range players {
		players[i] = Player{Index: uint8(i), Name: fmt.Sprintf("jogador%d", i)}
	}
	return &Replay{
		EngineVersion: 1,
		MatchID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Seed:          [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "classic",
		Players:       players,
		Entries: []Entry{
			{DeltaMs: 0, Type: EntryStart, Player: NoPlayer},
			{DeltaMs: 1000, Type: EntryPlay, Player: 0, Payload: []byte(`{"type":"TAKE_COINS"}`)},
			{DeltaMs: 2000, Type: EntryTick, Player: NoPlayer},
			{DeltaMs: 3000, Type: EntryEnd, Player: 0},
		},
	}
}

func fillCheckpoints(r *Replay) {
	n := checkpointCountFor(len(r.Entries))
	r.Checkpoints = make([]Checkpoint, n)
	for i := range r.Checkpoints {
		r.Checkpoints[i] = Checkpoint{byte(i), 1, 2, 3, 4, 5, 6, 7}
	}
}

func TestRoundTripUnsigned(t *testing.T) {
	r := validReplay(2)
	fillCheckpoints(r)

	data, err := Encode(r, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Official {
		t.Fatal("arquivo sem assinatura deveria ser Official=false")
	}

	assertEqual(t, r, got)
}

func TestRoundTripSigned(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewEd25519Signer("kid-teste", priv)
	lookup := StaticKeys(map[string]ed25519.PublicKey{"kid-teste": pub})

	r := validReplay(3)
	fillCheckpoints(r)

	data, err := Encode(r, signer)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Decode(data, lookup)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !got.Official {
		t.Fatal("assinatura válida deveria ser Official=true")
	}
	if got.Kid != "kid-teste" {
		t.Fatalf("kid = %q, quer kid-teste", got.Kid)
	}
	if len(got.Signature) != SignatureSize {
		t.Fatalf("assinatura com %d bytes, quer %d", len(got.Signature), SignatureSize)
	}

	assertEqual(t, r, got)

	if err := Verify(data, lookup); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func assertEqual(t *testing.T, want, got *Replay) {
	t.Helper()
	if got.FormatVersion != FormatVersion {
		t.Fatalf("FormatVersion = %d, quer %d", got.FormatVersion, FormatVersion)
	}
	if got.EngineVersion != want.EngineVersion {
		t.Fatalf("EngineVersion = %d, quer %d", got.EngineVersion, want.EngineVersion)
	}
	if got.MatchID != want.MatchID {
		t.Fatalf("MatchID = %x, quer %x", got.MatchID, want.MatchID)
	}
	if got.Seed != want.Seed {
		t.Fatalf("Seed = %x, quer %x", got.Seed, want.Seed)
	}
	if got.StartedAtMs != want.StartedAtMs {
		t.Fatalf("StartedAtMs = %d, quer %d", got.StartedAtMs, want.StartedAtMs)
	}
	if got.Mode != want.Mode {
		t.Fatalf("Mode = %q, quer %q", got.Mode, want.Mode)
	}
	if len(got.Players) != len(want.Players) {
		t.Fatalf("len(Players) = %d, quer %d", len(got.Players), len(want.Players))
	}
	for i := range want.Players {
		if got.Players[i] != want.Players[i] {
			t.Fatalf("Players[%d] = %+v, quer %+v", i, got.Players[i], want.Players[i])
		}
	}
	if len(got.Entries) != len(want.Entries) {
		t.Fatalf("len(Entries) = %d, quer %d", len(got.Entries), len(want.Entries))
	}
	for i := range want.Entries {
		if got.Entries[i].DeltaMs != want.Entries[i].DeltaMs ||
			got.Entries[i].Type != want.Entries[i].Type ||
			got.Entries[i].Player != want.Entries[i].Player ||
			!bytes.Equal(got.Entries[i].Connected, want.Entries[i].Connected) ||
			!bytes.Equal(got.Entries[i].Payload, want.Entries[i].Payload) {
			t.Fatalf("Entries[%d] = %+v, quer %+v", i, got.Entries[i], want.Entries[i])
		}
	}
	if len(got.Checkpoints) != len(want.Checkpoints) {
		t.Fatalf("len(Checkpoints) = %d, quer %d", len(got.Checkpoints), len(want.Checkpoints))
	}
	for i := range want.Checkpoints {
		if got.Checkpoints[i] != want.Checkpoints[i] {
			t.Fatalf("Checkpoints[%d] = %x, quer %x", i, got.Checkpoints[i], want.Checkpoints[i])
		}
	}
}

func TestTruncado(t *testing.T) {
	r := validReplay(2)
	fillCheckpoints(r)
	data, err := Encode(r, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, cut := range []int{1, 3, len(data) - 1} {
		if _, err := Decode(data[:cut], nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("truncado em %d deveria dar ErrInvalid, veio %v", cut, err)
		}
	}
}

func TestArquivoAcimaDe256KB(t *testing.T) {
	big := make([]byte, MaxFileSize+1)
	if _, err := Decode(big, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("arquivo de %d bytes deveria dar ErrInvalid, veio %v", len(big), err)
	}
}

func TestMagicErrado(t *testing.T) {
	body := buildBody(t, bodySpec{magic: "XXXX"})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("magic errado deveria dar ErrInvalid, veio %v", err)
	}
}

func TestOnzeJogadores(t *testing.T) {
	body := buildBody(t, bodySpec{players: 11})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("11 jogadores deveria dar ErrInvalid, veio %v", err)
	}
}

func TestNome40Bytes(t *testing.T) {
	body := buildBody(t, bodySpec{nameLen: 40})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nome de 40 bytes deveria dar ErrInvalid, veio %v", err)
	}
}

func Test5001Entradas(t *testing.T) {
	body := buildBody(t, bodySpec{entries: MaxEntries + 1})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("5001 entradas deveria dar ErrInvalid, veio %v", err)
	}
}

func TestPayload3KB(t *testing.T) {
	body := buildBody(t, bodySpec{entries: 3, payloadLen: 3 * 1024})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("payload de 3 KB deveria dar ErrInvalid, veio %v", err)
	}
}

func TestBombaDeGzip(t *testing.T) {
	bomb := bytes.Repeat([]byte("A"), MaxDecompressedSize+100)
	data := wrapGzip(t, bomb)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bomba de gzip deveria dar ErrInvalid, veio %v", err)
	}
}

func TestDeltaNegativo(t *testing.T) {
	body := buildBody(t, bodySpec{entries: 2, firstDelta: -1})
	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("delta negativo deveria dar ErrInvalid, veio %v", err)
	}
}

func TestBitTrocadoNoCabecalho(t *testing.T) {
	r := validReplay(2)
	fillCheckpoints(r)
	body, err := encodeBody(r, "")
	if err != nil {
		t.Fatal(err)
	}
	body[len(Magic)] ^= 0x01 // troca 1 bit na versão do formato

	data := wrapGzip(t, body)
	if _, err := Decode(data, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bit trocado no cabeçalho deveria dar ErrInvalid, veio %v", err)
	}
}

func TestKidDesconhecido(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewEd25519Signer("kid-desconhecido", priv)
	lookup := StaticKeys(map[string]ed25519.PublicKey{"outro-kid": pub})

	r := validReplay(2)
	fillCheckpoints(r)
	data, err := Encode(r, signer)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Decode(data, lookup)
	if !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("kid desconhecido deveria dar ErrUnknownKey, veio %v", err)
	}
	if got == nil || got.Official {
		t.Fatal("kid desconhecido deveria retornar replay com Official=false")
	}
}

func TestAssinaturaInvalida(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewEd25519Signer("kid-ok", priv)
	lookup := StaticKeys(map[string]ed25519.PublicKey{"kid-ok": pub})

	r := validReplay(2)
	fillCheckpoints(r)
	data, err := Encode(r, signer)
	if err != nil {
		t.Fatal(err)
	}

	// Corrompe a assinatura (últimos 64 bytes).
	data[len(data)-1] ^= 0x01

	got, err := Decode(data, lookup)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("assinatura inválida deveria dar ErrInvalidSignature, veio %v", err)
	}
	if got == nil || got.Official {
		t.Fatal("assinatura inválida deveria retornar replay com Official=false")
	}

	if err := Verify(data, lookup); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("Verify deveria dar ErrInvalidSignature, veio %v", err)
	}
}

func TestSemLookupNaoValida(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer := NewEd25519Signer("kid-x", priv)
	r := validReplay(2)
	fillCheckpoints(r)
	data, err := Encode(r, signer)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Decode(data, nil)
	if err != nil {
		t.Fatalf("Decode sem lookup deveria devolver o replay, veio %v", err)
	}
	if got.Official {
		t.Fatal("sem lookup não deveria marcar Official=true")
	}
}

// TestTamanhoRealista garante que uma partida realista (dados sintéticos) cabe
// no alvo de 4 KB.
func TestTamanhoRealista(t *testing.T) {
	players := []Player{
		{Index: 0, Name: "ana"},
		{Index: 1, Name: "bia"},
		{Index: 2, Name: "cris"},
		{Index: 3, Name: "davi"},
	}

	entries := []Entry{{DeltaMs: 0, Type: EntryStart, Player: NoPlayer}}
	plays := []string{
		`{"type":"TAKE_COINS"}`,
		`{"type":"ASSASSIN","payload":{"targetPlayer":"bia","targetCardIndex":1}}`,
		`{"type":"INVESTOR"}`,
		`{"type":"POLITICAL","payload":{"tax":2}}`,
		`{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":0}}`,
		`{"type":"KAMIKAZE","payload":{"targetPlayer":"cris","targetCardIndex":0,"targetAllyCardIndex":1}}`,
	}
	var delta int64 = 500
	for turn := 0; turn < 60; turn++ {
		who := uint8(turn % 4)
		entries = append(entries, Entry{
			DeltaMs: delta,
			Type:    EntryPlay,
			Player:  who,
			Payload: []byte(plays[turn%len(plays)]),
		})
		delta += 400
		entries = append(entries, Entry{DeltaMs: delta, Type: EntryTick, Player: NoPlayer})
		delta += 200
	}
	entries = append(entries, Entry{DeltaMs: delta, Type: EntryEnd, Player: 3})

	r := &Replay{
		EngineVersion: 1,
		MatchID:       [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Seed:          [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "classic",
		Players:       players,
		Entries:       entries,
	}
	fillCheckpoints(r)

	data, err := Encode(r, nil)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(data) > 4096 {
		t.Fatalf("arquivo com %d bytes (alvo ≤ 4096)", len(data))
	}
	t.Logf("tamanho do replay realista: %d bytes", len(data))
}

// --- helpers para forjar corpos hostis ---

type bodySpec struct {
	magic      string
	players    int
	nameLen    int
	entries    int
	firstDelta int64
	payloadLen int
}

// buildBody monta um corpo descomprimido SEM validação, para forjar arquivos que
// violam os limites do leitor.
func buildBody(t *testing.T, s bodySpec) []byte {
	t.Helper()
	if s.magic == "" {
		s.magic = Magic
	}
	if s.players == 0 {
		s.players = 2
	}
	if s.nameLen == 0 {
		s.nameLen = 1
	}
	if s.entries == 0 {
		s.entries = 1
	}

	var b bytes.Buffer
	b.WriteString(s.magic)
	b.WriteByte(FormatVersion)
	_ = binary.Write(&b, binary.BigEndian, uint16(1)) // engine version
	b.WriteByte(0)                                    // kid vazio

	var zero [16]byte
	b.Write(zero[:])                                 // matchId
	b.Write(zero[:])                                 // seed
	_ = binary.Write(&b, binary.BigEndian, int64(0)) // startedAtMs

	b.WriteByte(byte(len("classic")))
	b.WriteString("classic")

	b.WriteByte(byte(s.players))
	for i := 0; i < s.players; i++ {
		b.WriteByte(uint8(i))
		b.WriteByte(byte(s.nameLen))
		b.Write(bytes.Repeat([]byte("x"), s.nameLen))
	}

	b.WriteByte(0) // flags

	_ = binary.Write(&b, binary.BigEndian, uint16(s.entries))
	for i := 0; i < s.entries; i++ {
		delta := int64(1)
		if i == 0 {
			delta = s.firstDelta
		}
		var varint [binary.MaxVarintLen64]byte
		n := binary.PutVarint(varint[:], delta)
		b.Write(varint[:n])

		typ := byte(EntryStart)
		player := byte(NoPlayer)
		if i > 0 && i < s.entries-1 {
			typ = byte(EntryPlay)
			player = 0
		}
		if i == s.entries-1 {
			typ = byte(EntryEnd)
			player = 0
		}
		b.WriteByte(typ)
		b.WriteByte(player)

		if typ == byte(EntryPlay) && s.payloadLen > 0 {
			b.WriteByte(entryFlagPayload)
			_ = binary.Write(&b, binary.BigEndian, uint16(s.payloadLen))
			b.Write(bytes.Repeat([]byte("p"), s.payloadLen))
		} else {
			b.WriteByte(0)
		}
	}

	want := checkpointCountFor(s.entries)
	_ = binary.Write(&b, binary.BigEndian, uint16(want))
	for i := 0; i < want; i++ {
		var c [CheckpointSize]byte
		b.Write(c[:])
	}

	return b.Bytes()
}

// wrapGzip comprime um corpo e o emoldura com o prefixo de comprimento.
func wrapGzip(t *testing.T, body []byte) []byte {
	t.Helper()
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	zw.Header.ModTime = time.Unix(0, 0)
	zw.Header.OS = 255
	if _, err := zw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	out := binary.BigEndian.AppendUint32(nil, uint32(gz.Len()))
	out = append(out, gz.Bytes()...)
	return out
}
