# Demo do fantasmas (~30 s): painel --demo passando pelas abas e pela paleta (ctrl+p).
# Gerar a partir da raiz do repo:  go build -o bin/fantasmas . && vhs docs/demo/fantasmas.tape
# Precisa de vhs, ttyd e ffmpeg no PATH.

Output docs/demo/fantasmas.gif
Output docs/demo/fantasmas.webm

Set Shell "bash"
Set FontSize 15
Set Width 1280
Set Height 740
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set Framerate 24

# Ambiente limpo: binário do repo, HOME temporário (sem ajustes locais) e figura própria no topo.
Hide
Type "export PATH=$PWD/bin:$PATH FANTASMAS_SPRITE=$PWD/docs/demo/fantasminha.gif HOME=$(mktemp -d) PS1='$ ' && clear"
Enter
Sleep 500ms
Show

Type "fantasmas --demo"
Sleep 600ms
Enter
Sleep 4s

# 2 Missões
Type "2"
Sleep 3.5s
# 3 PRs
Type "3"
Sleep 3.5s
# 4 Máquina
Type "4"
Sleep 2.5s
# 5 Configurações
Type "5"
Sleep 3s
# 1 Agora
Type "1"
Sleep 2.5s

# Paleta de comandos
Ctrl+P
Sleep 4s
Escape
Sleep 2s

Hide
Type "q"
Sleep 800ms
