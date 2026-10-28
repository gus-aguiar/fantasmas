#!/usr/bin/env bash
# Renderiza uma gaveta da TUI com a demo e grava shots/<tela>.{ansi,html}; imprime o caminho do HTML.
set -euo pipefail

tela="${1:?uso: quadro.sh <tela> [largura=140] [altura=42]}"
largura="${2:-140}"
altura="${3:-42}"

app="$(cd "$(dirname "$0")/.." && pwd)"
spark="$(cd "$app/../../.." && pwd)"
bin="$app/bin/fantasmas"
mkdir -p "$app/shots"

if [ ! -x "$bin" ] || [ -n "$(find "$app" -path "$app/bin" -prune -o -path "$app/shots" -prune -o \( -name '*.go' -o -name go.mod -o -name go.sum \) -newer "$bin" -print -quit)" ]; then
  bash "$spark/scripts/orca-orchestration/heavy-run.sh" -C "$app" "fantasmas build" env GOTOOLCHAIN=local go build -o bin/fantasmas . >&2
fi

"$bin" --uma-vez --cor --demo --gaveta "$tela" --largura "$largura" --altura "$altura" > "$app/shots/$tela.ansi"
python3 -I "$app/ferramentas/ansi2html.py" --fundo '#1a1326' < "$app/shots/$tela.ansi" > "$app/shots/$tela.html"
echo "$app/shots/$tela.html"
