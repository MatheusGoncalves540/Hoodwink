// Package closure calcula o fecho de pacotes do motor puro do gameServer e o
// hash canônico (independente de path/CRLF) usado por enginecheck e enginebump.
//
// O fecho é descoberto com `go list -deps ./game/engine` no módulo do gameServer,
// filtrando os pacotes internos (prefixo do módulo) e os módulos de terceiros
// (para ler as linhas correspondentes do go.sum).
package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// GameServerModule é o módulo do gameServer (prefixo dos pacotes internos).
	GameServerModule = "github.com/MatheusGoncalves540/Hoodwink-gameServer"
	// ReplayModule é o módulo do replay (prefixo dos pacotes congelados).
	ReplayModule = "github.com/MatheusGoncalves540/Hoodwink-replay"
	// EnginePackage é o pacote raiz do motor, usado como âncora do fecho.
	EnginePackage = "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
)

// Module identifica um módulo de terceiros do fecho (path + versão).
type Module struct {
	Path    string
	Version string
}

// PackageInfo descreve um pacote interno do fecho.
type PackageInfo struct {
	// ImportPath é o caminho de importação completo (ex.: .../game/engine).
	ImportPath string
	// RelDir é o diretório relativo ao gameServer (ex.: game/engine).
	RelDir string
	// Files são os arquivos .go de produção (sem _test.go), ordenados.
	Files []string
}

// Normalize converte o conteúdo para LF e remove newlines finais, tornando o
// hash independente de CRLF e de newline final.
func Normalize(data []byte) string {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimRight(s, "\n")
}

// listDeps executa `go list -deps` no diretório do gameServer e devolve as
// linhas `importPath|modulePath|version` dos pacotes não-stdlib do fecho.
func listDeps(gameServerDir string) ([]string, error) {
	cmd := exec.Command("go", "list", "-deps", "-f",
		`{{if not .Standard}}{{.ImportPath}}|{{.Module.Path}}|{{.Module.Version}}{{end}}`,
		"./game/engine")
	cmd.Dir = gameServerDir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go list -deps falhou: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("go list -deps falhou: %w", err)
	}

	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// Packages retorna os pacotes internos do fecho do motor, ordenados por import path.
func Packages(gameServerDir string) ([]PackageInfo, error) {
	lines, err := listDeps(gameServerDir)
	if err != nil {
		return nil, err
	}

	byImport := map[string]PackageInfo{}
	for _, line := range lines {
		parts := strings.Split(line, "|")
		importPath := parts[0]
		if !strings.HasPrefix(importPath, GameServerModule+"/") {
			continue
		}
		relDir := strings.TrimPrefix(importPath, GameServerModule+"/")
		files, err := productionGoFiles(gameServerDir, relDir)
		if err != nil {
			return nil, err
		}
		byImport[importPath] = PackageInfo{ImportPath: importPath, RelDir: relDir, Files: files}
	}

	out := make([]PackageInfo, 0, len(byImport))
	for _, p := range byImport {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ImportPath < out[j].ImportPath })
	return out, nil
}

// ThirdParty retorna os módulos de terceiros do fecho, ordenados por path.
func ThirdParty(gameServerDir string) ([]Module, error) {
	lines, err := listDeps(gameServerDir)
	if err != nil {
		return nil, err
	}

	byKey := map[string]Module{}
	for _, line := range lines {
		parts := strings.Split(line, "|")
		importPath := parts[0]
		if strings.HasPrefix(importPath, GameServerModule+"/") {
			continue
		}
		modulePath, version := parts[1], parts[2]
		if modulePath == "" {
			continue
		}
		byKey[modulePath+"@"+version] = Module{Path: modulePath, Version: version}
	}

	out := make([]Module, 0, len(byKey))
	for _, m := range byKey {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Version < out[j].Version
	})
	return out, nil
}

// productionGoFiles lista os arquivos .go de produção (sem _test.go) de um
// diretório do gameServer, ordenados por nome.
func productionGoFiles(gameServerDir, relDir string) ([]string, error) {
	dir := filepath.Join(gameServerDir, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ler %s: %w", relDir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	return files, nil
}

// yamlDir é o diretório (relativo ao gameServer) dos modos de regras embutidos.
const yamlDir = "game/rules/modes"

// Hash calcula o sha256 canônico do fecho do motor:
//   - conteúdo normalizado dos arquivos .go dos pacotes internos (ordenados);
//   - conteúdo normalizado dos YAML de regras (game/rules/modes/*.yaml);
//   - linhas do go.sum dos módulos de terceiros que o fecho usa.
func Hash(gameServerDir string) (string, error) {
	pkgs, err := Packages(gameServerDir)
	if err != nil {
		return "", err
	}
	thirdParty, err := ThirdParty(gameServerDir)
	if err != nil {
		return "", err
	}

	h := sha256.New()

	for _, pkg := range pkgs {
		fmt.Fprintf(h, "PKG %s\n", pkg.ImportPath)
		dir := filepath.Join(gameServerDir, filepath.FromSlash(pkg.RelDir))
		for _, name := range pkg.Files {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return "", err
			}
			fmt.Fprintf(h, "FILE %s\n%s\n", name, Normalize(data))
		}
	}

	yamlFiles, err := yamlModeFiles(gameServerDir)
	if err != nil {
		return "", err
	}
	yamlAbs := filepath.Join(gameServerDir, filepath.FromSlash(yamlDir))
	for _, name := range yamlFiles {
		data, err := os.ReadFile(filepath.Join(yamlAbs, name))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "YAML %s\n%s\n", name, Normalize(data))
	}

	goSumLines, err := goSumForModules(gameServerDir, thirdParty)
	if err != nil {
		return "", err
	}
	for _, line := range goSumLines {
		fmt.Fprintf(h, "GOSUM %s\n", Normalize([]byte(line)))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// yamlModeFiles lista os arquivos .yaml de game/rules/modes, ordenados.
func yamlModeFiles(gameServerDir string) ([]string, error) {
	dir := filepath.Join(gameServerDir, filepath.FromSlash(yamlDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ler %s: %w", yamlDir, err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	return files, nil
}

// goSumForModules devolve as linhas do go.sum (ordenadas) das versões dos
// módulos de terceiros do fecho.
func goSumForModules(gameServerDir string, mods []Module) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(gameServerDir, "go.sum"))
	if err != nil {
		return nil, fmt.Errorf("ler go.sum: %w", err)
	}

	want := map[string]bool{}
	for _, m := range mods {
		want[m.Path+" "+m.Version] = true
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// fields[0] é o módulo; fields[1] é "vX.Y.Z" ou "vX.Y.Z/go.mod".
		ver := fields[1]
		if idx := strings.Index(ver, "/go.mod"); idx >= 0 {
			ver = ver[:idx]
		}
		if want[fields[0]+" "+ver] {
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	return lines, nil
}

// FindRepoRoot sobe a partir de start até achar o diretório que contém
// server/gameServer/go.mod (a raiz do repositório).
func FindRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "server", "gameServer", "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("raiz do repositório não encontrada (server/gameServer/go.mod)")
		}
		dir = parent
	}
}

// FrozenHash calcula o sha256 canônico de uma cópia congelada (diretório),
// percorrendo recursivamente e ignorando o arquivo engine.sum (o próprio hash) e
// diretórios testdata (goldens de replay não fazem parte do fecho do motor).
func FrozenHash(dir string) (string, error) {
	var rels []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if filepath.Base(rel) == "engine.sum" {
			return nil
		}
		rels = append(rels, rel)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(rels)

	h := sha256.New()
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "FILE %s\n%s\n", rel, Normalize(data))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ForbiddenDeps devolve os caminhos de importação do fecho que indicam go-redis,
// gorilla ou redisFuncs (violação do fecho puro do motor).
func ForbiddenDeps(gameServerDir string) ([]string, error) {
	lines, err := listDeps(gameServerDir)
	if err != nil {
		return nil, err
	}
	var bad []string
	for _, line := range lines {
		importPath := strings.SplitN(line, "|", 2)[0]
		if strings.Contains(importPath, "go-redis") ||
			strings.Contains(importPath, "gorilla") ||
			strings.Contains(importPath, "redisFuncs") {
			bad = append(bad, importPath)
		}
	}
	sort.Strings(bad)
	return bad, nil
}

// ForbiddenUses devolve as ocorrências de `time.Now`/`rand.` em arquivos .go de
// produção do fecho, fora de determinism.go (e fora de game/runtime, que não faz
// parte do fecho). Cada violação é "importPath:arquivo:linha:trecho".
func ForbiddenUses(gameServerDir string) ([]string, error) {
	pkgs, err := Packages(gameServerDir)
	if err != nil {
		return nil, err
	}

	var bad []string
	for _, pkg := range pkgs {
		dir := filepath.Join(gameServerDir, filepath.FromSlash(pkg.RelDir))
		for _, name := range pkg.Files {
			if name == "determinism.go" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return nil, err
			}
			for i, line := range strings.Split(string(data), "\n") {
				for _, needle := range []string{"time.Now", "rand."} {
					if strings.Contains(line, needle) {
						bad = append(bad, fmt.Sprintf("%s:%s:%d:%s",
							pkg.ImportPath, name, i+1, strings.TrimSpace(line)))
					}
				}
			}
		}
	}
	sort.Strings(bad)
	return bad, nil
}
