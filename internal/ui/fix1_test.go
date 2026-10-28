package ui

import (
	"context"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fantasmas/assets"
	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func atualiza(t *testing.T, m Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	n, cmd := m.Update(msg)
	return n.(Model), cmd
}

// onde acha a linha e a coluna (em células) da primeira ocorrência de s a partir da linha de.
func onde(t *testing.T, m Model, s string, de int) (int, int) {
	t.Helper()
	ls := strings.Split(render(m, m.largura, m.altura), "\n")
	for y := de; y < len(ls); y++ {
		if i := strings.Index(ls[y], s); i >= 0 {
			return lipgloss.Width(ls[y][:i]), y
		}
	}
	t.Fatalf("%q não está na tela:\n%s", s, strings.Join(ls, "\n"))
	return 0, 0
}

func clique(t *testing.T, m Model, x, y int) (Model, tea.Cmd) {
	t.Helper()
	return atualiza(t, m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
}

func TestTickColetaUmaCadeiaSo(t *testing.T) {
	var n int32
	m := Novo(func(context.Context) dados.Estado { atomic.AddInt32(&n, 1); return fixture() }, Acoes{}, time.Millisecond)
	var ticks []tea.Cmd
	m, c := atualiza(t, m, coletouMsg{fixture()})
	ticks = append(ticks, c)
	for i := 0; i < 3; i++ {
		m, _ = atualiza(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		m, c = atualiza(t, m, coletouMsg{fixture()})
		ticks = append(ticks, c)
	}
	coletas := 0
	for _, tk := range ticks {
		var cmd tea.Cmd
		m, cmd = atualiza(t, m, tk())
		if cmd != nil {
			coletas++
			m, _ = atualiza(t, m, coletouMsg{fixture()})
		}
	}
	if coletas != 1 {
		t.Fatalf("ticks antigos dispararam %d coletas, esperava 1", coletas)
	}
}

func TestTopoESecaoContamIgual(t *testing.T) {
	e := fixture()
	e.Tasks = append(e.Tasks, dados.Linha{Chave: "t-f", Rotulo: "c-pl-f", Detalhe: "sem report · 0/0 · 1min · blocked"}, dados.Linha{Chave: "t-g", Rotulo: "c-pl-g", Detalhe: "sem report · 0/0 · 1min · blocked"})
	m := modelo(e, &registro{})
	m.largura, m.altura = 96, 40
	out := render(m, 96, 40)
	num := func(re string) string {
		g := regexp.MustCompile(re).FindStringSubmatch(out)
		if g == nil {
			t.Fatalf("sem %s em:\n%s", re, out)
		}
		return g[1]
	}
	if a, b := num(`(\d+) pra você`), num(`› pra você\s+(\d+)`); a != b {
		t.Fatalf("topo diz %s pra você, seção diz %s", a, b)
	}
	if a, b := num(`(\d+) travados?`), num(`› travados\s+(\d+)`); a != b {
		t.Fatalf("topo diz %s travados, seção diz %s", a, b)
	}
}

func TestEnvioNaoDuplica(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m, c1 := atualiza(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m, c2 := atualiza(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	_, y := onde(t, m, "Aprovar", m.altura/2)
	m, c3 := clique(t, m, margem+3, y)
	if c1 == nil {
		t.Fatal("primeiro enter não enviou")
	}
	for _, cmd := range []tea.Cmd{c2, c3} {
		if cmd != nil {
			if _, ok := cmd().(respondeuMsg); ok {
				t.Fatal("enviou de novo com resposta em voo")
			}
		}
	}
	m, _ = atualiza(t, m, c1())
	if len(c.resps) != 1 || m.enviando != "" {
		t.Fatalf("respostas %v, enviando %q", c.resps, m.enviando)
	}
}

func TestRespostaSemTabNemQuebra(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	m.rascunho.SetValue("a\tb")
	m = passo(t, m, "enter")
	if len(c.resps) != 1 || strings.ContainsAny(c.resps[0].resposta, "\t\n") {
		t.Fatalf("resps %q", c.resps)
	}
}

func TestCliqueNoFeedAbreComoEnter(t *testing.T) {
	c := novaCena(t)
	m := c.modelo(c.estado("plano", "juntar"))
	_, y := onde(t, m, "juntar", 0)
	m, _ = clique(t, m, 4, y)
	if m.aberto == nil || m.aberto.Merge == nil {
		t.Fatalf("clique não abriu o juntar; status=%q", m.status)
	}
}

func TestCliqueNaOpcaoResponde(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	ls := strings.Split(render(m, 96, 40), "\n")
	y := -1
	for i := len(ls) - 1; i >= 0; i-- {
		if strings.Contains(ls[i], "Aprovar") {
			y = i
			break
		}
	}
	m, cmd := clique(t, m, margem+4, y)
	if cmd == nil {
		t.Fatal("clique na opção não respondeu")
	}
	m, _ = atualiza(t, m, cmd())
	if len(c.resps) != 1 || c.resps[0].resposta != "Aprovar" {
		t.Fatalf("resps %v", c.resps)
	}
}

func TestCliqueVoltarFecha(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	x, y := onde(t, m, "‹ voltar", 0)
	if y != 0 {
		t.Fatalf("voltar na linha %d", y)
	}
	m, _ = clique(t, m, x+1, y)
	if m.aberto != nil {
		t.Fatal("voltar não fechou")
	}
}

func TestCliqueNaAbaTrocaGaveta(t *testing.T) {
	for _, w := range []int{56, 96} {
		m := modelo(fixture(), &registro{})
		m.largura, m.altura = w, 30
		rot := "PRs"
		if w < 70 {
			rot = " 3 "
		}
		x, y := onde(t, m, rot, 29)
		m, _ = clique(t, m, x+1, y)
		if Gavetas[m.gaveta].ID != "prs" {
			t.Fatalf("w=%d: gaveta %s", w, Gavetas[m.gaveta].ID)
		}
	}
}

func TestCliqueNoAtualizarColeta(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m.largura, m.altura = 96, 30
	x, y := onde(t, m, "↻", 0)
	m, cmd := clique(t, m, x, y)
	if !m.coletando || cmd == nil {
		t.Fatal("↻ não atualizou")
	}
}

func TestRodaRola(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m.largura, m.altura = 96, 30
	m, _ = atualiza(t, m, tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if m.sel != 1 {
		t.Fatalf("sel %d", m.sel)
	}
}

func TestSelecaoVisivelNoAgora(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m.largura, m.altura = 96, 30
	_, y := onde(t, m, "decidir", 0)
	l := strings.Split(m.Render(96, 30), "\n")[y]
	if !strings.Contains(l, "48;2;58;42;85") {
		t.Fatalf("linha selecionada sem fundo sel: %q", l)
	}
}

func TestAjudaAbreEFecha(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m.largura, m.altura = 96, 40
	m = tecla(m, "?")
	out := render(m, 96, 40)
	for _, s := range []string{"clique", "enter", "esc", "? fecha"} {
		if !strings.Contains(out, s) {
			t.Fatalf("ajuda sem %q:\n%s", s, out)
		}
	}
	m = tecla(m, "?")
	if m.ajuda {
		t.Fatal("? não fechou")
	}
}

func TestToastSomeSozinho(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m.largura, m.altura = 96, 30
	m, cmd := atualiza(t, m, respondeuMsg{Entrega: dados.Entrega{Crew: "c-pl-1", Resposta: "Aprovar", Entregue: true}})
	if cmd == nil || !strings.Contains(render(m, 96, 30), "entregue a c-pl-1") {
		t.Fatal("sem toast")
	}
	velho := m.genToast
	m, _ = atualiza(t, m, respondeuMsg{Entrega: dados.Entrega{Crew: "c-pl-2", Motivo: "x"}})
	m, _ = atualiza(t, m, toastFimMsg{velho})
	if m.toast == "" {
		t.Fatal("toast velho apagou o novo")
	}
	m, _ = atualiza(t, m, toastFimMsg{m.genToast})
	if m.toast != "" {
		t.Fatal("toast não sumiu")
	}
}

func TestGengarCelulaCheiaMaior(t *testing.T) {
	if MeioBloco {
		t.Fatal("padrão deve ser célula cheia")
	}
	s := spriteFn(96)
	if s.Colunas != 16 || s.Linhas != 7 {
		t.Fatalf("largo %dx%d", s.Colunas, s.Linhas)
	}
	if strings.Contains(strings.Join(s.Quadro(0), ""), "▀") {
		t.Fatal("ainda meio bloco")
	}
}

func TestGengarPreservaOlhosEDentes(t *testing.T) {
	re := regexp.MustCompile(`48;2;(\d+);(\d+);(\d+)m`)
	for _, tam := range [][3]int{{16, 7, 9}, {14, 6, 7}, {12, 5, 8}} {
		s, err := CarregarSprite(assets.Gengar, tam[0], tam[1], T.Mantle)
		if err != nil {
			t.Fatal(err)
		}
		olhos, dentes := 0, 0
		for _, q := range s.Quadros {
			olho, dente := false, false
			for _, g := range re.FindAllStringSubmatch(strings.Join(q, ""), -1) {
				r, gg, b := atoi(g[1]), atoi(g[2]), atoi(g[3])
				olho = olho || r > gg+50 && r > b+50
				dente = dente || r > 220 && gg > 220 && b > 220
			}
			if olho {
				olhos++
			}
			if dente {
				dentes++
			}
		}
		t.Logf("%dx%d: olho em %d/%d quadros, sorriso em %d/%d", s.Colunas, s.Linhas, olhos, len(s.Quadros), dentes, len(s.Quadros))
		if olhos*10 < len(s.Quadros)*tam[2] || dentes*10 < len(s.Quadros)*9 {
			t.Fatalf("%dx%d olho=%d dente=%d de %d", s.Colunas, s.Linhas, olhos, dentes, len(s.Quadros))
		}
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

func TestTopo56SemCorte(t *testing.T) {
	for _, w := range []int{56, 96} {
		topo := ansi.Strip(Topo(fixture(), w, 0, false, ""))
		if strings.Contains(topo, "…") {
			t.Fatalf("w=%d topo cortado:\n%s", w, topo)
		}
		for _, l := range strings.Split(topo, "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d larga", w)
			}
		}
	}
}

func TestDelayMinimo(t *testing.T) {
	if d := delayDe(1); d != 50*time.Millisecond {
		t.Fatalf("delay %v", d)
	}
}
