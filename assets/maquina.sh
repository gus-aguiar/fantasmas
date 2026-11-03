#!/bin/bash
# maquina.sh do fantasmas (embutido no binário; nasceu do shared/mods/control-panel/scripts do Spark,
# com disco e swap a mais). Roda com bash -c; $1 é a trava de testes pesados, vazia fora do Spark.
# -> linhas chave<TAB>rotulo<TAB>detalhe<TAB>tom  (tom: ok | aviso | erro | neutro)
# O que derrubou tudo em 06/10: memória (o oomd mata o terminal inteiro), sessões demais e testes pesados em paralelo.
set -u
l() { printf '%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$4"; }

read -r total disp <<<"$(awk '/MemTotal/{t=$2} /MemAvailable/{a=$2} END{print t, a}' /proc/meminfo)"
usado=$(( (total - disp) * 100 / total ))
tg=$(awk -v k="$total" 'BEGIN{printf "%.0f", k/1048576}')
ug=$(awk -v k="$((total - disp))" 'BEGIN{printf "%.0f", k/1048576}')
tom=ok; [ "$usado" -ge 80 ] && tom=aviso; [ "$usado" -ge 90 ] && tom=erro
l memoria "memória" "${ug} de ${tg} GB (${usado}%)" "$tom"

# swap cheio é o último degrau antes do oomd matar processos
read -r st sf <<<"$(awk '/SwapTotal/{t=$2} /SwapFree/{f=$2} END{print t+0, f+0}' /proc/meminfo)"
if [ "$st" -gt 0 ]; then
  su=$(( (st - sf) * 100 / st ))
  tom=ok; [ "$su" -ge 50 ] && tom=aviso; [ "$su" -ge 80 ] && tom=erro
  l swap "swap" "$(awk -v k="$((st - sf))" 'BEGIN{printf "%.1f", k/1048576}') de $(awk -v k="$st" 'BEGIN{printf "%.0f", k/1048576}') GB (${su}%)" "$tom"
else
  l swap "swap" "sem swap" neutro
fi

press=$(awk '/^some/{split($3,a,"="); print a[2]}' /proc/pressure/memory 2>/dev/null)
tom=ok; awk -v p="${press:-0}" 'BEGIN{exit !(p>10)}' && tom=aviso; awk -v p="${press:-0}" 'BEGIN{exit !(p>40)}' && tom=erro
l pressao "pressão de memória" "${press:-?}% (média 60s)" "$tom"

oom=$(journalctl --since today 2>/dev/null | grep -ciE 'systemd-oomd.*Killed|Out of memory: Killed')
tom=ok; [ "${oom:-0}" -gt 0 ] && tom=erro
l oom "mortes por memória hoje" "${oom:-0}" "$tom"

read -r c1 _ < /proc/loadavg; n=$(nproc)
tom=ok; awk -v c="$c1" -v n="$n" 'BEGIN{exit !(c>n)}' && tom=aviso; awk -v c="$c1" -v n="$n" 'BEGIN{exit !(c>3*n)}' && tom=erro
l carga "carga" "${c1} em ${n} núcleos" "$tom"

s=$(pgrep -fc '^claude' 2>/dev/null); s=${s:-0}
tom=ok; [ "$s" -gt 15 ] && tom=aviso; [ "$s" -gt 25 ] && tom=erro
l sessoes "sessões do Claude" "$s" "$tom"

# disco da home, onde ficam worktrees, caches e builds
read -r dt du dp <<<"$(df -P -k "$HOME" 2>/dev/null | awk 'NR==2{gsub("%","",$5); print $2, $3, $5}')"
if [ -n "${dt:-}" ]; then
  tom=ok; [ "$dp" -ge 85 ] && tom=aviso; [ "$dp" -ge 95 ] && tom=erro
  l disco "disco" "$(awk -v k="$du" 'BEGIN{printf "%.0f", k/1048576}') de $(awk -v k="$dt" 'BEGIN{printf "%.0f", k/1048576}') GB (${dp}%)" "$tom"
fi

# a trava vem como $1 (<raiz>/.delegate/heavy.lock); sem raiz do Spark não há trava.
trava="${1:-}"
[ -z "$trava" ] && exit 0
fila=$(pgrep -fc "flock $trava" 2>/dev/null); fila=${fila:-0}
dono=$(fuser "$trava" 2>/dev/null | awk '{print $1}')
cmd=""
[ -n "$dono" ] && cmd=$(ps -o args= --ppid "$dono" 2>/dev/null | head -1 | cut -c1-50)
[ -z "$cmd" ] && [ -n "$dono" ] && cmd=$(ps -o args= -p "$dono" 2>/dev/null | cut -c1-50)
if [ "$fila" -gt 0 ]; then
  l trava "trava de testes pesados" "${fila} na trava${cmd:+ · rodando: $cmd}" aviso
else
  l trava "trava de testes pesados" "livre" ok
fi
