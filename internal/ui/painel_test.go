package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestBotoesAbreEFechaComB(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "3")
	m = tecla(m, "b")
	if !m.botoes || Gavetas[m.gaveta].ID != "prs" {
		t.Fatalf("b devia abrir o painel sem trocar de gaveta (aberto=%v, %s)", m.botoes, Gavetas[m.gaveta].ID)
	}
	if !strings.Contains(render(m, 140, 40), "/dia abrir") {
		t.Fatal("painel sem os botões")
	}
	if m = tecla(m, "b"); m.botoes {
		t.Fatal("b de novo devia fechar")
	}
	if strings.Contains(render(m, 140, 40), "/dia abrir") {
		t.Fatal("painel fechado ainda aparece")
	}
}

func TestBotoesEscFecha(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "b")
	n, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if n.(Model).botoes {
		t.Fatal("esc devia fechar o painel")
	}
}

func TestBotoesFicamComAsTeclas(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}), "b")
	m = tecla(m, "3")
	if Gavetas[m.gaveta].ID != "agora" || !m.botoes {
		t.Fatal("com o painel aberto, 3 não devia trocar de gaveta")
	}
	m = tecla(m, "j")
	if a, _ := m.selecionado(); a.comando != dados.ComandosHerdr[1].ID {
		t.Fatalf("j devia andar no painel: %+v", a)
	}
	m = tecla(m, "b")
	if m.sel != 0 {
		t.Fatalf("ao fechar, a seleção da gaveta volta: %d", m.sel)
	}
}

func TestBotoesEnterDuasVezes(t *testing.T) {
	var pedidos []string
	m := modelo(fixture(), &registro{})
	m.acoes.Disparar = func(id string) (dados.Disparo, error) {
		pedidos = append(pedidos, id)
		return dados.Disparo{ID: id, Aba: "w2:t9", Nome: "dia-1106-0900", Inicio: time.Now()}, nil
	}
	m = tecla(m, "b")
	n, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = n.(Model)
	if cmd != nil || m.armadoCmd != "dia-abrir" || len(pedidos) != 0 {
		t.Fatal("primeiro enter devia só armar")
	}
	n, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("segundo enter devia disparar")
	}
	n, _ = n.(Model).Update(cmd())
	m = n.(Model)
	if len(pedidos) != 1 || pedidos[0] != "dia-abrir" || !m.botoes {
		t.Fatalf("pedidos %v, aberto %v", pedidos, m.botoes)
	}
	m = tecla(m, "g")
	if !m.botoes {
		t.Fatal("g não fecha o painel")
	}
}

func TestConfiguracoesSoAjustes(t *testing.T) {
	m, _, _ := modeloAjustes(t, nil)
	p := render(m, 110, 50)
	for _, s := range []string{"/dia abrir", "/saude-multas", "reconectar MCPs", "atualizar app"} {
		if strings.Contains(p, s) {
			t.Fatalf("configurações com o botão %q:\n%s", s, p)
		}
	}
	if !strings.Contains(p, "FANTASMAS_TEMA") || !strings.Contains(p, "b botões") {
		t.Fatalf("configurações sem ajustes ou sem a dica:\n%s", p)
	}
}

func TestApelidoComandos(t *testing.T) {
	if !GavetaValida("comandos") || Gavetas[indiceGaveta("comandos")].ID != "configuracoes" {
		t.Fatal("comandos devia levar a configurações")
	}
}

func TestBotoesLarguraEstreitaOcupaTudo(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComBotoes()
	for _, w := range []int{60, 79} {
		ls := strings.Split(m.Render(w, 30), "\n")
		for _, l := range ls[1 : len(ls)-1] {
			if p := ansi.Strip(l); strings.Contains(p, "botões") && !strings.HasPrefix(p, "▏") {
				t.Fatalf("largura %d: painel devia começar na coluna 0: %q", w, p)
			}
		}
		if larguraPainel(w) != w {
			t.Fatalf("largura %d: painel com %d", w, larguraPainel(w))
		}
	}
	if larguraPainel(140) != LarguraBotoes {
		t.Fatal("em 140 colunas o painel tem largura fixa")
	}
	if p := render(m, 140, 30); !strings.Contains(p, "agora") || !strings.Contains(p, "▏ botões") {
		t.Fatalf("em 140 colunas a gaveta segue à esquerda:\n%s", p)
	}
}
