package utils

import (
	"net"
	"net/http"
	"strings"
)

const maxUserAgentLength = 256

// ClientIP devolve o IP de quem fez a request.
//
// Só confia no cabeçalho X-Real-IP quando a conexão vem de loopback (o nginx da
// borda roda na mesma máquina e repassa o IP real por ali) E o valor do cabeçalho
// é um IP válido. Em qualquer outro caso usa o RemoteAddr sem porta. Nunca lê
// X-Forwarded-For nem CF-Connecting-IP (um CF-Connecting-IP forjado é ignorado
// pelo módulo realip quando a conexão não vem de faixa da Cloudflare).
func ClientIP(r *http.Request) string {
	host := hostOnly(r.RemoteAddr)
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		if x := strings.TrimSpace(r.Header.Get("X-Real-IP")); x != "" && net.ParseIP(x) != nil {
			return x
		}
	}
	return host
}

// UserAgent devolve o User-Agent cortado em 256 caracteres.
func UserAgent(r *http.Request) string {
	ua := r.UserAgent()
	if len(ua) > maxUserAgentLength {
		return ua[:maxUserAgentLength]
	}
	return ua
}

// hostOnly remove a porta do RemoteAddr no formato host:port. IPv6 com colchetes
// e IPv4 passam por net.SplitHostPort; IPv6 sem colchetes e sem porta volta como está.
func hostOnly(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return strings.Trim(addr, "[]")
}
