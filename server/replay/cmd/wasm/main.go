//go:build js && wasm

// Motor de replay do Hoodwink para o navegador.
//
// O binário é um por versão do motor (replay/engines/vN) e não tem rede nem
// Redis: recebe os bytes do .hwr, anda passo a passo com o runner e devolve o
// estado como JSON. O worker (T13) chama estas funções:
//
//	hoodwinkReplayLoad(bytes)  -> bool  (Uint8Array ou ArrayBuffer)
//	hoodwinkReplayStep()       -> bool  (false no fim ou em erro)
//	hoodwinkReplaySeek(n)      -> bool
//	hoodwinkReplayState()      -> string (JSON da sala; "" antes do início)
//	hoodwinkReplayStateJSON()  -> string (alias de State, compatibilidade)
//	hoodwinkReplayHash()       -> string (sha256 hex do estado canônico)
//	hoodwinkReplayKid()        -> string (kid da assinatura; "" sem assinatura)
//	hoodwinkReplayVerify(pub)  -> bool   (confere a assinatura com a pública de 32 bytes)
//	hoodwinkReplayError()      -> string (último erro; "" quando deu certo)
//	hoodwinkReplayInfo()       -> string (JSON {buildVersion, engineVersion, ...})
//
// Toda chamada limpa o erro anterior (menos Error/State/StateJSON/Hash/Info) e
// converte panic em mensagem de erro, para o wasm nunca derrubar o worker.
package main

import (
	"fmt"
	"syscall/js"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
)

func main() {
	s := &session{}

	js.Global().Set("hoodwinkReplayLoad", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return recoverBool(s, func() bool {
			if len(args) != 1 {
				return s.fail("hoodwinkReplayLoad espera 1 argumento (Uint8Array)")
			}
			return s.load(bytesFromJS(args[0]))
		})
	}))

	js.Global().Set("hoodwinkReplayStep", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverBool(s, s.step)
	}))

	js.Global().Set("hoodwinkReplaySeek", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return recoverBool(s, func() bool {
			if len(args) != 1 || args[0].Type() != js.TypeNumber {
				return s.fail("hoodwinkReplaySeek espera 1 argumento numérico")
			}
			return s.seek(args[0].Int())
		})
	}))

	state := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverString(s, s.state)
	})
	js.Global().Set("hoodwinkReplayState", state)
	js.Global().Set("hoodwinkReplayStateJSON", state)

	js.Global().Set("hoodwinkReplayHash", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverString(s, s.hash)
	}))

	js.Global().Set("hoodwinkReplayKid", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverString(s, s.kid)
	}))

	js.Global().Set("hoodwinkReplayVerify", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return recoverBool(s, func() bool {
			if len(args) != 1 {
				return s.fail("hoodwinkReplayVerify espera 1 argumento (Uint8Array com 32 bytes)")
			}
			if err := s.verify(bytesFromJS(args[0])); err != nil {
				return s.fail("%v", err)
			}
			return true
		})
	}))

	js.Global().Set("hoodwinkReplayError", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverString(s, s.err)
	}))

	js.Global().Set("hoodwinkReplayInfo", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		return recoverString(s, s.info)
	}))

	// Mantém o runtime Go vivo: o worker chama as funções exportadas até
	// encerrar o Worker (ou recarregar o replay do zero).
	select {}
}

// recoverBool executa fn convertendo panic em erro no session.
func recoverBool(s *session, fn func() bool) (ok bool) {
	defer func() {
		if rec := recover(); rec != nil {
			s.lastErr = fmt.Sprintf("panic no motor de replay: %v", rec)
			ok = false
		}
	}()
	return fn()
}

// recoverString executa fn convertendo panic em erro no session.
func recoverString(s *session, fn func() string) (out string) {
	defer func() {
		if rec := recover(); rec != nil {
			s.lastErr = fmt.Sprintf("panic no motor de replay: %v", rec)
			out = ""
		}
	}()
	return fn()
}

// bytesFromJS copia um Uint8Array (ou ArrayBuffer) para []byte. Acima do teto
// do formato devolve nil: o Decode recusa com "replay inválido" sem alocar o
// tamanho pedido pelo JS.
func bytesFromJS(v js.Value) []byte {
	if v.IsUndefined() || v.IsNull() || v.Type() != js.TypeObject {
		return nil
	}

	// ArrayBuffer não tem length; Uint8Array tem. Envolve o primeiro.
	if v.Get("length").Type() != js.TypeNumber {
		v = js.Global().Get("Uint8Array").New(v)
	}

	n := v.Get("length").Int()
	if n <= 0 || n > format.MaxFileSize {
		return nil
	}

	buf := make([]byte, n)
	js.CopyBytesToGo(buf, v)
	return buf
}
