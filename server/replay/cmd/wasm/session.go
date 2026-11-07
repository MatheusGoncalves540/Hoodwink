// Package main é o entrypoint do motor de replay para o navegador.
//
// A lógica fica em session.go (sem build tag), para poder ser testada no host;
// main.go é só a ponte com syscall/js e main_stub.go mantém `go build ./...` e
// `go vet ./...` funcionando fora de js/wasm.
package main

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/MatheusGoncalves540/Hoodwink-replay/runner"
)

// buildEngineVersion é preenchido no build por
// `-ldflags "-X main.buildEngineVersion=N"`, um binário por versão do motor.
// Vazio em builds de desenvolvimento (go test / go build sem ldflags).
var buildEngineVersion string

// session embrulha o runner para a API exportada ao JavaScript.
type session struct {
	runner *runner.Runner
	replay *format.Replay
	// data guarda o arquivo original: a verificação de assinatura (Verify) precisa
	// do envelope gzipado exatamente como chegou, não do estado decodificado.
	data    []byte
	lastErr string
}

// clearRunner fecha e descarta o runner atual.
func (s *session) clearRunner() {
	if s.runner != nil {
		_ = s.runner.Close()
	}
	s.runner = nil
	s.replay = nil
	s.data = nil
}

// fail guarda a mensagem de erro e devolve false.
func (s *session) fail(msg string, args ...any) bool {
	s.lastErr = fmt.Sprintf(msg, args...)
	return false
}

// load decodifica o .hwr, confere a versão do motor compilado e prepara o
// runner. A assinatura não é conferida aqui: quem valida o selo "oficial" é o
// bundle (T13), com as chaves públicas embutidas e a função verify abaixo.
func (s *session) load(data []byte) bool {
	s.lastErr = ""
	s.clearRunner()

	replay, err := format.Decode(data, nil)
	if err != nil {
		return s.fail("%v", err)
	}

	// Cada .wasm é compilado para uma versão; recusar outra versão evita rodar
	// o replay com o motor errado caso o manifesto esteja desatualizado.
	if want, err := strconv.Atoi(buildEngineVersion); err == nil && want > 0 && replay.EngineVersion != want {
		return s.fail("replay da versão %d; este motor compila a v%s", replay.EngineVersion, buildEngineVersion)
	}

	r, err := runner.New(replay, runner.Options{})
	if err != nil {
		return s.fail("%v", err)
	}

	s.runner = r
	s.replay = replay
	s.data = data
	return true
}

// step aplica a próxima entrada. No fim do replay devolve false sem erro.
func (s *session) step() bool {
	s.lastErr = ""
	if s.runner == nil {
		return s.fail("replay não carregado")
	}
	if err := s.runner.Step(); err != nil {
		if errors.Is(err, runner.ErrDone) {
			return false
		}
		return s.fail("%v", err)
	}
	return true
}

// seek posiciona o replay no passo n (0..total), reaproveitando pontos de
// retorno. Não confere checkpoints: é navegação de barra de tempo.
func (s *session) seek(n int) bool {
	s.lastErr = ""
	if s.runner == nil {
		return s.fail("replay não carregado")
	}
	if err := s.runner.Seek(n); err != nil {
		return s.fail("%v", err)
	}
	return true
}

// kid devolve o kid da chave que assinou o replay carregado ("" quando o
// arquivo não tem assinatura). O bundle resolve esse kid para a chave pública.
func (s *session) kid() string {
	if s.replay == nil {
		return ""
	}
	return s.replay.Kid
}

// verify confere a assinatura do arquivo carregado com a chave pública dada
// (32 bytes Ed25519 crus). Devolve nil quando a assinatura confere e um erro
// tipado nos demais casos; o T13 usa isso para o selo "oficial".
func (s *session) verify(pub []byte) error {
	if s.replay == nil || len(s.data) == 0 {
		return errors.New("replay não carregado")
	}
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("chave pública inválida: %d bytes", len(pub))
	}
	if s.replay.Kid == "" {
		return format.ErrInvalid
	}

	keys := format.StaticKeys(map[string]ed25519.PublicKey{s.replay.Kid: ed25519.PublicKey(pub)})
	return format.Verify(s.data, keys)
}

// state devolve o JSON do estado atual da sala ("" antes do início).
func (s *session) state() string {
	if s.runner == nil {
		return ""
	}
	room := s.runner.State()
	if room == nil {
		return ""
	}
	b, err := json.Marshal(room)
	if err != nil {
		s.lastErr = fmt.Sprintf("serializar estado: %v", err)
		return ""
	}
	return string(b)
}

// hash devolve o sha256 canônico do estado atual, em hex.
func (s *session) hash() string {
	if s.runner == nil {
		return ""
	}
	return s.runner.Hash()
}

// err devolve o último erro (vazio quando a última chamada deu certo).
func (s *session) err() string { return s.lastErr }

// info descreve o replay carregado e a posição atual, em JSON.
type sessionInfo struct {
	BuildVersion  string `json:"buildVersion"`
	EngineVersion int    `json:"engineVersion"`
	FormatVersion int    `json:"formatVersion"`
	Mode          string `json:"mode"`
	Players       int    `json:"players"`
	Steps         int    `json:"steps"`
	Position      int    `json:"position"`
	Done          bool   `json:"done"`
}

func (s *session) info() string {
	info := sessionInfo{BuildVersion: buildEngineVersion}
	if s.replay != nil {
		info.EngineVersion = s.replay.EngineVersion
		info.FormatVersion = s.replay.FormatVersion
		info.Mode = s.replay.Mode
		info.Players = len(s.replay.Players)
		info.Steps = len(s.replay.Entries)
	}
	if s.runner != nil {
		info.Position = s.runner.Position()
		info.Done = info.Position >= info.Steps
	}
	b, err := json.Marshal(info)
	if err != nil {
		s.lastErr = fmt.Sprintf("serializar info: %v", err)
		return "{}"
	}
	return string(b)
}
