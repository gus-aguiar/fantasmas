package ui

import (
	"fmt"
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

// Gaveta é uma aba do rodapé; Descricao é a frase curta que diz para que ela serve, ao lado da primeira seção.
type Gaveta struct{ ID, Tecla, Nome, Descricao string }

var Gavetas = []Gaveta{
	{"agora", "1", "agora", "O que precisa de você e quem está trabalhando"},
	{"missoes", "2", "missões", "Frentes abertas, sidequests e decisões"},
	{"prs", "3", "PRs", "Seus PRs e o CI"},
	{"maquina", "4", "máquina", "Computador, serviços e worktrees"},
	{"configuracoes", "5", "configurações", "Tema, figura e variáveis"},
}

// apelidosGaveta mantém os nomes antigos do --gaveta valendo.
var apelidosGaveta = map[string]string{
	"crew": "agora", "ambiente": "maquina", "worktrees": "maquina", "comandos": "configuracoes",
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
	ls, _ := topoArea(e, largura, 40, quadro, coletando, status, false)
	return strings.Join(ls, "\n")
}

// AlturaSemMascote: abaixo dela o mascote some e o topo fica numa linha.
const AlturaSemMascote = 30

// resumoTopo é "▲ N suas · ● N trabalhando" mais o alerta da máquina, só quando existe.
func resumoTopo(e dados.Estado) (string, string) {
	suas := 0
	for _, i := range dados.MontarFeed(e) {
		if i.Tipo == dados.TipoDecidir || i.Tipo == dados.TipoJuntar {
			suas++
		}
	}
	trab := 0
	if len(e.Linhas) > 0 {
		for _, l := range e.Linhas {
			if l.Estado == dados.Trabalhando {
				trab++
			}
		}
	} else {
		for _, t := range e.Tasks {
			if strings.Contains(t.Detalhe, "working") {
				trab++
			}
		}
	}
	var ps []string
	if suas > 0 {
		ps = append(ps, fmt.Sprintf("▲ %d %s", suas, map[bool]string{true: "sua", false: "suas"}[suas == 1]))
	}
	if trab > 0 {
		ps = append(ps, fmt.Sprintf("● %d trabalhando", trab))
	}
	if len(ps) == 0 {
		ps = append(ps, "✓ nada pedindo você")
	}
	alerta := ""
	if as := alertasMaquina(e); len(as) > 0 {
		alerta = "⚠ " + as[0]
	}
	return strings.Join(ps, " · "), alerta
}

// linhaTopo é a linha única do topo: resumo e alerta à esquerda, ↻ hora à direita.
func linhaTopo(e dados.Estado, w, quadro int, coletando bool, status string) string {
	stl := func(c lipgloss.Color) lipgloss.Style { return fundo(T.Mantle).Foreground(c) }
	resumo, alerta := resumoTopo(e)
	esq := stl(T.Texto).Bold(true).Render(resumo)
	if alerta != "" {
		esq += stl(T.Dim).Render(" · ") + stl(T.Espera).Bold(true).Render(alerta)
	}
	if status != "" {
		esq += stl(T.Dim).Render("   ") + stl(T.Sub).Render(status)
	}
	glifo := "↻ " + hora(e)
	if coletando {
		glifo = []string{"◐", "◓", "◑", "◒"}[((quadro%4)+4)%4] + " atualizando"
	}
	dir := stl(T.Dim).Render(glifo + "  ? ajuda")
	if lipgloss.Width(esq)+lipgloss.Width(dir)+2 > w {
		dir = stl(T.Dim).Render(glifo)
	}
	if lipgloss.Width(esq)+lipgloss.Width(dir)+2 > w {
		return linhaFundo(esq, w, false)
	}
	return esq + espacos(w-lipgloss.Width(esq)-lipgloss.Width(dir), T.Mantle) + dir
}

// topoArea desenha o topo; com img, reserva em branco a área da imagem e devolve onde ela fica.
// O mascote tem no máximo 4 linhas e some abaixo de AlturaSemMascote.
func topoArea(e dados.Estado, largura, altura, quadro int, coletando bool, status string, img bool) ([]string, areaImg) {
	if largura <= 0 {
		return []string{""}, areaImg{}
	}
	if altura < AlturaSemMascote || Faixa(largura) == "minimo" {
		return []string{espacos(1, T.Mantle) + linhaTopo(e, largura-2, quadro, coletando, status) + espacos(1, T.Mantle)}, areaImg{}
	}
	gap := 2
	arte, artCol, artLin := QuadroMascote, MascoteColunas, MascoteLinhas
	if img {
		artCol, artLin = tamanhoImagem(largura)
		c, l := artCol, artLin
		arte = func(int, lipgloss.Color) []string { return centraV(nil, c, l) }
	} else if sp := spriteFn(largura); sp != nil {
		arte = func(i int, _ lipgloss.Color) []string { return sp.Quadro(i) }
		artCol, artLin = sp.Colunas, sp.Linhas
	}
	a := arte(quadro, T.Mantle)
	if len(a) > 4 {
		a = a[:4]
	}
	artLin = min(artLin, 4)
	alt := artLin
	a = centraV(a, artCol, alt)
	w := max(1, largura-2-artCol-gap)
	cartao := centraV([]string{linhaTopo(e, w, quadro, coletando, status)}, w, alt)
	g := espacos(gap, T.Mantle)
	linhas := make([]string, alt)
	for i := range linhas {
		linhas[i] = linhaFundo(espacos(1, T.Mantle)+a[i]+g+linhaFundo(cartao[i], w, false), largura, false)
	}
	return linhas, areaImg{x: 1, y: 0, c: artCol, r: artLin, ok: img}
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
	if dica := fundo(T.Mantle).Foreground(T.Dim).Render("  " + DicaBotoes); lipgloss.Width(s)+lipgloss.Width(dica) <= largura {
		s += dica
	}
	esq := max(0, (largura-lipgloss.Width(s))/2)
	for i := range abas {
		abas[i][0] += esq
		abas[i][1] += esq
	}
	return linhaFundo(s, largura, true), abas
}
