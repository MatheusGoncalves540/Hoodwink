// enginecheck verifica que o motor puro do gameServer está congelado na versão
// certa. Roda no CI (make engine-check) e falha com mensagem clara quando:
//
//	(a) o hash do fecho mudou sem subir a versão (engine.sum desatualizado);
//	(b) existe time.Now/rand. no motor fora de determinism.go (e game/runtime);
//	(c) o fecho contém go-redis/gorilla/redisFuncs;
//	(d) uma cópia congelada replay/engines/vN foi editada;
//	(e) services.EngineVersion divergiu de game/ENGINE_VERSION.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-tools/closure"
)

func main() {
	repo := flag.String("repo", "", "raiz do repositório (vazio = auto-detecta)")
	flag.Parse()

	root, err := closure.FindRepoRoot(mustCwd())
	if err != nil {
		fatal(err)
	}
	if *repo != "" {
		root = *repo
	}

	gameServerDir := filepath.Join(root, "server", "gameServer")
	replayDir := filepath.Join(root, "server", "replay")

	if err := run(gameServerDir, replayDir); err != nil {
		fatal(err)
	}
	fmt.Println("engine-check: ok")
}

func run(gameServerDir, replayDir string) error {
	// Versão: inteiro >= 1.
	version, err := readEngineVersion(gameServerDir)
	if err != nil {
		return err
	}

	// (e) o recorder precisa gravar a versão que o fecho diz.
	recorderVersion, err := readRecorderVersion(gameServerDir)
	if err != nil {
		return err
	}
	if recorderVersion != version {
		return fmt.Errorf("services.EngineVersion = %d, mas game/ENGINE_VERSION = %d; rode make engine-bump", recorderVersion, version)
	}

	// (c) fecho não pode conter go-redis/gorilla/redisFuncs.
	deps, err := closure.ForbiddenDeps(gameServerDir)
	if err != nil {
		return err
	}
	if len(deps) > 0 {
		return fmt.Errorf("fecho do motor contém dependência proibida:\n  %s", strings.Join(deps, "\n  "))
	}

	// (b) time.Now/rand. fora de determinism.go (game/runtime fica fora do fecho).
	uses, err := closure.ForbiddenUses(gameServerDir)
	if err != nil {
		return err
	}
	if len(uses) > 0 {
		return fmt.Errorf("time.Now/rand. no motor fora de determinism.go:\n  %s", strings.Join(uses, "\n  "))
	}

	// (a) hash do fecho deve bater com engine.sum.
	got, err := closure.Hash(gameServerDir)
	if err != nil {
		return err
	}
	want, err := readTrimmed(filepath.Join(gameServerDir, "game", "engine.sum"))
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf(
			"o hash do motor mudou sem subir a versão (engine.sum desatualizado); rode make engine-bump\n  atual:  %s\n  gravado: %s",
			got, want)
	}

	// A cópia congelada da versão atual precisa existir (bump sempre congela).
	if _, err := os.Stat(filepath.Join(replayDir, "engines", fmt.Sprintf("v%d", version))); err != nil {
		return fmt.Errorf("faltou congelar a v%d em server/replay/engines/v%d; rode make engine-bump", version, version)
	}

	// (d) cópias congeladas não podem ter sido editadas.
	violations, err := checkFrozen(replayDir)
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		return fmt.Errorf("cópia congelada editada:\n  %s", strings.Join(violations, "\n  "))
	}

	return nil
}

var recorderVersionRE = regexp.MustCompile(`(?m)^const EngineVersion = (\d+)\r?$`)

// readRecorderVersion lê o const que o recorder usa no cabeçalho do .hwr.
func readRecorderVersion(gameServerDir string) (int, error) {
	data, err := os.ReadFile(filepath.Join(gameServerDir, "services", "replayRecorder.go"))
	if err != nil {
		return 0, err
	}
	m := recorderVersionRE.FindSubmatch(data)
	if m == nil {
		return 0, fmt.Errorf("não achei `const EngineVersion = N` em services/replayRecorder.go")
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("EngineVersion inválido: %q (esperado inteiro >= 1)", m[1])
	}
	return n, nil
}

func readEngineVersion(gameServerDir string) (int, error) {
	s, err := readTrimmed(filepath.Join(gameServerDir, "game", "ENGINE_VERSION"))
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("ENGINE_VERSION inválido: %q (esperado inteiro >= 1)", s)
	}
	return n, nil
}

// checkFrozen confere o hash de cada replay/engines/vN contra o vN/engine.sum.
func checkFrozen(replayDir string) ([]string, error) {
	enginesDir := filepath.Join(replayDir, "engines")
	entries, err := os.ReadDir(enginesDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var versions []int
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "v") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "v"))
		if err != nil {
			continue
		}
		versions = append(versions, n)
	}
	sort.Ints(versions)

	var bad []string
	for _, n := range versions {
		dir := filepath.Join(enginesDir, fmt.Sprintf("v%d", n))
		want, err := readTrimmed(filepath.Join(dir, "engine.sum"))
		if err != nil {
			return nil, fmt.Errorf("v%d sem engine.sum: %w", n, err)
		}
		got, err := closure.FrozenHash(dir)
		if err != nil {
			return nil, err
		}
		if got != want {
			bad = append(bad, fmt.Sprintf("v%d: hash gravado %s != atual %s", n, want, got))
		}
	}
	return bad, nil
}

func readTrimmed(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func mustCwd() string {
	wd, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	return wd
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "engine-check: erro:", err)
	os.Exit(1)
}
