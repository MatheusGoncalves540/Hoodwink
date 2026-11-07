package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-tools/closure"
)

func TestReadEngineVersion(t *testing.T) {
	dir := t.TempDir()
	game := filepath.Join(dir, "game")
	if err := os.MkdirAll(game, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(game, "ENGINE_VERSION"), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("1\n")
	if n, err := readEngineVersion(dir); err != nil || n != 1 {
		t.Fatalf("1 => (%d, %v)", n, err)
	}

	write("0")
	if _, err := readEngineVersion(dir); err == nil {
		t.Fatal("0 deveria falhar (versão >= 1)")
	}

	write("abc")
	if _, err := readEngineVersion(dir); err == nil {
		t.Fatal("abc deveria falhar")
	}
}

func TestReadRecorderVersion(t *testing.T) {
	dir := t.TempDir()
	servicesDir := filepath.Join(dir, "services")
	if err := os.MkdirAll(servicesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(servicesDir, "replayRecorder.go")
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("package services\n\nconst EngineVersion = 2\n")
	if n, err := readRecorderVersion(dir); err != nil || n != 2 {
		t.Fatalf("2 => (%d, %v)", n, err)
	}

	// Checkout no Windows aplica CRLF; o check precisa funcionar assim também.
	write("package services\r\n\r\nconst EngineVersion = 3\r\n")
	if n, err := readRecorderVersion(dir); err != nil || n != 3 {
		t.Fatalf("3 em CRLF => (%d, %v)", n, err)
	}

	write("package services\n\nconst engineVersion = 2\n")
	if _, err := readRecorderVersion(dir); err == nil {
		t.Fatal("const com nome errado deveria falhar")
	}

	write("package services\n")
	if _, err := readRecorderVersion(dir); err == nil {
		t.Fatal("arquivo sem o const deveria falhar")
	}
}

func TestCheckFrozenDetectaEdicao(t *testing.T) {
	replay := t.TempDir()
	v1 := filepath.Join(replay, "engines", "v1")
	if err := os.MkdirAll(v1, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeFile(filepath.Join(v1, "game", "engine", "play.go"), "package engine\nfunc Play() {}\n")

	hash, err := closure.FrozenHash(v1)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(filepath.Join(v1, "engine.sum"), hash+"\n")

	// Sem edição: nenhuma violação.
	if bad, err := checkFrozen(replay); err != nil || len(bad) != 0 {
		t.Fatalf("sem edição => (%v, %v)", bad, err)
	}

	// testdata (golden de replay) não faz parte do fecho do motor.
	writeFile(filepath.Join(v1, "testdata", "golden_1.hwr"), "golden")
	hashComTestdata, err := closure.FrozenHash(v1)
	if err != nil {
		t.Fatal(err)
	}
	if hashComTestdata != hash {
		t.Fatal("testdata não deveria mudar o hash da cópia congelada")
	}

	// Edita 1 caractere: deve acusar.
	writeFile(filepath.Join(v1, "game", "engine", "play.go"), "package engine\nfunc Play2() {}\n")
	bad, err := checkFrozen(replay)
	if err != nil {
		t.Fatal(err)
	}
	if len(bad) != 1 || !strings.Contains(bad[0], "v1") {
		t.Fatalf("edição deveria acusar v1, veio %v", bad)
	}
}

func TestFrozenHashIndependeDeCRLF(t *testing.T) {
	dir := t.TempDir()
	writeFile := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeFile("a.go", "package a\r\nfunc A() {}\r\n")
	writeFile("b.go", "package a\nfunc B() {}\n")
	h1, err := closure.FrozenHash(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Mesmo conteúdo com LF em tudo e newline final diferente.
	writeFile("a.go", "package a\nfunc A() {}\n")
	writeFile("b.go", "package a\nfunc B() {}")
	h2, err := closure.FrozenHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("hash deveria ser independente de CRLF/newline final:\n  %s\n  %s", h1, h2)
	}

	// Um caractere a mais muda o hash.
	writeFile("b.go", "package a\nfunc B() {}\nX")
	h3, err := closure.FrozenHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	if h3 == h2 {
		t.Fatal("conteúdo diferente deveria mudar o hash")
	}
}
