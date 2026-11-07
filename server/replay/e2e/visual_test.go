package e2e

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// E2E_VISUAL=1 (ou true) imprime o andamento no stderr. Sem a env, os helpers
// devolvem nil e o teste não paga nada além de um branch por passo — é o modo
// usado no CI. A saída é escrita direto no stderr para aparecer ao vivo, sem
// esperar o go test drenar o t.Log do subtest.
var e2eVisualEnabled = func() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("E2E_VISUAL")))
	return v == "1" || v == "true"
}()

type visualReporter struct {
	label   string
	start   time.Time
	last    time.Time
	lastLen int
}

func newVisualReporter(label string, players int) *visualReporter {
	if !e2eVisualEnabled {
		return nil
	}
	v := &visualReporter{label: label, start: time.Now(), last: time.Now()}
	fmt.Fprintf(os.Stderr, "\n▶ %s | %d jogadores\n", label, players)
	return v
}

// step é chamado em todo passo, mas só escreve a cada ~150 ms ou 25 passos.
// O branch e o time.Since são baratos; a formatação só acontece quando imprime.
func (v *visualReporter) step(n int, action string, room *rooms.Room) {
	if v == nil {
		return
	}
	now := time.Now()
	if n%25 != 0 && now.Sub(v.last) < 150*time.Millisecond {
		return
	}
	v.last = now

	line := fmt.Sprintf("[%s] passo %d | turno %d | vivos %d | %s",
		v.label, n, room.Turn, aliveCount(room), action)
	if v.lastLen > len(line) {
		line += strings.Repeat(" ", v.lastLen-len(line))
	}
	v.lastLen = len(line)
	fmt.Fprintf(os.Stderr, "\r%s", line)
}

func (v *visualReporter) end(entries, size int, hash string) {
	if v == nil {
		return
	}
	short := hash
	if len(short) > 12 {
		short = short[:12]
	}
	fmt.Fprintf(os.Stderr, "\r%s\r✔ %s | %d entradas | %d B | hash %s\n",
		strings.Repeat(" ", v.lastLen), v.label, entries, size, short)
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// shortAction extrai o type do JSON da jogada para a linha de progresso.
func shortAction(raw string) string {
	const marker = `"type":"`
	if i := strings.Index(raw, marker); i >= 0 {
		rest := raw[i+len(marker):]
		if j := strings.Index(rest, `"`); j >= 0 {
			return rest[:j]
		}
	}
	if len(raw) > 40 {
		return raw[:40]
	}
	return raw
}

func visualSummary(games, min, median, max int) {
	if !e2eVisualEnabled {
		return
	}
	fmt.Fprintf(os.Stderr, "\n✔ %d partidas | tamanho min=%d B, mediano=%d B, máximo=%d B\n",
		games, min, median, max)
}
