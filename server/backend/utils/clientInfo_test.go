package utils

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "loopback com X-Real-IP válido usa o cabeçalho",
			remoteAddr: "127.0.0.1:5000",
			headers:    map[string]string{"X-Real-IP": "1.2.3.4"},
			want:       "1.2.3.4",
		},
		{
			name:       "loopback com X-Real-IP lixo cai no RemoteAddr",
			remoteAddr: "127.0.0.1:5000",
			headers:    map[string]string{"X-Real-IP": "lixo"},
			want:       "127.0.0.1",
		},
		{
			name:       "Radmin (26.x) ignora X-Real-IP",
			remoteAddr: "26.0.0.5:5000",
			headers:    map[string]string{"X-Real-IP": "1.2.3.4"},
			want:       "26.0.0.5",
		},
		{
			name:       "IPv4-mapped IPv6 loopback conta como loopback",
			remoteAddr: "[::ffff:127.0.0.1]:5000",
			headers:    map[string]string{"X-Real-IP": "9.9.9.9"},
			want:       "9.9.9.9",
		},
		{
			name:       "IPv6 sem colchetes e sem porta",
			remoteAddr: "::1",
			want:       "::1",
		},
		{
			name:       "nunca lê X-Forwarded-For",
			remoteAddr: "127.0.0.1:5000",
			headers:    map[string]string{"X-Forwarded-For": "6.6.6.6"},
			want:       "127.0.0.1",
		},
		{
			name:       "nunca lê CF-Connecting-IP",
			remoteAddr: "127.0.0.1:5000",
			headers:    map[string]string{"CF-Connecting-IP": "6.6.6.6"},
			want:       "127.0.0.1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			if got := ClientIP(req); got != tc.want {
				t.Fatalf("ClientIP = %q, esperado %q", got, tc.want)
			}
		})
	}
}

func TestUserAgentTruncado(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", strings.Repeat("a", 300))

	got := UserAgent(req)
	if len(got) != 256 {
		t.Fatalf("UserAgent com %d caracteres, esperado 256", len(got))
	}

	short := httptest.NewRequest("GET", "/", nil)
	short.Header.Set("User-Agent", "hoodwink/1.0")
	if got := UserAgent(short); got != "hoodwink/1.0" {
		t.Fatalf("UserAgent = %q, esperado hoodwink/1.0", got)
	}
}
