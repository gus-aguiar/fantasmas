package ui

import (
	"fmt"
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
	pipe                          int
	missao                        string
	ajuste                        string
	// decisão vetável: sidequest, etapa e o rótulo ("plano")
	vetoID            int
	vetoEtapa, vetoDe string
	// decisao marca a linha de decisão (vetável ou não), para o v responder mesmo sem etapa
	decisao bool
	// alterna abre ou recolhe uma parte da gaveta (rascunhos, maquina); comando é um disparo da gaveta comandos
	alterna, comando string
	// dica troca o atalho padrão mostrado na linha selecionada (ver dicaItem)
	dica string
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

	// desc é a frase curta da gaveta: vai na primeira seção, depois do título (ver secaoN).
	desc string
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

func (c *corpo) secao(titulo, extra string) { c.secaoN(titulo, -1, extra) }

// secaoN é o cabeçalho de seção esmaecido em caixa alta com o contador (zero não aparece), seguido
// de uma linha em branco. A primeira seção da gaveta leva também a descrição dela.
func (c *corpo) secaoN(titulo string, n int, extra string) {
	titulo = strings.ReplaceAll(strings.ToUpper(titulo), "PRS", "PRs")
	if n > 0 {
		titulo += fmt.Sprintf(" %d", n)
	}
	desc := c.desc
	c.desc = ""
	c.texto(func(bg lipgloss.Color) string {
		s := st(bg, T.Secundario()).Bold(true).Render(titulo)
		if desc != "" {
			s += st(bg, T.Dim).Render("  " + desc)
		}
		if extra != "" {
			sep := "  "
			if desc != "" {
				sep = " · "
			}
			s += st(bg, T.Dim).Render(sep + extra)
		}
		if w := c.largura; w > 0 && lipgloss.Width(s) > w {
			s = ansi.Truncate(s, w, "…")
		}
		return s
	})
	c.branco()
}

// vazio é o estado vazio neutro (○); vazioIcone escolhe o ícone e a cor.
func (c *corpo) vazio(msg string) { c.vazioIcone("○", T.Dim, msg) }

// vazioIcone desenha o estado vazio como uma caixa de borda arredondada na largura toda, com 2 linhas
// de folga em cima e embaixo, o texto centrado e o ícone antes da primeira linha.
func (c *corpo) vazioIcone(icone string, cor lipgloss.Color, linhas ...string) {
	w := c.largura
	caixa := max(4, w-2) // largura toda, com 1 coluna de margem de cada lado
	dentro := caixa - 2
	esq := max(0, (w-caixa)/2)
	borda := func(bg lipgloss.Color, a, b string) string {
		return espacos(esq, bg) + st(bg, T.Borda).Render(a+strings.Repeat("─", dentro)+b)
	}
	meio := func(bg lipgloss.Color, s string) string {
		s = Corte(s, max(1, dentro-2))
		e := (dentro - lipgloss.Width(s)) / 2
		return espacos(esq, bg) + st(bg, T.Borda).Render("│") + espacos(e, bg) + s +
			espacos(dentro-e-lipgloss.Width(s), bg) + st(bg, T.Borda).Render("│")
	}
	c.texto(func(bg lipgloss.Color) string { return borda(bg, "╭", "╮") })
	c.texto(func(bg lipgloss.Color) string { return meio(bg, "") })
	c.texto(func(bg lipgloss.Color) string { return meio(bg, "") })
	for i, l := range linhas {
		l, i := l, i
		c.texto(func(bg lipgloss.Color) string {
			if i == 0 {
				return meio(bg, st(bg, cor).Render(icone)+espacos(2, bg)+st(bg, T.Sub).Render(Corte(l, max(1, dentro-6))))
			}
			return meio(bg, st(bg, T.Dim).Render(Corte(l, max(1, dentro-4))))
		})
	}
	c.texto(func(bg lipgloss.Color) string { return meio(bg, "") })
	c.texto(func(bg lipgloss.Color) string { return meio(bg, "") })
	c.texto(func(bg lipgloss.Color) string { return borda(bg, "╰", "╯") })
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

// colunas é a linha de lista: glifo colorido, título na cor normal e metadados em cinza alinhados
// à direita; o que não cabe é cortado com "…".
func colunas(bg lipgloss.Color, largura int, glifo string, cor lipgloss.Color, titulo, meta string) string {
	esq := st(bg, cor).Render(glifo + " ")
	resto := largura - lipgloss.Width(esq)
	if resto <= 0 {
		return ansi.Truncate(esq, largura, "")
	}
	wMeta := 0
	if meta != "" && resto > 20 {
		wMeta = min(lipgloss.Width(meta), resto*45/100)
	}
	wTit := resto - wMeta
	if wMeta > 0 {
		wTit -= 2
	}
	tit := Corte(titulo, wTit)
	s := esq + st(bg, T.Texto).Render(tit)
	if wMeta > 0 {
		m := Corte(meta, wMeta)
		s += espacos(resto-lipgloss.Width(tit)-lipgloss.Width(m), bg) + st(bg, T.Secundario()).Render(m)
	}
	return s
}

// linhaLista desenha um item em colunas: glifo, nome, e estado · detalhe à direita.
func (c *corpo) linhaLista(a alvo, glifo string, cor lipgloss.Color, nome, estado string, _ lipgloss.Color, detalhe string) {
	if !c.passa(nome, estado, detalhe, a.url) {
		return
	}
	var meta []string
	for _, p := range []string{estado, detalhe} {
		if p != "" {
			meta = append(meta, p)
		}
	}
	w := c.largura - 2
	c.item(a, func(bg lipgloss.Color) string { return colunas(bg, w, glifo, cor, nome, juntar(meta)) })
}

func juntar(partes []string) string { return strings.Join(partes, " · ") }
