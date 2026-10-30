package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fantasmas/internal/dados"
	"fantasmas/internal/kitty"
	"fantasmas/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/termenv"
)

func main() {
	raiz := flag.String("raiz", "", "raiz do workspace Spark")
	intervalo := flag.Duration("intervalo", 30*time.Second, "intervalo da coleta automática")
	largura := flag.Int("largura", 0, "largura do quadro (só com --uma-vez)")
	altura := flag.Int("altura", 0, "altura do quadro (só com --uma-vez)")
	umaVez := flag.Bool("uma-vez", false, "coleta, imprime um quadro e sai")
	sprite := flag.String("sprite", os.Getenv("FANTASMAS_SPRITE"), "GIF do topo no lugar do Gengar embutido (ou FANTASMAS_SPRITE)")
	cor := flag.Bool("cor", false, "força cor de 24 bits (só com --uma-vez)")
	demo := flag.Bool("demo", false, "usa o estado de demonstração (só com --uma-vez)")
	demoVerde := flag.Bool("demo-verde", false, "demonstração com todos os PRs de CI verde (só com --uma-vez)")
	gaveta := flag.String("gaveta", "", "gaveta inicial: agora|crew|pipes|prs|ci|ambiente|maquina|mais (só com --uma-vez)")
	pipe := flag.Int("pipe", 0, "abre o detalhe do run N da gaveta pipes (só com --uma-vez)")
	flag.Parse()
	if *intervalo <= 0 {
		fmt.Fprintln(os.Stderr, "fantasmas: --intervalo deve ser positivo")
		os.Exit(2)
	}
	if err := validarGaveta(*gaveta); err != nil {
		fmt.Fprintln(os.Stderr, "fantasmas:", err)
		os.Exit(2)
	}
	if *raiz == "" {
		*raiz, _ = os.Getwd()
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
		Pipes:          &dados.Pipes{R: runner, Raiz: *raiz, Script: dados.ScriptDdb(*raiz)},
		Editar:         func(arq string) *exec.Cmd { return ui.ComandoEditor(os.Getenv("FANTASMAS_EDITOR"), arq) },
	}
	m := ui.Novo(coleta, acoes, *intervalo)
	if *sprite != "" {
		if err := ui.UsarSprite(*sprite); err != nil {
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
		if *gaveta != "" {
			m = m.ComGaveta(*gaveta)
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
	_, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithOutput(saida)).Run()
	if g != nil {
		g.Encerrar()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fantasmas:", err)
		os.Exit(1)
	}
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
	for _, k := range []string{"HERDR_PANE_ID", "WEZTERM_EXECUTABLE", "KITTY_WINDOW_ID"} {
		if env(k) != "" {
			return "kitty", k + " presente", false
		}
	}
	return "", "nenhum indício no ambiente", true
}

// escolherModoImagem: FANTASMAS_IMAGEM_MODO manda; sem ele, T dentro do Herdr e animação fora.
func escolherModoImagem(env func(string) string) string {
	switch v := env("FANTASMAS_IMAGEM_MODO"); v {
	case ui.ModoAnimacao, ui.ModoQuadros, ui.ModoT:
		return v
	}
	if env("HERDR_PANE_ID") != "" {
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

// validarGaveta recusa nome desconhecido em vez de cair calado na primeira gaveta.
func validarGaveta(id string) error {
	if id == "" || ui.GavetaValida(id) {
		return nil
	}
	return fmt.Errorf("--gaveta %q desconhecida; válidas: %s", id, strings.Join(ui.IDsGavetas(), ", "))
}
