package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"fantasmas/internal/config"
	"fantasmas/internal/dados"
	"fantasmas/internal/kitty"
	"fantasmas/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"
)

func main() {
	// antes de qualquer os.Getenv, inclusive os padrões das flags
	loja, aviso := abrirAjustes()
	raiz := flag.String("raiz", "", "raiz do workspace Spark (ou FANTASMAS_RAIZ; sem as duas, a pasta atual)")
	intervalo := flag.Duration("intervalo", 30*time.Second, "intervalo da coleta automática")
	largura := flag.Int("largura", 0, "largura do quadro (só com --uma-vez)")
	altura := flag.Int("altura", 0, "altura do quadro (só com --uma-vez)")
	umaVez := flag.Bool("uma-vez", false, "coleta, imprime um quadro e sai")
	sprite := flag.String("sprite", os.Getenv("FANTASMAS_SPRITE"), "GIF do topo no lugar do Gengar embutido (ou FANTASMAS_SPRITE)")
	cor := flag.Bool("cor", false, "força cor de 24 bits (só com --uma-vez)")
	demo := flag.Bool("demo", false, "usa o estado de demonstração (só com --uma-vez)")
	demoVerde := flag.Bool("demo-verde", false, "demonstração com todos os PRs de CI verde (só com --uma-vez)")
	gaveta := flag.String("gaveta", "", "gaveta inicial: missoes|agora|crew|pipes|prs|ambiente|maquina|worktrees|configuracoes (só com --uma-vez)")
	missao := flag.String("missao", "", "foca a missão pelo slug (só com --uma-vez)")
	pipe := flag.Int("pipe", 0, "abre o detalhe do run N da gaveta pipes (só com --uma-vez)")
	acharRaiz := flag.Bool("achar-raiz", false, "procura a raiz do Spark na home, salva em configurações (FANTASMAS_RAIZ) e sai; o setup.sh usa")
	flag.Parse()
	if *acharRaiz {
		var l lojaRaiz
		if loja != nil {
			l = loja
		}
		home, _ := os.UserHomeDir()
		achar := func() []string {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			return dados.AcharRaizes(ctx, home, profBuscaRaiz)
		}
		os.Exit(configurarRaiz(l, achar, home, os.Stdin, os.Stdout, term.IsTerminal(os.Stdin.Fd())))
	}
	if *intervalo <= 0 {
		fmt.Fprintln(os.Stderr, "fantasmas: --intervalo deve ser positivo")
		os.Exit(2)
	}
	if err := validarGaveta(*gaveta); err != nil {
		fmt.Fprintln(os.Stderr, "fantasmas:", err)
		os.Exit(2)
	}
	var avisoRaiz string
	*raiz, avisoRaiz = resolverRaiz(*raiz, os.Getenv, os.Getwd, os.UserHomeDir, os.Stat)
	if avisoRaiz != "" {
		aviso = avisoRaiz
	}

	runner := dados.ExecRunner{}
	coletor := &dados.Coletor{Raiz: *raiz, R: runner}
	var mu sync.Mutex
	var recursos dados.Recursos
	coleta := func(ctx context.Context) dados.Estado {
		e := coletor.Coletar(ctx)
		mu.Lock()
		recursos = e.Recursos
		mu.Unlock()
		return e
	}
	acoes := ui.Acoes{
		AbrirURL: abrir,
		FocarAba: func(aba string) error { return exec.Command("herdr", "tab", "focus", aba).Run() },
		Responder: func(chave, resposta string) (dados.Entrega, error) {
			mu.Lock()
			rec := recursos
			mu.Unlock()
			return dados.Responder(context.Background(), runner, *raiz, rec, chave, resposta)
		},
		LerArquivo: os.ReadFile,
		Arquivar:   dados.Arquivador(runner, *raiz),
		Retomar: func(c dados.Caido) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			return dados.Retomar(ctx, runner, c)
		},
		MarcarVisto: dados.MarcarVisto,
		Daydreaming: func() (string, error) {
			brief := filepath.Join(*raiz, "shared", "tui", "fantasmas", "assets", "daydreamer-brief.md")
			return dados.AtualizarDaydreaming(context.Background(), runner, *raiz, brief)
		},
		ReconectarMCPs: func() error { return dados.ReconectarMCPs(*raiz) },
		ProgressoMCP:   func() dados.ProgressoMCP { return dados.LerProgressoMCP(*raiz, time.Now()) },
		Atualizar: func() (dados.Resultado, error) {
			exe, err := os.Executable()
			if err != nil {
				return dados.Resultado{}, fmt.Errorf("repo não encontrado: %w", err)
			}
			repo, err := dados.AcharRepo(context.Background(), runner, exe)
			if err != nil {
				return dados.Resultado{}, err
			}
			return dados.Atualizar(context.Background(), runner, repo)
		},
		Pipes:  &dados.Pipes{R: runner, Raiz: *raiz, Script: dados.ScriptDdb(*raiz)},
		Editar: func(arq string) *exec.Cmd { return ui.ComandoEditor(os.Getenv("FANTASMAS_EDITOR"), arq) },
	}
	if loja != nil {
		acoes.Ajustes = loja
		acoes.DirFiguras = filepath.Join(filepath.Dir(loja.Caminho()), "figuras")
	}
	config.DefinirValores(config.ChaveFigura, ui.NomesFiguras(acoes.DirFiguras))
	if err := ui.UsarTema(os.Getenv(config.ChaveTema)); err != nil {
		aviso = err.Error()
	}
	m := ui.Novo(coleta, acoes, *intervalo)
	if aviso != "" {
		m = m.ComStatus(aviso)
	}
	if *sprite != "" {
		if err := ui.UsarSprite(*sprite); err != nil {
			m = m.ComStatus(err.Error())
		}
	} else if f := os.Getenv(config.ChaveFigura); f != "" {
		if err := ui.UsarFigura(f, acoes.DirFiguras); err != nil {
			m = m.ComStatus(err.Error())
		}
	}

	if *umaVez {
		w, h := *largura, *altura
		if w <= 0 {
			w = 96
		}
		if h <= 0 {
			h = 30
		}
		if *cor {
			lipgloss.SetColorProfile(termenv.TrueColor)
		}
		if g := gavetaInicial(*gaveta, os.Getenv); g != "" {
			m = m.ComGaveta(g)
		}
		est := func() dados.Estado { return coleta(context.Background()) }
		if *demo {
			est = func() dados.Estado { return ui.EstadoDemo(time.Now()) }
		}
		if *demoVerde {
			est = func() dados.Estado { return ui.EstadoDemoVerde(time.Now()) }
		}
		m = m.ComEstado(est())
		if *pipe != 0 {
			m = m.ComPipe(*pipe)
		}
		if *missao != "" {
			m = m.ComMissao(*missao)
		}
		fmt.Println(m.Render(w, h))
		return
	}
	logf := abrirLog(os.Getenv("FANTASMAS_DEBUG"))
	for _, k := range []string{"TERM", "TERM_PROGRAM", "HERDR_PANE_ID", "WEZTERM_EXECUTABLE", "KITTY_WINDOW_ID", "FANTASMAS_IMAGEM", "FANTASMAS_IMAGEM_MODO"} {
		logf("env %s=%q", k, os.Getenv(k))
	}
	saida := &kitty.Saida{File: os.Stdout}
	var g *ui.Grafico
	modo, motivo, consultar := escolherModo(os.Getenv)
	if consultar {
		if consultarTerminal() {
			modo, motivo = "kitty", motivo+"; consulta respondeu OK"
		} else {
			modo, motivo = "celulas", motivo+"; consulta sem OK em 300 ms"
		}
	}
	rotulo := "células"
	if modo == "kitty" {
		modoImg := escolherModoImagem(os.Getenv)
		var err error
		if g, err = ui.NovoGrafico(saida, modoImg); err == nil {
			g.Log = func(s string) { logf("envio %s", s) }
			m = m.ComGrafico(g)
			rotulo = map[string]string{ui.ModoAnimacao: "kitty-animação", ui.ModoQuadros: "kitty-quadros", ui.ModoT: "kitty-T"}[modoImg]
			for i, d := range g.Dimensoes() {
				logf("png %d: %s", i, d)
			}
		} else {
			motivo += "; erro ao montar PNG: " + err.Error()
		}
	}
	logf("modo %s (%s)", rotulo, motivo)
	m = m.ComModoImagem(rotulo)
	final, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithOutput(saida)).Run()
	// a figura pode ter trocado pela aba configurações: encerra a imagem que está no terminal agora.
	if fm, ok := final.(ui.Model); ok && fm.Grafico() != nil {
		fm.Grafico().Encerrar()
	} else if g != nil {
		g.Encerrar()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fantasmas:", err)
		os.Exit(1)
	}
	if fm, ok := final.(ui.Model); ok && fm.Reiniciar() != "" {
		// o tea já saiu do alt-screen e restaurou o terminal; troca o processo pelo binário novo.
		err := syscall.Exec(fm.Reiniciar(), os.Args, os.Environ())
		fmt.Fprintln(os.Stderr, "fantasmas: atualizado; feche e abra de novo:", err)
	}
}

// resolverRaiz: --raiz manda; sem ela, FANTASMAS_RAIZ (aceita ~/); sem as duas, a pasta atual.
// FANTASMAS_RAIZ que não é pasta cai na pasta atual com um aviso que não mostra o caminho,
// porque o valor pode vir da aba configurações, que nunca o exibe.
func resolverRaiz(flag string, getenv func(string) string, getwd func() (string, error),
	home func() (string, error), stat func(string) (os.FileInfo, error)) (string, string) {
	atual := func() string { d, _ := getwd(); return d }
	if flag != "" {
		return flag, ""
	}
	r := strings.TrimSpace(getenv("FANTASMAS_RAIZ"))
	if r == "" {
		return atual(), ""
	}
	if r == "~" || strings.HasPrefix(r, "~/") {
		if h, err := home(); err == nil {
			r = filepath.Join(h, strings.TrimPrefix(r, "~"))
		}
	}
	if st, err := stat(r); err != nil || !st.IsDir() {
		return atual(), "FANTASMAS_RAIZ não é uma pasta; usei a pasta atual"
	}
	return r, ""
}

// profBuscaRaiz: quantos níveis abaixo da home o --achar-raiz desce (~/orca/workspaces/spark/<ws> é 4).
const profBuscaRaiz = 5

// lojaRaiz é o que configurarRaiz usa da config.Loja.
type lojaRaiz interface {
	Origem(chave string) config.Origem
	Valor(chave string) string
	Definir(chave, valor string) error
}

// configurarRaiz é o --achar-raiz: mantém a FANTASMAS_RAIZ que já serve; senão procura e salva. Com uma
// raiz achada, não mostra o caminho (a aba configurações também não mostra); com várias, lista para escolher
// no terminal, ou fica com a mais ativa sem terminal. Nunca falha o setup: só devolve 1 sem configurações.
func configurarRaiz(l lojaRaiz, achar func() []string, home string, entrada io.Reader, saida io.Writer, interativo bool) int {
	const chave = "FANTASMAS_RAIZ"
	diz := func(f string, a ...any) { fmt.Fprintf(saida, "▸ "+f+"\n", a...) }
	if l == nil {
		diz("sem arquivo de configurações; não salvei a raiz do Spark")
		return 1
	}
	switch l.Origem(chave) {
	case config.Ambiente:
		diz("%s vem do shell; não mexi", chave)
		return 0
	case config.Arquivo:
		if v := expandirHome(l.Valor(chave), home); dados.EhRaizSpark(v) {
			diz("%s já está em configurações e tem os scripts do Spark; não mexi", chave)
			return 0
		}
		diz("a %s salva não tem os scripts do Spark; procurando outra", chave)
	}
	rs := achar()
	curto := func(p string) string {
		if home != "" && strings.HasPrefix(p, home+string(os.PathSeparator)) {
			return "~" + p[len(home):]
		}
		return p
	}
	escolha := 0
	switch {
	case len(rs) == 0:
		diz("não achei a raiz do Spark na home (até %d níveis)", profBuscaRaiz)
		diz("defina em configurações (8 → %s) ou rode: gengar --raiz <pasta do Spark>", chave)
		return 0
	case len(rs) > 1 && interativo:
		diz("achei %d raízes do Spark (da mais ativa para a menos ativa):", len(rs))
		for i, r := range rs {
			fmt.Fprintf(saida, "    %d) %s\n", i+1, curto(r))
		}
		fmt.Fprintf(saida, "  qual usar? [1] ")
		var resp string
		fmt.Fscanln(entrada, &resp)
		if n, err := strconv.Atoi(strings.TrimSpace(resp)); err == nil && n >= 1 && n <= len(rs) {
			escolha = n - 1
		}
	case len(rs) > 1:
		diz("achei %d raízes do Spark; fiquei com a mais ativa (troque em configurações)", len(rs))
	}
	if err := l.Definir(chave, rs[escolha]); err != nil {
		diz("não salvei a raiz: %v", err)
		return 0
	}
	diz("raiz do Spark salva em configurações (%s); o gengar abre nela de qualquer pasta", chave)
	return 0
}

// expandirHome troca o ~ inicial pela home, como o resolverRaiz.
func expandirHome(p, home string) string {
	if home != "" && (p == "~" || strings.HasPrefix(p, "~/")) {
		return filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return p
}

// abrirAjustes aplica o arquivo da aba configurações ao ambiente; o aviso vai para a linha de status.
func abrirAjustes() (*config.Loja, string) {
	caminho, err := config.Caminho()
	if err != nil {
		return nil, err.Error()
	}
	l, err := config.Abrir(caminho)
	if err != nil {
		return l, err.Error()
	}
	return l, ""
}

// escolherModo decide kitty ou células pelo ambiente; consultar indica que só a consulta ao terminal decide.
func escolherModo(env func(string) string) (modo, motivo string, consultar bool) {
	switch v := env("FANTASMAS_IMAGEM"); v {
	case "kitty", "1":
		return "kitty", "FANTASMAS_IMAGEM=" + v, false
	case "celulas", "0":
		return "celulas", "FANTASMAS_IMAGEM=" + v, false
	}
	switch tp := env("TERM_PROGRAM"); tp {
	case "WezTerm", "kitty":
		return "kitty", "TERM_PROGRAM=" + tp, false
	}
	if v := env("ORCA_IMAGE_PROTOCOL"); v == "kitty" {
		return "kitty", "ORCA_IMAGE_PROTOCOL=" + v, false
	}
	for _, k := range []string{"HERDR_PANE_ID", "WEZTERM_EXECUTABLE", "KITTY_WINDOW_ID"} {
		if env(k) != "" {
			return "kitty", k + " presente", false
		}
	}
	return "", "nenhum indício no ambiente", true
}

// escolherModoImagem: FANTASMAS_IMAGEM_MODO manda; sem ele, T dentro do Herdr e do Orca e animação fora.
// O terminal do Orca (addon de imagem do xterm.js) só aceita as ações t, T, q, p e d: sem a=f e a=a
// a animação nativa fica parada no primeiro quadro, então o modo T reenvia cada quadro.
func escolherModoImagem(env func(string) string) string {
	switch v := env("FANTASMAS_IMAGEM_MODO"); v {
	case ui.ModoAnimacao, ui.ModoQuadros, ui.ModoT:
		return v
	}
	if env("HERDR_PANE_ID") != "" || env("TERM_PROGRAM") == "Orca" || env("ORCA_IMAGE_PROTOCOL") != "" {
		return ui.ModoT
	}
	return ui.ModoAnimacao
}

func abrirLog(caminho string) func(string, ...any) {
	if caminho == "" {
		return func(string, ...any) {}
	}
	f, err := os.OpenFile(caminho, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return func(string, ...any) {}
	}
	var mu sync.Mutex
	return func(s string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		fmt.Fprintf(f, time.Now().Format("15:04:05.000")+" "+s+"\n", a...)
	}
}

func consultarTerminal() bool {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return false
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	defer tty.Close()
	st, err := term.MakeRaw(tty.Fd())
	if err != nil {
		return false
	}
	defer term.Restore(tty.Fd(), st)
	return kitty.Detectar(tty, tty, 300*time.Millisecond)
}

func abrir(alvo string) error {
	c := comandoAbrir(runtime.GOOS, os.Getenv("FANTASMAS_EDITOR"), alvo)
	return exec.Command(c[0], c[1:]...).Start()
}

// comandoAbrir: link vai ao abridor do sistema; arquivo ou pasta vai ao FANTASMAS_EDITOR, se houver.
func comandoAbrir(goos, editor, alvo string) []string {
	if ed := strings.Fields(editor); len(ed) > 0 && !strings.Contains(alvo, "://") {
		return append(ed, alvo)
	}
	switch goos {
	case "darwin":
		return []string{"open", alvo}
	case "linux":
		return []string{"xdg-open", alvo}
	}
	return []string{"cmd", "/c", "start", "", alvo}
}

// gavetaInicial: a flag --gaveta manda; sem ela, FANTASMAS_FONTE=ddb abre em "missoes". Vazio = a primeira.
func gavetaInicial(flag string, getenv func(string) string) string {
	if flag != "" {
		return flag
	}
	if getenv("FANTASMAS_FONTE") == "ddb" {
		return "missoes"
	}
	return ""
}

// validarGaveta recusa nome desconhecido em vez de cair calado na primeira gaveta.
func validarGaveta(id string) error {
	if id == "" || ui.GavetaValida(id) {
		return nil
	}
	return fmt.Errorf("--gaveta %q desconhecida; válidas: %s", id, strings.Join(ui.IDsGavetas(), ", "))
}
