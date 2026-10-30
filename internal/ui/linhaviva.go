package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	wNomeViva  = 16
	wEtapaViva = 11
	wBarraViva = 8
	wHoraViva  = 6
)

func barraCtx(pct int) string { return barraCtxEm(pct, T.Base) }

func barraCtxEm(pct int, bg lipgloss.Color) string {
	if pct < 0 {
		return st(bg, T.Dim).Render("—")
	}
	pct = min(pct, 100)
	cheios := min(3, int(math.Round(float64(pct)/33.4)))
	corPct := T.Sub
	if pct >= 80 {
		corPct = T.Espera
	}
	return st(bg, corPct).Render(fmt.Sprintf("%d%%", pct)) +
		st(bg, T.BarraCheia).Render(strings.Repeat("▪", cheios)) +
		st(bg, T.BarraVazia).Render(strings.Repeat("▫", 3-cheios))
}

func glifoLinha(e dados.EstadoLinha) (string, lipgloss.Color) {
	switch e {
	case dados.EsperaVoce:
		return "▲", T.Espera
	case dados.Trabalhando:
		return "●", T.Andamento
	case dados.Parado:
		return "◌", T.Dim
	}
	return "✓", T.Ok
}

func corAgora(e dados.EstadoLinha) lipgloss.Color {
	switch e {
	case dados.EsperaVoce:
		return T.Espera
	case dados.Trabalhando:
		return T.Texto
	case dados.Parado:
		return T.Dim
	}
	return T.Sub
}

func renderLinhaViva(l dados.LinhaViva, largura int, bg lipgloss.Color) string {
	if largura <= 0 {
		return ""
	}
	glifo, cor := glifoLinha(l.Estado)
	wNome := wNomeViva
	if largura < 46 {
		wNome = 12
	}
	if largura < 30 {
		wNome = max(4, largura/2-2)
	}
	s := st(bg, cor).Render(glifo+" ") + st(bg, T.Texto).Bold(true).Render(Col(l.Nome, wNome)) + espacos(1, bg)
	resto := largura - 2 - wNome - 1
	largo := largura >= 56
	if largo {
		etapa := l.Etapa
		if etapa == "" {
			etapa = "—"
		}
		s += st(T.Chip, T.ChipTexto).Render(" "+Col(etapa, wEtapaViva)+" ") + espacos(1, bg)
		resto -= wEtapaViva + 3
	}
	if largura >= 30 {
		b := barraCtxEm(l.CtxPct, bg)
		s += b + espacos(wBarraViva-lipgloss.Width(b), bg)
		resto -= wBarraViva
	}
	if largo {
		resto -= wHoraViva
	}
	if resto > 0 {
		s += st(bg, corAgora(l.Estado)).Render(Col(l.Agora, resto))
	}
	if largo {
		hora := "—"
		if !l.Hora.IsZero() {
			hora = l.Hora.In(time.Local).Format("15:04")
		}
		s += st(bg, T.Dim).Render(fmt.Sprintf("%*s", wHoraViva, hora))
	}
	if lipgloss.Width(s) > largura {
		s = ansi.Truncate(s, largura, "")
	}
	return s
}

// cabecalhoViva rotula as colunas da linha viva, alinhado com renderLinhaViva.
func cabecalhoViva(largura int, bg lipgloss.Color) string {
	if largura < 56 {
		return ""
	}
	resto := largura - 2 - wNomeViva - 1 - (wEtapaViva + 3) - wBarraViva - wHoraViva
	s := espacos(2, bg) + Col("fantasminha", wNomeViva+1) + Col("etapa", wEtapaViva+3) + Col("ctx", wBarraViva) + Col("agora", max(0, resto)) + fmt.Sprintf("%*s", wHoraViva, "hora")
	return st(bg, T.Dim).Render(s)
}

func alvoViva(l dados.LinhaViva) alvo {
	a := alvo{aba: l.Aba}
	if l.Estado == dados.Terminou {
		a.relatorio = l.Pasta
	}
	return a
}

func (c *corpo) linhaViva(l dados.LinhaViva) {
	w := c.largura - 2
	c.item(alvoViva(l), func(bg lipgloss.Color) string { return renderLinhaViva(l, w, bg) })
}

func linhasPorEstado(ls []dados.LinhaViva) []dados.LinhaViva {
	out := append([]dados.LinhaViva(nil), ls...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Estado < out[j].Estado })
	return out
}
