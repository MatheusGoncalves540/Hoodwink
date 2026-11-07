// Package version carrega a versão do jogo embutida no binário.
//
// A fonte da verdade é o arquivo VERSION na raiz do repositório. O build injeta
// os valores por ldflags (ver gameServer/Dockerfile e server/makefile):
//
//	-X github.com/MatheusGoncalves540/Hoodwink-gameServer/version.Version=<VERSION>
//	-X github.com/MatheusGoncalves540/Hoodwink-gameServer/version.Commit=<sha curto>
//
// Sem ldflags (go run/go test) os valores ficam em "dev"/"unknown".
// Esta versão é a do JOGO e é independente da versão do motor de replay
// (game/ENGINE_VERSION e replay/engines/vN).
package version

var (
	// Version é a versão semver do jogo (ex.: 0.1.0).
	Version = "dev"
	// Commit é o SHA curto do commit que gerou o binário.
	Commit = "unknown"
)
