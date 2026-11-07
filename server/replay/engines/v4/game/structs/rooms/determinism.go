package rooms

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"math/rand/v2"
)

// Seed é a semente de 128 bits da partida. Ela é serializada no JSON da sala
// (Redis) para re-seed do gerador, mas o valor nunca aparece em String/GoString/
// LogValue — e, por isso, nunca vaza para log, erro ou cliente.
type Seed [16]byte

// NewSeed gera uma semente de 128 bits a partir de crypto/rand.
func NewSeed() Seed {
	var s Seed
	if _, err := cryptorand.Read(s[:]); err != nil {
		panic(fmt.Sprintf("gerar semente da sala: %v", err))
	}
	return s
}

// String esconde o valor da semente (formatação %v/%s).
func (s Seed) String() string { return "[oculto]" }

// GoString esconde o valor da semente (formatação %#v).
func (s Seed) GoString() string { return "[oculto]" }

// LogValue esconde o valor da semente no log estruturado (slog).
func (s Seed) LogValue() slog.Value { return slog.StringValue("[oculto]") }

// MarshalJSON serializa a semente em hexadecimal no JSON da sala (Redis).
func (s Seed) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(s[:]))
}

// UnmarshalJSON lê a semente em hexadecimal do JSON da sala.
func (s *Seed) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	b, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	if len(b) != len(Seed{}) {
		return fmt.Errorf("semente com tamanho inválido: %d", len(b))
	}
	copy(s[:], b)
	return nil
}

// pcgSeed deriva os dois uint64 usados como semente do PCG.
func (s Seed) pcgSeed() (uint64, uint64) {
	return binary.LittleEndian.Uint64(s[:8]), binary.LittleEndian.Uint64(s[8:])
}

// RngState mantém o gerador PCG da sala e o serializa no JSON via MarshalBinary.
// O estado avança a cada sorteio e sobrevive ao Clone (ida e volta JSON).
type RngState struct {
	rng *rand.Rand
	src *rand.PCG
}

// NewRngState cria o gerador da sala a partir da semente.
func NewRngState(seed Seed) RngState {
	hi, lo := seed.pcgSeed()
	src := rand.NewPCG(hi, lo)
	return RngState{rng: rand.New(src), src: src}
}

// Rand retorna o gerador ativo.
func (s RngState) Rand() *rand.Rand { return s.rng }

// MarshalJSON serializa o estado do PCG (base64) no JSON da sala.
func (s RngState) MarshalJSON() ([]byte, error) {
	if s.src == nil {
		return []byte("null"), nil
	}
	b, err := s.src.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

// UnmarshalJSON reconstrói o PCG a partir do estado serializado.
func (s *RngState) UnmarshalJSON(data []byte) error {
	var raw []byte
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	src := new(rand.PCG)
	if len(raw) > 0 {
		if err := src.UnmarshalBinary(raw); err != nil {
			return err
		}
	}
	s.src = src
	s.rng = rand.New(src)
	return nil
}

// Now retorna o relógio da sala: o valor fixo capturado pelo WithRoom durante a
// chamada, ou time.Now().UTC() fora dela.
func (r *Room) Now() time.Time {
	if r.now != nil {
		return *r.now
	}
	return time.Now().UTC()
}

// SetNow fixa o relógio da sala para a duração da chamada do WithRoom.
func (r *Room) SetNow(t time.Time) {
	r.now = &t
}

// ClearNow solta o relógio fixo (volta a usar time.Now().UTC()).
func (r *Room) ClearNow() {
	r.now = nil
}

// Rand retorna o gerador da sala, criando-o a partir da semente/estado se preciso.
func (r *Room) Rand() *rand.Rand {
	if r.Rng.rng == nil {
		if r.Rng.src == nil {
			r.Rng.src = rand.NewPCG(0, 0)
		}
		r.Rng.rng = rand.New(r.Rng.src)
	}
	return r.Rng.rng
}

// SortedPlayerIDs retorna os ids dos jogadores em ordem lexicográfica, para que
// laços sobre o mapa Players tenham ordem determinística.
func (r *Room) SortedPlayerIDs() []string {
	ids := make([]string, 0, len(r.Players))
	for id := range r.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
