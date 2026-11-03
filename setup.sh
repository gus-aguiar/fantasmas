#!/usr/bin/env bash
# setup.sh: instala o Go (se faltar), compila bin/fantasmas e cria o comando gengar no PATH.
# Pode rodar de novo quando quiser: é idempotente.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${GENGAR_BIN_DIR:-$HOME/.local/bin}"
GO_DIR="${GENGAR_GO_DIR:-$HOME/.local/share/go}"
LINK="$BIN_DIR/gengar"
ALVO="$REPO/bin/fantasmas"
PATH_ORIGINAL="$PATH"
TMP=""
trap '[ -z "$TMP" ] || rm -rf "$TMP"' EXIT

info() { printf '\033[35m▸\033[0m %s\n' "$*"; }
erro() { printf '\033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

# Versão mínima do Go vem do go.mod (linha "go 1.xx").
GO_MIN="$(awk '$1 == "go" { print $2; exit }' "$REPO/go.mod")"
[ -n "$GO_MIN" ] || erro "não achei a versão do Go no go.mod"

# Devolve 0 se a versão $1 (ex.: 1.22.5) é >= $2 (ex.: 1.22).
versao_ok() {
	local a b
	IFS=. read -r -a a <<<"$1"
	IFS=. read -r -a b <<<"$2"
	for i in 0 1 2; do
		local x="${a[i]:-0}" y="${b[i]:-0}"
		((10#$x > 10#$y)) && return 0
		((10#$x < 10#$y)) && return 1
	done
	return 0
}

go_ok() {
	command -v go >/dev/null 2>&1 || return 1
	local v
	v="$(GOTOOLCHAIN=local go env GOVERSION 2>/dev/null)" || return 1
	v="${v#go}"
	versao_ok "${v%%[!0-9.]*}" "$GO_MIN"
}

instalar_go_tarball() {
	local os arch versao
	case "$(uname -s)" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) erro "sistema não suportado: $(uname -s)" ;;
	esac
	case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) erro "arquitetura não suportada: $(uname -m)" ;;
	esac
	command -v curl >/dev/null 2>&1 || erro "preciso do curl para baixar o Go"
	versao="$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n1)"
	[ -n "$versao" ] || erro "não consegui descobrir a última versão do Go"
	info "baixando $versao ($os-$arch) para $GO_DIR"
	TMP="$(mktemp -d)"
	curl -fsSL "https://go.dev/dl/$versao.$os-$arch.tar.gz" -o "$TMP/go.tgz"
	rm -rf "$GO_DIR"
	mkdir -p "$GO_DIR"
	tar -xzf "$TMP/go.tgz" -C "$GO_DIR" --strip-components=1
	mkdir -p "$BIN_DIR"
	ln -sfn "$GO_DIR/bin/go" "$BIN_DIR/go"
	ln -sfn "$GO_DIR/bin/gofmt" "$BIN_DIR/gofmt"
	export PATH="$BIN_DIR:$PATH"
}

# 1. Go
if go_ok; then
	info "Go já instalado: $(go env GOVERSION)"
elif command -v mise >/dev/null 2>&1; then
	info "instalando Go pelo mise"
	mise use -g go@latest
	export PATH="$(mise where go)/bin:$PATH"
else
	instalar_go_tarball
fi
go_ok || erro "Go >= $GO_MIN não ficou disponível no PATH"

# 2. Build (GOTOOLCHAIN=local igual ao botão atualizar do app)
info "compilando $ALVO"
mkdir -p "$REPO/bin"
GOTOOLCHAIN=local go -C "$REPO" build -o "$ALVO.novo" .
mv -f "$ALVO.novo" "$ALVO"

# 3. Comando gengar: symlink, para o atualizador achar o repo a partir do binário.
mkdir -p "$BIN_DIR"
if [ -e "$LINK" ] && [ ! -L "$LINK" ]; then
	erro "$LINK já existe e não é symlink; remova e rode de novo"
fi
ln -sfn "$ALVO" "$LINK"
info "gengar -> $ALVO"

# 4. PATH
case ":$PATH_ORIGINAL:" in
*":$BIN_DIR:"*) ;;
*)
	case "$(basename "${SHELL:-bash}")" in
	zsh) rc="$HOME/.zshrc" ;;
	*) rc="$HOME/.bashrc" ;;
	esac
	linha="export PATH=\"$BIN_DIR:\$PATH\""
	if ! grep -qsF "$linha" "$rc"; then
		printf '\n# gengar (fantasmas)\n%s\n' "$linha" >>"$rc"
		info "adicionei $BIN_DIR ao PATH em $rc"
	fi
	info "abra um terminal novo (ou rode: source $rc)"
	;;
esac

info "pronto: rode 'gengar' de qualquer lugar"
