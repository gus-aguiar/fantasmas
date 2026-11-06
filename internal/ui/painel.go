package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// LarguraBotoes é a largura do painel de botões; abaixo de LarguraBotoesCheia colunas ele ocupa a tela.
const (
	LarguraBotoes      = 34
	LarguraBotoesCheia = 80
	DicaBotoes         = "b botões"
)

func larguraPainel(largura int) int {
	if largura < LarguraBotoesCheia {
		return largura
	}
	return LarguraBotoes
}

// ComBotoes abre o painel de botões (modo quadro).
func (m Model) ComBotoes() Model { return m.abrirBotoes() }

func (m Model) abrirBotoes() Model {
	m.selGaveta, m.sel, m.botoes, m.armadoCmd = m.sel, 0, true, ""
	return m
}

func (m Model) fecharBotoes() Model {
	m.botoes, m.sel, m.armadoCmd = false, m.selGaveta, ""
	return m
}

// teclaBotoes: com o painel aberto, as teclas ficam com ele.
func (m Model) teclaBotoes(k string) (tea.Model, tea.Cmd) {
	armadoCmd := m.armadoCmd
	m.armadoCmd = ""
	switch k {
	case "b", "esc":
		return m.fecharBotoes(), nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		return m.mover(-1), nil
	case "down", "j":
		return m.mover(1), nil
	case "g":
		return m.irAba(), nil
	case "enter":
		m.armadoCmd = armadoCmd
		return m.enterCmd()
	}
	return m, nil
}

func (m Model) corpoBotoes() *corpo { return m.corpoBotoesEm(larguraPainel(m.larguraEfetiva())) }

func (m Model) corpoBotoesEm(pw int) *corpo {
	c := &corpo{largura: pw - 1}
	painelBotoes(c, m.estado, estadoComandos{disparos: m.disparos, armado: m.armadoCmd, disparando: m.disparando,
		agora: m.agora(), pedindoDD: m.pedindoDD, atualizando: m.atualizando, armadoAtu: m.armadoAtu})
	return c
}

// sobreporBotoes desenha o painel à direita, por cima das linhas da gaveta.
func (m Model) sobreporBotoes(fundo []string, largura int) []string {
	pw := larguraPainel(largura)
	c := m.corpoBotoesEm(pw)
	borda := st(T.Surface, T.Acao).Render("▏")
	titulo := borda + encaixa(st(T.Surface, T.Sub).Bold(true).Render(" botões")+st(T.Surface, T.Dim).Render("  enter 2× · g aba · esc"), pw-1, T.Surface)
	ps := []string{titulo, borda + espacos(pw-1, T.Surface)}
	for _, l := range c.linhas {
		bg := T.Surface
		if l.item >= 0 && l.item == m.sel {
			bg = T.Sel
		}
		ps = append(ps, borda+encaixa(espacos(1, bg)+l.desenha(bg), pw-1, bg))
	}
	out := make([]string, len(fundo))
	for i, l := range fundo {
		p := borda + espacos(pw-1, T.Surface)
		if i < len(ps) {
			p = ps[i]
		}
		esq := ""
		if largura > pw {
			esq = encaixa(ansi.Truncate(l, largura-pw, ""), largura-pw, T.Base)
		}
		out[i] = esq + p
	}
	return out
}
