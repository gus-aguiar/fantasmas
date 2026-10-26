package ui

import (
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func Faixa(largura int) string {
	switch {
	case largura >= 90:
		return "largo"
	case largura >= 70:
		return "medio"
	case largura >= 46:
		return "estreito"
	}
	return "minimo"
}

type Gaveta struct{ ID, Tecla, Nome string }

var Gavetas = []Gaveta{
	{"agora", "1", "agora"}, {"crew", "2", "fantasminhas"}, {"prs", "3", "PRs"}, {"ci", "4", "CI"},
	{"ambiente", "5", "ambiente"}, {"maquina", "6", "máquina"}, {"mais", "7", "mais"},
}

func fundo(c lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Background(c) }

func espacos(n int, c lipgloss.Color) string {
	if n <= 0 {
		return ""
	}
	return fundo(c).Render(strings.Repeat(" ", n))
}

// linhaFundo corta s à largura e completa com fundo mantle; alinha ao centro ou à esquerda.
func linhaFundo(s string, largura int, centro bool) string {
	if lipgloss.Width(s) > largura {
		s = ansi.Truncate(s, largura, "")
	}
	sobra := largura - lipgloss.Width(s)
	esq := 0
	if centro {
		esq = sobra / 2
	}
	return espacos(esq, T.Mantle) + s + espacos(sobra-esq, T.Mantle)
}

func hora(e dados.Estado) string {
	if e.At.IsZero() {
		return "—"
	}
	return e.At.Format("15:04")
}

func Topo(e dados.Estado, largura, quadro int, coletando bool, status string) string {
	ls, _ := topoArea(e, largura, quadro, coletando, status, false)
	return strings.Join(ls, "\n")
}

// topoArea desenha o topo; com img, reserva em branco a área da imagem e devolve onde ela fica.
func topoArea(e dados.Estado, largura, quadro int, coletando bool, status string, img bool) ([]string, areaImg) {
	if largura <= 0 {
		return []string{""}, areaImg{}
	}
	f := Faixa(largura)
	st := func(c lipgloss.Color) lipgloss.Style { return fundo(T.Mantle).Foreground(c) }
	resumo := dados.Resumo(e)
	if f == "minimo" {
		s := st(T.Acao).Bold(true).Render("fantasmas") + st(T.Sub).Render("  "+Corte(resumo, largura))
		return []string{linhaFundo(s, largura, false)}, areaImg{}
	}

	spin := []string{"◐", "◓", "◑", "◒"}
	glifo := "↻"
	quando := "atualizado " + hora(e)
	if coletando {
		glifo = spin[((quadro%4)+4)%4]
		quando = "atualizando…"
	}
	livre, detalhe := true, ""
	for _, l := range e.Maquina {
		if l.Chave == "trava" {
			livre, detalhe = l.Extra == "ok", l.Detalhe
		}
	}
	dica := status
	cor := T.Texto
	if dica == "" {
		dica, cor = "r atualiza · p painel · ? ajuda · q sai", T.Dim
	}

	amplo := f == "largo" || f == "medio"
	gap := 4
	if !amplo {
		gap = 2
	}
	arte, artCol, artLin := QuadroMascote, MascoteColunas, MascoteLinhas
	if img {
		artCol, artLin = tamanhoImagem(largura)
		c, l := artCol, artLin
		arte = func(int, lipgloss.Color) []string { return centraV(nil, c, l) }
	} else if sp := spriteFn(largura); sp != nil {
		arte = func(i int, _ lipgloss.Color) []string { return sp.Quadro(i) }
		artCol, artLin = sp.Colunas, sp.Linhas
	}
	maxCard := max(1, largura-2-artCol-gap)

	l1 := st(T.Acao).Bold(true).Render("fantasmas") + st(T.Dim).Render("  "+glifo+" "+quando)
	l2 := st(T.Texto).Render(resumo)
	var l3 string
	if livre {
		l3 = st(T.Ok).Render("● ") + st(T.Sub).Render("fila de pesados livre")
	} else {
		l3 = st(T.Espera).Render("◆ ") + st(T.Sub).Render("pesados: "+detalhe)
	}
	l4 := st(cor).Render(dica)
	bruto := []string{l1, "", l2, l3, "", l4}
	if artLin < 6 {
		bruto = []string{l1, l2, l3, l4}
	}
	cardW := 0
	for i, l := range bruto {
		if lipgloss.Width(l) > maxCard {
			l = ansi.Truncate(l, maxCard, "…")
			bruto[i] = l
		}
		if w := lipgloss.Width(l); w > cardW {
			cardW = w
		}
	}

	a := arte(quadro, T.Mantle)
	g := espacos(gap, T.Mantle)
	alt := max(artLin, len(bruto))
	a = centraV(a, artCol, alt)
	cartao := centraV(bruto, 0, alt)
	linhas := make([]string, alt)
	for i := range linhas {
		linhas[i] = linhaFundo(espacos(1, T.Mantle)+a[i]+g+linhaFundo(cartao[i], cardW, false), largura, amplo)
	}
	area := areaImg{y: (alt - artLin) / 2, c: artCol, r: artLin, ok: img}
	if sobra := largura - (1 + artCol + gap + cardW); amplo && sobra > 0 {
		area.x = sobra / 2
	}
	area.x++
	return linhas, area
}

// TopoCompacto é a linha única do topo com o detalhe aberto: voltar à esquerda, estado à direita.
func TopoCompacto(e dados.Estado, largura, quadro int, coletando bool, status string) string {
	st := func(c lipgloss.Color) lipgloss.Style { return fundo(T.Mantle).Foreground(c) }
	glifo, quando := "↻", "atualizado "+hora(e)
	if coletando {
		glifo, quando = []string{"◐", "◓", "◑", "◒"}[((quadro%4)+4)%4], "atualizando…"
	}
	esq := st(T.Acao).Bold(true).Render(" "+Voltar) + st(T.Dim).Render("   ")
	dir := st(T.Dim).Render(glifo+" "+quando) + espacos(1, T.Mantle)
	meio := st(T.Sub).Render("fantasmas")
	if status != "" {
		meio = st(T.Texto).Render(status)
	}
	w := largura - lipgloss.Width(esq) - lipgloss.Width(dir)
	if w < 4 {
		return linhaFundo(esq, largura, false)
	}
	return esq + linhaFundo(ansi.Truncate(meio, w, "…"), w, false) + dir
}

const Voltar = "‹ voltar"

// centraV completa ls com linhas de fundo (largura w) em cima e embaixo até a altura alt.
func centraV(ls []string, w, alt int) []string {
	cima := (alt - len(ls)) / 2
	out := make([]string, 0, alt)
	for i := 0; i < cima; i++ {
		out = append(out, espacos(w, T.Mantle))
	}
	out = append(out, ls...)
	for len(out) < alt {
		out = append(out, espacos(w, T.Mantle))
	}
	return out
}

func rotulo(g Gaveta, ativa bool, nivel int) string {
	switch {
	case nivel == 0:
		return g.Tecla + " " + g.Nome
	case nivel == 1, nivel == 2 && ativa:
		return g.Nome
	}
	return g.Tecla
}

// Rodape é sempre uma linha: usa o rótulo mais rico que cabe.
func Rodape(ativa string, largura int) string {
	s, _ := rodapeAbas(ativa, largura)
	return s
}

// rodapeAbas devolve o rodapé e o intervalo [x0, x1) de cada aba, para o clique.
func rodapeAbas(ativa string, largura int) (string, [][2]int) {
	if largura <= 0 {
		return "", nil
	}
	tentativas := [][2]int{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {3, 0}}
	var s string
	var abas [][2]int
	for _, t := range tentativas {
		nivel, gap := t[0], t[1]
		var b strings.Builder
		abas = abas[:0]
		x := 0
		for _, g := range Gavetas {
			on := g.ID == ativa
			var r string
			if on {
				r = fundo(T.Acao).Foreground(T.Mantle).Bold(true).Render(" " + rotulo(g, on, nivel) + " ")
			} else {
				r = fundo(T.Mantle).Foreground(T.Dim).Render(rotulo(g, on, nivel))
			}
			w := lipgloss.Width(r) + 2*gap
			abas = append(abas, [2]int{x, x + w})
			x += w
			b.WriteString(espacos(gap, T.Mantle) + r + espacos(gap, T.Mantle))
		}
		s = b.String()
		if lipgloss.Width(s) <= largura {
			break
		}
	}
	esq := max(0, (largura-lipgloss.Width(s))/2)
	for i := range abas {
		abas[i][0] += esq
		abas[i][1] += esq
	}
	return linhaFundo(s, largura, true), abas
}
