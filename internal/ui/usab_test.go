package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func aperta(t *testing.T, m Model, k string) Model {
	t.Helper()
	var msg tea.KeyMsg
	switch k {
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	n, _ := m.Update(msg)
	return n.(Model)
}

func digita(t *testing.T, m Model, s string) Model {
	for _, r := range s {
		m = aperta(t, m, string(r))
	}
	return m
}

func cabeTudo(t *testing.T, m Model, w, h int) {
	t.Helper()
	ls := strings.Split(m.Render(w, h), "\n")
	if len(ls) != h {
		t.Fatalf("%dx%d: %d linhas", w, h, len(ls))
	}
	for i, l := range ls {
		if lipgloss.Width(l) > w {
			t.Fatalf("%dx%d linha %d larga: %d", w, h, i, lipgloss.Width(l))
		}
	}
}

// ---- A: filtro ----

func TestFiltroNoFeed(t *testing.T) {
	r := &registro{}
	m := modelo(fixture(), r)
	m.largura, m.altura = 96, 40
	m = aperta(t, m, "/")
	if m.filtro == nil {
		t.Fatal("/ não abriu o filtro")
	}
	m = digita(t, m, "PL-2")
	out := render(m, 96, 40)
	if !strings.Contains(out, "c-pl-2") || strings.Contains(out, "c-pl-1") {
		t.Fatalf("filtro não filtrou:\n%s", out)
	}
	if !strings.Contains(out, "1 de ") {
		t.Fatalf("sem contador:\n%s", out)
	}
	m = aperta(t, m, "enter")
	if m.filtro != nil || m.busca == "" {
		t.Fatal("enter deveria fechar o campo e manter o filtro")
	}
	m = aperta(t, m, "o")
	if len(r.urls) != 1 || r.urls[0] != "https://github.com/o/PrologWebReact/pull/77" {
		t.Fatalf("a seleção não bate com o visível: %v", r.urls)
	}
	m = aperta(t, m, "esc")
	if m.busca != "" || !strings.Contains(render(m, 96, 40), "c-pl-1") {
		t.Fatal("esc não limpou o filtro")
	}
}

func TestFiltroNasGavetasEClique(t *testing.T) {
	r := &registro{}
	m := modelo(fixture(), r)
	m.largura, m.altura = 96, 40
	m = aperta(t, m, "3")
	m = aperta(t, m, "/")
	m = digita(t, m, "#4")
	out := render(m, 96, 40)
	if !strings.Contains(out, "WebService #4") || strings.Contains(out, "WebService #2") || !strings.Contains(out, "1 de 5") {
		t.Fatalf("gaveta PRs não filtrou:\n%s", out)
	}
	x, y := onde(t, m, "WebService #4", 0)
	m, _ = clique(t, m, x, y)
	if len(r.urls) != 1 || !strings.HasSuffix(r.urls[0], "/pull/4") {
		t.Fatalf("clique não bateu com o visível: %v", r.urls)
	}
	m = aperta(t, m, "esc")
	m = aperta(t, m, "esc")
	if m.filtro != nil || m.busca != "" {
		t.Fatal("esc deveria fechar e limpar")
	}
}

func TestFiltroSemResultado(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m = aperta(t, m, "/")
	m = digita(t, m, "zzz")
	out := render(m, 96, 40)
	if !strings.Contains(out, "0 de ") || !strings.Contains(out, "nada bate") {
		t.Fatalf("sem aviso de vazio:\n%s", out)
	}
	for _, w := range []int{46, 56, 96} {
		cabeTudo(t, m, w, 30)
	}
}

// ---- B: confirmação em 2 passos ----

func abrirJuntar(t *testing.T, c *cena) Model {
	t.Helper()
	m := c.modelo(c.estado("juntar"))
	return passo(t, m, "enter")
}

func TestLiberarMergePedeConfirmacao(t *testing.T) {
	c := novaCena(t)
	m := abrirJuntar(t, c)
	m = passo(t, m, "enter")
	if len(c.resps) != 0 || m.armado != 0+1 {
		t.Fatalf("1º enter deveria só armar: resps=%v armado=%d", c.resps, m.armado)
	}
	if out := render(m, 96, 60); !strings.Contains(out, "Enter de novo para confirmar") {
		t.Fatalf("sem aviso de confirmação:\n%s", out)
	}
	for _, w := range []int{46, 56, 96} {
		cabeTudo(t, m, w, 44)
	}
	m = passo(t, m, "enter")
	if len(c.resps) != 1 || c.resps[0].resposta != "Liberar merge" {
		t.Fatalf("2º enter não respondeu: %v", c.resps)
	}
}

func TestOutraTeclaDesarma(t *testing.T) {
	c := novaCena(t)
	m := abrirJuntar(t, c)
	m = passo(t, m, "enter")
	m = passo(t, m, "down")
	m = passo(t, m, "up")
	if m.armado != 0 {
		t.Fatal("trocar a seleção deveria desarmar")
	}
	m = passo(t, m, "enter")
	m = passo(t, m, "esc")
	if m.armado != 0 || m.aberto == nil {
		t.Fatal("esc deveria só desarmar, sem fechar o detalhe")
	}
	if len(c.resps) != 0 {
		t.Fatalf("respondeu sem confirmar: %v", c.resps)
	}
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	if m.rascunho == nil || m.armado != 0 {
		t.Fatal("Segurar deveria abrir o texto em 1 passo")
	}
}

func TestNaoVouFazerConfirmaAntesDoTexto(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "env")
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	if m.rascunho != nil || m.armado == 0 {
		t.Fatal("1º enter em Não vou fazer deveria armar")
	}
	m = passo(t, m, "enter")
	if m.rascunho == nil {
		t.Fatal("2º enter deveria abrir o texto")
	}
}

func TestCliqueDuploConfirma(t *testing.T) {
	c := novaCena(t)
	m := abrirJuntar(t, c)
	x, y := onde(t, m, "Liberar merge", 20)
	m, _ = clique(t, m, x, y)
	if len(c.resps) != 0 || m.armado == 0 {
		t.Fatal("1º clique deveria armar")
	}
	x, y = onde(t, m, "Liberar merge", 20)
	var cmd tea.Cmd
	m, cmd = clique(t, m, x, y)
	if cmd == nil {
		t.Fatal("2º clique não respondeu")
	}
	cmd()
	if len(c.resps) != 1 {
		t.Fatalf("resps=%v", c.resps)
	}
}

// ---- C: números e foco ----

func TestNumeroEscolheOpcao(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	out := render(m, 96, 60)
	if !strings.Contains(out, "1 Aprovar") || !strings.Contains(out, "2 Segurar") {
		t.Fatalf("sem número nas opções:\n%s", out)
	}
	m = passo(t, m, "2")
	if m.rascunho == nil || m.opcao != 1 {
		t.Fatal("2 deveria abrir o texto da opção livre Segurar")
	}
	m = passo(t, m, "esc")
	m = passo(t, m, "1")
	if len(c.resps) != 1 || c.resps[0].resposta != "Aprovar" {
		t.Fatalf("1 deveria responder Aprovar: %v", c.resps)
	}
}

func TestNumeroArmaEConfirma(t *testing.T) {
	c := novaCena(t)
	m := abrirJuntar(t, c)
	m = passo(t, m, "1")
	if len(c.resps) != 0 || m.armado == 0 {
		t.Fatal("1 deveria armar Liberar merge")
	}
	m = passo(t, m, "1")
	if len(c.resps) != 1 {
		t.Fatalf("1 de novo deveria confirmar: %v", c.resps)
	}
}

func TestTabAlternaFoco(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m.altura = 30
	m = passo(t, m, "tab")
	if !m.focoTexto {
		t.Fatal("tab deveria focar o conteúdo")
	}
	if out := render(m, 96, 30); !strings.Contains(out, "tab volta às opções") {
		t.Fatalf("sem indicação de foco:\n%s", out)
	}
	m = passo(t, m, "down")
	if m.rolagem != 1 || m.opcao != 0 {
		t.Fatalf("↓ no conteúdo deveria rolar: rolagem=%d opcao=%d", m.rolagem, m.opcao)
	}
	m = passo(t, m, "tab")
	m = passo(t, m, "down")
	if m.opcao != 1 {
		t.Fatal("↓ nas opções deveria escolher")
	}
}

// ---- D: ir à aba e abrir ----

func comAba(e dados.Estado) dados.Estado {
	e.Tasks = append(e.Tasks, dados.Linha{Chave: "c-pl-1", Rotulo: "c-pl-1", Detalhe: "esperando", Extra: "aba-1"})
	return e
}

func TestIrAbaEAbrirNoFeed(t *testing.T) {
	r := &registro{}
	m := modelo(comAba(fixture()), r)
	m.largura, m.altura = 96, 40
	out := render(m, 96, 40)
	if !strings.Contains(out, "ir à aba →") || strings.Contains(out, "abrir ↗") {
		t.Fatalf("botões errados no decidir:\n%s", out)
	}
	m = aperta(t, m, "g")
	if len(r.abas) != 1 || r.abas[0] != "aba-1" {
		t.Fatalf("g: %v", r.abas)
	}
	x, y := onde(t, m, "ir à aba →", 0)
	m, _ = clique(t, m, x, y)
	if len(r.abas) != 2 || m.aberto != nil {
		t.Fatalf("clique no botão: abas=%v aberto=%v", r.abas, m.aberto != nil)
	}
	m = aperta(t, m, "j")
	if out := render(m, 96, 40); !strings.Contains(out, "abrir ↗") {
		t.Fatalf("juntar sem abrir ↗:\n%s", out)
	}
	x, y = onde(t, m, "abrir ↗", 0)
	m, _ = clique(t, m, x, y)
	if len(r.urls) != 1 || !strings.Contains(r.urls[0], "pull/77") {
		t.Fatalf("abrir ↗: %v", r.urls)
	}
	for _, w := range []int{46, 56, 96} {
		cabeTudo(t, m, w, 30)
	}
}

func TestSemAbaNemURLSemBotao(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m = aperta(t, m, "6")
	out := render(m, 96, 40)
	if strings.Contains(out, "ir à aba →") || strings.Contains(out, "abrir ↗") {
		t.Fatalf("máquina não tem aba nem link:\n%s", out)
	}
}

func TestIrAbaEAbrirNoDetalhe(t *testing.T) {
	c := novaCena(t)
	e := c.estado("juntar")
	e.Tasks = []dados.Linha{{Chave: filepath.Join(c.dir, "q"), Rotulo: "c-pl-17724", Detalhe: "esperando", Extra: "aba-j"}}
	m := c.modelo(e)
	m = passo(t, m, "enter")
	out := render(m, 96, 60)
	if !strings.Contains(out, "ir à aba →") || !strings.Contains(out, "abrir ↗") {
		t.Fatalf("detalhe sem botões:\n%s", out)
	}
	m = passo(t, m, "g")
	if len(c.abas) != 1 || c.abas[0] != "aba-j" {
		t.Fatalf("g no detalhe: %v", c.abas)
	}
	x, y := onde(t, m, "abrir ↗", 0)
	m, _ = clique(t, m, x, y)
	if len(c.urls) != 1 || !strings.Contains(c.urls[0], "pull/3589") {
		t.Fatalf("abrir ↗ no detalhe: %v", c.urls)
	}
	x, y = onde(t, m, "ir à aba →", 0)
	_, _ = clique(t, m, x, y)
	if len(c.abas) != 2 {
		t.Fatalf("clique ir à aba: %v", c.abas)
	}
}

func TestAjudaCobreTudo(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m = aperta(t, m, "?")
	out := render(m, 96, 60)
	for _, s := range []string{"filtra", "tab", "1-9", "ir à aba", "abrir ↗", "confirma"} {
		if !strings.Contains(out, s) {
			t.Errorf("ajuda sem %q", s)
		}
	}
}
