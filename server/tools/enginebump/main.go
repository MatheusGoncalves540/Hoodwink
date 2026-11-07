// enginebump sobe a versão do motor e congela o fecho atual em
// server/replay/engines/vN, reescrevendo os imports internos para o módulo do
// replay e embutindo o YAML de regras. Regenera versions.go e registry.go e
// sincroniza services.EngineVersion com game/ENGINE_VERSION.
//
// Uso: make engine-bump (a partir de server/).
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-tools/closure"
)

// registryTemplate é o bloco de uma versão da interface Engine; {{N}} é o número.
//
//go:embed registry.tmpl
var registryTemplate string

// errNothingToFreeze sinaliza que o fecho não mudou desde a última gravação.
var errNothingToFreeze = errors.New("fecho do motor não mudou desde a última versão")

func main() {
	repo := flag.String("repo", "", "raiz do repositório (vazio = auto-detecta)")
	registryOnly := flag.Bool("registry-only", false, "regenera só engines/registry.go a partir das cópias congeladas")
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

	if *registryOnly {
		if err := writeRegistry(replayDir); err != nil {
			fatal(err)
		}
		fmt.Println("engine-bump: registry.go regenerado")
		return
	}

	next, err := run(gameServerDir, replayDir)
	if err != nil {
		if errors.Is(err, errNothingToFreeze) {
			fmt.Println("engine-bump: nada a congelar (o fecho do motor não mudou)")
			return
		}
		fatal(err)
	}
	fmt.Printf("engine-bump: motor congelado como v%d\n", next)
}

func run(gameServerDir, replayDir string) (int, error) {
	current, err := readVersion(gameServerDir)
	if err != nil {
		return 0, err
	}

	// Primeira versão ainda não congelada → congela na versão atual; senão sobe.
	next := current
	if _, err := os.Stat(filepath.Join(replayDir, "engines", fmt.Sprintf("v%d", current))); err == nil {
		next = current + 1
	}

	sourceHash, err := closure.Hash(gameServerDir)
	if err != nil {
		return 0, err
	}

	// Idempotência: se o fecho não mudou desde a última gravação, não há o que
	// congelar (evita duplicar versões com o mesmo conteúdo).
	if stored, err := readTrimmed(filepath.Join(gameServerDir, "game", "engine.sum")); err == nil && stored == sourceHash {
		// Nada a congelar, mas o recorder pode ter ficado para trás num bump
		// interrompido; sincronizar aqui deixa o bump idempotente e auto-corretivo.
		if err := writeRecorderVersion(gameServerDir, current); err != nil {
			return 0, err
		}
		return current, errNothingToFreeze
	}

	pkgs, err := closure.Packages(gameServerDir)
	if err != nil {
		return 0, err
	}

	target := filepath.Join(replayDir, "engines", fmt.Sprintf("v%d", next))
	if err := os.RemoveAll(target); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return 0, err
	}

	// Copia os pacotes do fecho, reescrevendo o prefixo do módulo gameServer.
	fromPrefix := closure.GameServerModule + "/"
	toPrefix := closure.ReplayModule + "/engines/v" + strconv.Itoa(next) + "/"

	for _, pkg := range pkgs {
		srcDir := filepath.Join(gameServerDir, filepath.FromSlash(pkg.RelDir))
		dstDir := filepath.Join(target, filepath.FromSlash(pkg.RelDir))
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return 0, err
		}
		for _, name := range pkg.Files {
			data, err := os.ReadFile(filepath.Join(srcDir, name))
			if err != nil {
				return 0, err
			}
			rewritten := strings.ReplaceAll(string(data), fromPrefix, toPrefix)
			if err := os.WriteFile(filepath.Join(dstDir, name), []byte(rewritten), 0o644); err != nil {
				return 0, err
			}
		}
	}

	// Embuta o YAML de regras (o //go:embed de game/rules espera modes/*.yaml).
	if err := copyYaml(gameServerDir, target); err != nil {
		return 0, err
	}

	// Grava o hash da cópia congelada (exclui o próprio engine.sum).
	frozenHash, err := closure.FrozenHash(target)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filepath.Join(target, "engine.sum"), []byte(frozenHash+"\n"), 0o644); err != nil {
		return 0, err
	}

	// Regenera versions.go e registry.go a partir de todas as cópias congeladas.
	if err := writeVersions(replayDir); err != nil {
		return 0, err
	}
	if err := writeRegistry(replayDir); err != nil {
		return 0, err
	}

	// Atualiza a versão e o hash do fecho no gameServer.
	if err := os.WriteFile(filepath.Join(gameServerDir, "game", "ENGINE_VERSION"), []byte(strconv.Itoa(next)+"\n"), 0o644); err != nil {
		return 0, err
	}
	if err := os.WriteFile(filepath.Join(gameServerDir, "game", "engine.sum"), []byte(sourceHash+"\n"), 0o644); err != nil {
		return 0, err
	}
	if err := writeRecorderVersion(gameServerDir, next); err != nil {
		return 0, err
	}

	return next, nil
}

var recorderVersionRE = regexp.MustCompile(`(?m)^const EngineVersion = \d+\r?$`)

// writeRecorderVersion mantém o const do recorder igual à versão congelada.
func writeRecorderVersion(gameServerDir string, version int) error {
	path := filepath.Join(gameServerDir, "services", "replayRecorder.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !recorderVersionRE.Match(data) {
		return fmt.Errorf("não achei `const EngineVersion = N` em services/replayRecorder.go")
	}
	out := recorderVersionRE.ReplaceAll(data, []byte(fmt.Sprintf("const EngineVersion = %d", version)))
	return os.WriteFile(path, out, 0o644)
}

func readVersion(gameServerDir string) (int, error) {
	data, err := os.ReadFile(filepath.Join(gameServerDir, "game", "ENGINE_VERSION"))
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("ENGINE_VERSION inválido: %q", strings.TrimSpace(string(data)))
	}
	return n, nil
}

func readTrimmed(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func copyYaml(gameServerDir, target string) error {
	src := filepath.Join(gameServerDir, "game", "rules", "modes")
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	dst := filepath.Join(target, "game", "rules", "modes")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// frozenVersions devolve os números de versão congelados, ordenados.
func frozenVersions(replayDir string) ([]int, error) {
	dir := filepath.Join(replayDir, "engines")
	entries, err := os.ReadDir(dir)
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
	return versions, nil
}

func frozenHash(replayDir string, n int) (string, error) {
	data, err := os.ReadFile(filepath.Join(replayDir, "engines", fmt.Sprintf("v%d", n), "engine.sum"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeVersions(replayDir string) error {
	versions, err := frozenVersions(replayDir)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("// Code generated pelo enginebump (make engine-bump). NÃO edite à mão.\n")
	b.WriteString("//\n")
	b.WriteString("// Package versions lista as versões congeladas do motor sem importá-las\n")
	b.WriteString("// (por isso fica num pacote separado de engines).\n")
	b.WriteString("package versions\n\n")
	b.WriteString("// Version identifica uma cópia congelada do motor.\n")
	b.WriteString("type Version struct {\n")
	b.WriteString("\tNumber int\n")
	b.WriteString("\tHash   string // sha256 (hex) dos arquivos de replay/engines/vN\n")
	b.WriteString("}\n\n")
	b.WriteString("// All lista todas as versões congeladas, em ordem crescente.\n")
	b.WriteString("var All = []Version{\n")
	for _, n := range versions {
		h, err := frozenHash(replayDir, n)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "\t{Number: %d, Hash: %q},\n", n, h)
	}
	b.WriteString("}\n\n")
	b.WriteString("// Latest devolve o número da versão mais recente.\n")
	b.WriteString("func Latest() int {\n")
	b.WriteString("\tif len(All) == 0 {\n\t\treturn 0\n\t}\n")
	b.WriteString("\treturn All[len(All)-1].Number\n")
	b.WriteString("}\n\n")
	b.WriteString("// ByNumber localiza uma versão pelo número.\n")
	b.WriteString("func ByNumber(n int) (Version, bool) {\n")
	b.WriteString("\tfor _, v := range All {\n")
	b.WriteString("\t\tif v.Number == n {\n\t\t\treturn v, true\n\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn Version{}, false\n")
	b.WriteString("}\n")

	if err := os.MkdirAll(filepath.Join(replayDir, "versions"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(replayDir, "versions", "versions.go"), []byte(b.String()), 0o644)
}

func writeRegistry(replayDir string) error {
	versions, err := frozenVersions(replayDir)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("// Code generated pelo enginebump (make engine-bump). NÃO edite à mão.\n")
	b.WriteString("//\n")
	b.WriteString("// Package engines registra as cópias congeladas do motor por versão.\n")
	b.WriteString("// O runner consulta Registry e conversa com a sala como any: cada versão\n")
	b.WriteString("// faz o cast para os seus próprios tipos congelados.\n")
	b.WriteString("package engines\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"context\"\n\t\"encoding/json\"\n\t\"errors\"\n\t\"fmt\"\n\t\"sort\"\n\t\"strconv\"\n\t\"sync\"\n\t\"time\"\n\n")
	for _, n := range versions {
		v := strconv.Itoa(n)
		fmt.Fprintf(&b, "\tenginev%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/engine")
		fmt.Fprintf(&b, "\teffectsv%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/engine/effects")
		fmt.Fprintf(&b, "\trulesv%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/rules")
		fmt.Fprintf(&b, "\tstructsv%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/structs")
		fmt.Fprintf(&b, "\tplayersv%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/structs/players")
		fmt.Fprintf(&b, "\troomsv%s %q\n", v, closure.ReplayModule+"/engines/v"+v+"/game/structs/rooms")
	}
	fmt.Fprintf(&b, "\n\t%q\n", closure.ReplayModule+"/format")
	b.WriteString(")\n\n")

	b.WriteString(registryShared)

	for _, n := range versions {
		b.WriteString(strings.ReplaceAll(registryTemplate, "{{N}}", strconv.Itoa(n)))
		b.WriteString("\n")
	}

	b.WriteString("// Registry associa o número da versão ao motor congelado.\n")
	b.WriteString("var Registry = map[int]Engine{\n")
	for _, n := range versions {
		fmt.Fprintf(&b, "\t%d: EngineV%d{},\n", n, n)
	}
	b.WriteString("}\n")

	dir := filepath.Join(replayDir, "engines")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("formatar registry.go: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "registry.go"), src, 0o644)
}

// registryShared é escrito uma vez no registry.go: a interface que o runner vê
// e os helpers que não dependem da versão.
const registryShared = `// Engine isola o runner dos tipos congelados de cada versão. A sala circula
// como any; cada versão faz o cast para os seus próprios tipos.
type Engine interface {
	Version() int
	NewRoom(*format.Replay) (any, error)
	Clone(any) any
	SetNow(any, time.Time)
	ClearNow(any)
	SetConnected(any, []string)
	StartGame(raw, sink any) error
	ApplyPlay(raw, sink any, player string, payload []byte) (bool, error)
	Tick(raw, sink any) error
	CheckMode(string) error
	GameOver(any) bool
	WinnerID(any) string
	CanonicalJSON(any) ([]byte, error)
}

// matchIDString formata os 16 bytes do MatchID como UUID canônico.
func matchIDString(b [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// PlayerID mapeia o índice de jogador do replay para o id usado na sala.
func PlayerID(index uint8) string { return strconv.Itoa(int(index)) }

func truncateForHash(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.Truncate(time.Millisecond)
}

`

func mustCwd() string {
	wd, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	return wd
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "engine-bump: erro:", err)
	os.Exit(1)
}
