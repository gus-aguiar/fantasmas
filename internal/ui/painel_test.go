package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func manda(m Model, msg tea.Msg) Model {
	n, _ := m.Update(msg)
	return n.(Model)
}

func digitaPaleta(m Model, s string) Model {
	for _, r := range s {
		m = manda(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestPaletaAbreComCtrlPEB(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "3")
	m = manda(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	if !m.botoes || Gavetas[m.gaveta].ID != "prs" {
		t.Fatalf("ctrl+p devia abrir a paleta sem trocar de gaveta (aberta=%v, %s)", m.botoes, Gavetas[m.gaveta].ID)
	}
	p := render(m, 140, 40)
	for _, s := range []string{"Comandos", "Abrir o dia", "Fechar o dia", "Saúde de multas", "esc fecha"} {
		if !strings.Contains(p, s) {
			t.Fatalf("paleta sem %q:\n%s", s, p)
		}
	}
	for _, s := range []string{"DayDreaming", "Reconectar MCPs", "Atualizar app", "SISTEMA", "buscar comando"} {
		if strings.Contains(p, s) {
			t.Fatalf("paleta ainda com %q:\n%s", s, p)
		}
	}
	if n := len(m.alvos()); n != 3 {
		t.Fatalf("paleta devia ter 3 comandos: %d", n)
	}
	if m = manda(m, tea.KeyMsg{Type: tea.KeyCtrlP}); m.botoes {
		t.Fatal("ctrl+p de novo devia fechar")
	}
	if m = tecla(m, "b"); !m.botoes {
		t.Fatal("b devia abrir a paleta")
	}
	if m = tecla(m, "b"); m.botoes {
		t.Fatal("b com a busca vazia devia fechar")
	}
	if strings.Contains(render(m, 140, 40), "Abrir o dia") {
		t.Fatal("paleta fechada ainda aparece")
	}
}

func TestPaletaCentralizadaComFundoEsmaecido(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComBotoes()
	ls := strings.Split(m.Render(140, 40), "\n")
	topo := -1
	for i, l := range ls {
		if strings.Contains(ansi.Strip(l), "Comandos") && topo < 0 {
			topo = i - 3
		}
		if lipgloss.Width(l) > 140 {
			t.Fatalf("linha %d estoura", i)
		}
	}
	if topo < 0 {
		t.Fatal("sem a borda com o título")
	}
	x := strings.Index(ansi.Strip(ls[topo]), "╭")
	if esq, dir := x, 140-x-larguraPaleta(140); esq-dir > 1 || dir-esq > 1 {
		t.Fatalf("caixa fora do centro: %d × %d", esq, dir)
	}
}

func TestPaletaSemBuscaENumeros(t *testing.T) {
	var pedidos []string
	m := tecla(modelo(fixture(), &registro{}), "b")
	m.acoes.Disparar = func(id string) (dados.Disparo, error) {
		pedidos = append(pedidos, id)
		return dados.Disparo{ID: id, Aba: "w2:t9", Inicio: time.Now()}, nil
	}
	if m.buscaCmd != nil {
		t.Fatal("com até 5 comandos a paleta não tem busca")
	}
	m = digitaPaleta(m, "multas")
	if len(m.alvos()) != 3 || m.textoBusca() != "" {
		t.Fatal("letras não deviam filtrar")
	}
	n, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = n.(Model)
	if cmd != nil || m.armadoCmd != "saude-multas" || len(pedidos) != 0 {
		t.Fatalf("primeiro 3 devia só armar: %q", m.armadoCmd)
	}
	if !strings.Contains(render(m, 140, 40), "enter de novo confirma") {
		t.Fatal("item armado sem aviso")
	}
	n, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if cmd == nil {
		t.Fatal("segundo 3 devia disparar")
	}
	n.(Model).Update(cmd())
	if len(pedidos) != 1 || pedidos[0] != "saude-multas" {
		t.Fatalf("pedidos %v", pedidos)
	}
	m = tecla(tecla(m, "1"), "2")
	if m.armadoCmd != "dia-fechar" {
		t.Fatalf("2 depois de 1 devia armar o 2, não disparar o 1: %q", m.armadoCmd)
	}
}

func TestPaletaMostraOQueCadaUmFazSemComando(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComBotoes()
	p := render(m, 140, 40)
	if !strings.Contains(p, "começa o dia") || !strings.Contains(p, "fecha o dia e faz") {
		t.Fatalf("cada comando devia mostrar o que faz:\n%s", p)
	}
	if strings.Contains(p, "/dia abrir") || strings.Contains(p, "ᗣ") {
		t.Fatalf("paleta não mostra o comando nem fantasma:\n%s", p)
	}
}

func TestPaletaEscFecha(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "b")
	m = digitaPaleta(m, "dia")
	if m = manda(m, tea.KeyMsg{Type: tea.KeyEsc}); m.botoes || m.textoBusca() != "" {
		t.Fatal("esc devia fechar a paleta e limpar a busca")
	}
}

func TestPaletaFicaComAsTeclas(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "b")
	m = tecla(m, "j")
	if a, _ := m.selecionado(); a.comando != dados.ComandosHerdr[1].ID {
		t.Fatalf("j com busca vazia devia andar na paleta: %+v", a)
	}
	m = tecla(m, "3")
	if Gavetas[m.gaveta].ID != "agora" || !m.botoes {
		t.Fatal("com a paleta aberta, 3 não devia trocar de gaveta")
	}
	m = manda(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.sel != 0 {
		t.Fatalf("ao fechar, a seleção da gaveta volta: %d", m.sel)
	}
}

func TestPaletaEnterDuasVezes(t *testing.T) {
	var pedidos []string
	m := modelo(fixture(), &registro{})
	m.acoes.Disparar = func(id string) (dados.Disparo, error) {
		pedidos = append(pedidos, id)
		return dados.Disparo{ID: id, Aba: "w2:t9", Nome: "dia-1106-0900", Inicio: time.Now()}, nil
	}
	m = manda(m, tea.KeyMsg{Type: tea.KeyCtrlP})
	n, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = n.(Model)
	if cmd != nil || m.armadoCmd != "dia-abrir" || len(pedidos) != 0 {
		t.Fatal("primeiro enter devia só armar")
	}
	if !strings.Contains(render(m, 140, 40), "enter de novo confirma") {
		t.Fatal("item armado sem aviso")
	}
	n, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("segundo enter devia disparar")
	}
	n, _ = n.(Model).Update(cmd())
	m = n.(Model)
	if len(pedidos) != 1 || pedidos[0] != "dia-abrir" || !m.botoes {
		t.Fatalf("pedidos %v, aberta %v", pedidos, m.botoes)
	}
	if !strings.Contains(render(m, 140, 40), "rodando há") {
		t.Fatal("estado do disparo não aparece à direita")
	}
	if a, _ := m.selecionado(); a.aba != "w2:t9" {
		t.Fatalf("g não iria à aba do disparo: %+v", a)
	}
}

func TestConfiguracoesTemAcoes(t *testing.T) {
	m, _, _ := modeloAjustes(t, nil)
	p := render(m, 110, 60)
	for _, s := range []string{"FANTASMAS_TEMA", "AÇÕES", "Reconectar MCPs", "Atualizar app"} {
		if !strings.Contains(p, s) {
			t.Fatalf("configurações sem %q:\n%s", s, p)
		}
	}
	if strings.Contains(p, "/dia abrir") || strings.Contains(p, "DayDreaming") {
		t.Fatalf("configurações com comando da paleta:\n%s", p)
	}
}

func TestApelidoComandos(t *testing.T) {
	if !GavetaValida("comandos") || Gavetas[indiceGaveta("comandos")].ID != "configuracoes" {
		t.Fatal("comandos devia levar a configurações")
	}
}

func TestPaletaJanelaEstreita(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComBotoes()
	for _, w := range []int{40, 60, 79} {
		for i, l := range strings.Split(m.Render(w, 30), "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("largura %d linha %d estoura", w, i)
			}
		}
	}
	if larguraPaleta(140) != LarguraPaleta || larguraPaleta(50) != 46 {
		t.Fatal("largura da paleta")
	}
}

func TestCabecalhosDeSecao(t *testing.T) {
	e := fixture()
	e.Linhas = linhasEx()
	p := render(modelo(e, &registro{}), 140, 60)
	for _, s := range []string{"ESPERANDO VOCÊ ", "TRABALHANDO ", "PARADOS "} {
		if !strings.Contains(p, s) {
			t.Fatalf("sem o cabeçalho %q:\n%s", s, p)
		}
	}
	if strings.Contains(p, "› precisa de você") || strings.Contains(p, "(0)") {
		t.Fatalf("cabeçalho antigo:\n%s", p)
	}
}

func TestColunasMetaADireitaETruncada(t *testing.T) {
	s := ansi.Strip(colunas(T.Base, 40, "●", T.Ok, "um título bem comprido que não cabe na linha", "c-pl-1 · 55% · há 2h"))
	if lipgloss.Width(s) != 40 || !strings.HasPrefix(s, "● um título") || !strings.Contains(s, "…") {
		t.Fatalf("%q", s)
	}
	curto := ansi.Strip(colunas(T.Base, 60, "▲", T.Espera, "curto", "há 2h"))
	if !strings.HasSuffix(curto, "há 2h") || lipgloss.Width(curto) != 60 {
		t.Fatalf("metadado devia encostar à direita: %q", curto)
	}
}

func TestGengarParadoSemAtividade(t *testing.T) {
	e := fixture()
	e.Tasks, e.Linhas = nil, []dados.LinhaViva{{Nome: "c-x", Estado: dados.Parado}}
	m := modelo(e, &registro{})
	q := m.quadro
	if m = manda(m, tickQuadroMsg{}); m.quadro != q {
		t.Fatal("sem ninguém trabalhando o gengar devia ficar parado")
	}
	e.Linhas[0].Estado = dados.Trabalhando
	m = m.ComEstado(e)
	if m = manda(m, tickQuadroMsg{}); m.quadro != q+1 {
		t.Fatal("com alguém trabalhando o gengar devia animar")
	}
}

func TestTopoGengarCheio(t *testing.T) {
	m := modelo(fixture(), &registro{})
	sp := spriteFn(140)
	if sp == nil {
		t.Skip("sem sprite")
	}
	ls, _ := m.linhasTopo(140, 42)
	if len(ls) < sp.Linhas || sp.Linhas <= 4 {
		t.Fatalf("gengar devia vir em tamanho cheio: sprite %d linhas, topo %d", sp.Linhas, len(ls))
	}
	p := render(m, 140, 42)
	for _, s := range []string{"fantasmas  ↻ atualizado", "precisa de você", "1 Agora · 2 Missões"} {
		if !strings.Contains(p, s) {
			t.Fatalf("topo sem %q:\n%s", s, p)
		}
	}
	if ls, _ := m.linhasTopo(140, AlturaSemMascote-1); len(ls) != 1 {
		t.Fatal("abaixo de 30 linhas o gengar some")
	}
}
