#!/bin/sh
# Controle de versão do JOGO (semver). A fonte da verdade é VERSION na raiz do
# repositório; client/web/package.json espelha o mesmo número para o npm.
#
# Uso (a partir da raiz do repo, ou via `make -C server ...`):
#   sh server/scripts/game-version.sh show
#   sh server/scripts/game-version.sh check [--tag v0.1.0]
#   sh server/scripts/game-version.sh bump [major|minor|patch]
#   sh server/scripts/game-version.sh set 0.2.0
#
# A versão do jogo NÃO é a versão do motor de replay (game/ENGINE_VERSION e
# replay/engines/vN), que continua sendo um inteiro próprio.
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
version_file="$repo_root/VERSION"
package_file="$repo_root/client/web/package.json"
semver_re='^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'

die() {
	echo "game-version: $*" >&2
	exit 1
}

read_version() {
	[ -f "$version_file" ] || die "arquivo VERSION nao encontrado em $version_file"
	v=$(tr -d '\r\n' <"$version_file")
	[ -n "$v" ] || die "VERSION esta vazio"
	printf '%s' "$v"
}

package_version() {
	[ -f "$package_file" ] || die "package.json nao encontrado em $package_file"
	sed -n 's/^[[:space:]]*"version":[[:space:]]*"\([^"]*\)".*/\1/p' "$package_file" | head -n 1
}

validate() {
	v=$1
	printf '%s' "$v" | grep -Eq "$semver_re" || die "versao invalida: '$v' (esperado MAJOR.MINOR.PATCH)"
}

write_all() {
	v=$1
	validate "$v"
	printf '%s\n' "$v" >"$version_file"
	tmp=$(mktemp)
	awk -v v="$v" '
		!done && /^[[:space:]]*"version":[[:space:]]*"/ {
			sub(/"version":[[:space:]]*"[^"]*"/, "\"version\": \"" v "\"")
			done = 1
		}
		{ print }
	' "$package_file" >"$tmp"
	mv "$tmp" "$package_file"
}

cmd_show() {
	printf '%s\n' "$(read_version)"
}

cmd_check() {
	v=$(read_version)
	validate "$v"
	pkg=$(package_version)
	[ "$pkg" = "$v" ] || die "client/web/package.json esta em $pkg, mas VERSION esta em $v; rode make game-bump"

	if [ "$#" -gt 0 ]; then
		[ "$1" = "--tag" ] && [ "$#" -ge 2 ] || die "uso: check [--tag vX.Y.Z]"
		tag=$(printf '%s' "$2" | sed 's/^v//')
		[ "$tag" = "$v" ] || die "a tag e v$tag, mas VERSION esta em $v"
	fi
	printf 'game-version: %s ok\n' "$v"
}

cmd_bump() {
	level=${1:-patch}
	cur=$(read_version)
	core=$(printf '%s' "$cur" | sed 's/[-+].*$//')
	major=$(printf '%s' "$core" | cut -d. -f1)
	minor=$(printf '%s' "$core" | cut -d. -f2)
	patch=$(printf '%s' "$core" | cut -d. -f3)
	case "$level" in
	major)
		major=$((major + 1))
		minor=0
		patch=0
		;;
	minor)
		minor=$((minor + 1))
		patch=0
		;;
	patch) patch=$((patch + 1)) ;;
	*) die "nivel invalido: $level (use major, minor ou patch)" ;;
	esac
	next="$major.$minor.$patch"
	write_all "$next"
	printf 'game-version: %s -> %s\n' "$cur" "$next"
}

cmd_set() {
	[ "$#" -ge 1 ] || die "uso: set X.Y.Z"
	write_all "$1"
	printf 'game-version: %s\n' "$1"
}

command=${1:-show}
if [ "$#" -gt 0 ]; then
	shift
fi

case "$command" in
show) cmd_show ;;
check) cmd_check "$@" ;;
bump) cmd_bump "$@" ;;
set) cmd_set "$@" ;;
*)
	die "comando desconhecido: $command (use show, check, bump ou set)"
	;;
esac
