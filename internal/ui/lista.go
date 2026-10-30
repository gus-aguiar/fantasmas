package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// alvo é o que as teclas fazem no item selecionado.
type alvo struct {
	url, aba, pergunta, relatorio string
	arquivar                      string
	retomar                       string
	antigas                       bool
	acao                          string
	pipe                          int
	missao                        string
}

// linha do corpo; item é o índice do alvo dono da linha, ou -1.
type linha struct {
	desenha func(bg lipgloss.Color) string
	item    int
}

func st(bg, fg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(bg).Foreground(fg)
}

// encaixa corta por largura de célula e completa com o fundo.
func encaixa(s string, largura int, bg lipgloss.Color) string {
	if largura <= 0 {
		return ""
	}
	if lipgloss.Width(s) > largura {
		s = ansi.Truncate(s, largura, "…")
	}
	return s + espacos(largura-lipgloss.Width(s), bg)
}

type corpo struct {
	linhas     []linha
	alvos      []alvo
	largura    int
	filtro     string
	vis, total int

	antigasAbertas, rascunhos bool
}

func (c *corpo) casa(partes ...string) bool {
	return c.filtro == "" || strings.Contains(normaliza(strings.Join(partes, " ")), c.filtro)
}

// passa conta o item no total e diz se ele bate com o filtro.
func (c *corpo) passa(partes ...string) bool {
	c.total++
	if !c.casa(partes...) {
		return false
	}
	c.vis++
	return true
}

func (c *corpo) texto(f func(bg lipgloss.Color) string) {
	c.linhas = append(c.linhas, linha{desenha: f, item: -1})
}

func (c *corpo) branco() { c.texto(func(lipgloss.Color) string { return "" }) }

func (c *corpo) secao(titulo, extra string) {
	c.texto(func(bg lipgloss.Color) string {
		s := st(bg, T.Acao).Render("› ") + st(bg, T.Sub).Bold(true).Render(titulo)
		if extra != "" {
			s += st(bg, T.Dim).Render("  " + extra)
		}
		return s
	})
}

func (c *corpo) vazio(msg string) {
	w := c.largura
	c.texto(func(bg lipgloss.Color) string {
		m := Corte(msg, w)
		esq := (w - lipgloss.Width(m)) / 2
		return espacos(esq, bg) + st(bg, T.Dim).Render(m)
	})
}

func (c *corpo) item(a alvo, fs ...func(bg lipgloss.Color) string) {
	i := len(c.alvos)
	c.alvos = append(c.alvos, a)
	for _, f := range fs {
		if f == nil {
			continue
		}
		c.linhas = append(c.linhas, linha{desenha: f, item: i})
	}
}

// linhaLista é o porte de gavetas/lista.tsx: glifo, nome, estado e detalhe dim.
func (c *corpo) linhaLista(a alvo, glifo string, cor lipgloss.Color, nome, estado string, corEstado lipgloss.Color, detalhe string) {
	estreito := c.largura < 70
	wNome, wEstado := 28, 14
	if estreito {
		wNome, wEstado = 20, 11
	}
	wDet := c.largura - 4 - wNome - wEstado
	if !c.passa(nome, estado, detalhe, a.url) {
		return
	}
	c.item(a, func(bg lipgloss.Color) string {
		s := st(bg, cor).Render(glifo+" ") + st(bg, T.Texto).Render(Col(nome, wNome)) + st(bg, corEstado).Render(Col(estado, wEstado))
		if !estreito && wDet > 6 {
			s += st(bg, T.Dim).Render(Corte(detalhe, wDet))
		}
		return s
	})
}

func juntar(partes []string) string { return strings.Join(partes, " · ") }
