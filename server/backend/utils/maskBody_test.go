package utils

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMaskBody(t *testing.T) {
	invalidJSON := `{"token":`
	lixoDepoisDoValor := `{"token":"x"} lixo`
	semContentType := `{"token":"x"}`
	textoPlano := `{"token":"x"}`

	tests := []struct {
		name        string
		contentType string
		raw         string
		want        string
	}{
		{
			name:        "sem corpo",
			contentType: "application/json",
			raw:         "",
			want:        "",
		},
		{
			name:        "content type nao json",
			contentType: "text/plain",
			raw:         textoPlano,
			want:        fmt.Sprintf("<%d bytes>", len(textoPlano)),
		},
		{
			name:        "sem content type",
			contentType: "",
			raw:         semContentType,
			want:        fmt.Sprintf("<%d bytes>", len(semContentType)),
		},
		{
			name:        "json invalido",
			contentType: "application/json",
			raw:         invalidJSON,
			want:        fmt.Sprintf("<%d bytes>", len(invalidJSON)),
		},
		{
			name:        "json com lixo depois do valor",
			contentType: "application/json",
			raw:         lixoDepoisDoValor,
			want:        fmt.Sprintf("<%d bytes>", len(lixoDepoisDoValor)),
		},
		{
			name:        "chaves aninhadas",
			contentType: "application/json",
			raw:         `{"user":{"name":"Ana","token":"abc","nested":{"password":"p","authorization":"Bearer x"}},"ok":true}`,
			want:        `{"ok":true,"user":{"name":"Ana","nested":{"authorization":"***","password":"***"},"token":"***"}}`,
		},
		{
			name:        "array com objetos",
			contentType: "application/json",
			raw:         `[{"ticket":"abc"},{"name":"x","jwt":"y"}]`,
			want:        `[{"ticket":"***"},{"jwt":"***","name":"x"}]`,
		},
		{
			name:        "Token e accessToken",
			contentType: "application/json",
			raw:         `{"Token":"x","accessToken":"y"}`,
			want:        `{"Token":"***","accessToken":"***"}`,
		},
		{
			name:        "demais chaves sensiveis",
			contentType: "application/json",
			raw:         `{"senha":"1","secret":"2","apiKey":"3","api-key":"4","credential":"5","jwt":"6","ticket":"7","password":"8","authorization":"9"}`,
			want:        `{"api-key":"***","apiKey":"***","authorization":"***","credential":"***","jwt":"***","password":"***","secret":"***","senha":"***","ticket":"***"}`,
		},
		{
			name:        "valor sensivel nao string",
			contentType: "application/json",
			raw:         `{"token":{"a":1},"senha":["x"]}`,
			want:        `{"senha":"***","token":"***"}`,
		},
		{
			name:        "arrays numeros null e booleano",
			contentType: "application/json",
			raw:         `[null,1,2.5,true,"texto",{"jwt":"abc"}]`,
			want:        `[null,1,2.5,true,"texto",{"jwt":"***"}]`,
		},
		{
			name:        "numero grande preservado",
			contentType: "application/json",
			raw:         `{"n":12345678901234567890}`,
			want:        `{"n":12345678901234567890}`,
		},
		{
			name:        "content type com charset",
			contentType: "application/json; charset=utf-8",
			raw:         `{"token":"x"}`,
			want:        `{"token":"***"}`,
		},
		{
			name:        "content type +json",
			contentType: "application/problem+json",
			raw:         `{"token":"x"}`,
			want:        `{"token":"***"}`,
		},
		{
			name:        "unicode",
			contentType: "application/json",
			raw:         `{"nome":"José 🎮","token":"x"}`,
			want:        `{"nome":"José 🎮","token":"***"}`,
		},
		{
			name:        "espacos em volta",
			contentType: "application/json",
			raw:         "  {\"token\":\"x\"}  ",
			want:        `{"token":"***"}`,
		},
		{
			name:        "json escalar",
			contentType: "application/json",
			raw:         `"texto sem chaves"`,
			want:        `"texto sem chaves"`,
		},
		{
			name:        "somente espacos",
			contentType: "application/json",
			raw:         "   ",
			want:        fmt.Sprintf("<%d bytes>", len("   ")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskBody(tt.contentType, tt.raw); got != tt.want {
				t.Fatalf("MaskBody(%q, %q) = %q, quer %q", tt.contentType, tt.raw, got, tt.want)
			}
		})
	}
}

func TestMaskBodyCortaEm4096(t *testing.T) {
	t.Run("chave sensivel depois dos 4 KB nao vaza", func(t *testing.T) {
		raw := `["` + strings.Repeat("a", 5000) + `",{"token":"segredo-nao-pode-vazar"}]`

		got := MaskBody("application/json", raw)

		if len(got) > maxMaskedBody {
			t.Fatalf("corpo com %d bytes, quer no maximo %d", len(got), maxMaskedBody)
		}
		if strings.Contains(got, "segredo-nao-pode-vazar") {
			t.Fatalf("segredo vazou no corpo truncado: %q", got)
		}
		if !utf8.ValidString(got) {
			t.Fatal("corpo truncado nao é UTF-8 válido")
		}
	})

	t.Run("unicode truncado sem quebrar rune", func(t *testing.T) {
		raw := `{"texto":"` + strings.Repeat("é", 3000) + `"}`

		got := MaskBody("application/json", raw)

		if len(got) > maxMaskedBody {
			t.Fatalf("corpo com %d bytes, quer no maximo %d", len(got), maxMaskedBody)
		}
		if len(got) < maxMaskedBody-4 {
			t.Fatalf("corpo curto demais (%d bytes): o corte não aconteceu", len(got))
		}
		if !utf8.ValidString(got) {
			t.Fatal("corte quebrou uma sequência UTF-8")
		}
	})

	t.Run("json gigante nao entra em panico", func(t *testing.T) {
		raw := `{"data":"` + strings.Repeat("x", 1<<20) + `","token":"abc"}`

		got := MaskBody("application/json", raw)

		if len(got) > maxMaskedBody {
			t.Fatalf("corpo com %d bytes, quer no maximo %d", len(got), maxMaskedBody)
		}
		if !utf8.ValidString(got) {
			t.Fatal("corpo truncado nao é UTF-8 válido")
		}
	})

	t.Run("json invalido gigante vira contagem", func(t *testing.T) {
		raw := strings.Repeat("x", 1<<20)
		want := fmt.Sprintf("<%d bytes>", len(raw))

		if got := MaskBody("application/json", raw); got != want {
			t.Fatalf("MaskBody(json invalido gigante) = %q, quer %q", got, want)
		}
	})
}
