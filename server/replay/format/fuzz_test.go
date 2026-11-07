package format

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

// FuzzDecode garante que o leitor nunca entra em panic para qualquer entrada.
func FuzzDecode(f *testing.F) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		f.Fatal(err)
	}
	signer := NewEd25519Signer("kid-fuzz", priv)
	lookup := StaticKeys(map[string]ed25519.PublicKey{"kid-fuzz": pub})

	// Sementes: replay válido sem assinatura e com assinatura.
	r := validReplay(2)
	fillCheckpoints(r)
	unsigned, err := Encode(r, nil)
	if err != nil {
		f.Fatal(err)
	}
	signed, err := Encode(r, signer)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(unsigned)
	f.Add(signed)
	f.Add([]byte{})
	f.Add([]byte("HWR1"))
	f.Add([]byte{0x00, 0x00, 0x00, 0x00})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Não pode panic em nenhuma das duas variantes (com e sem chaves).
		_, _ = Decode(data, nil)
		_, _ = Decode(data, lookup)
		_ = Verify(data, lookup)
	})
}
