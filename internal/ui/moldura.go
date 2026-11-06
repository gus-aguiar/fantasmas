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

// trabalhando conta os fantasminhas trabalhando agora.
func trabalhando(e dados.Estado) int {
	n := 0
	if len(e.Linhas) > 0 {
		for _, l := range e.Linhas {
			if l.Estado == dados.Trabalhando {
				n++
			}
		}
		return n
	}
	for _, t := range e.Tasks {
		if strings.Contains(t.Detalhe, "working") {
			n++
		}
	}
	return n
}

// resumoTopo é "N precisa de você · N trabalhando" (contador zerado não aparece) e o alerta da máquina, se houver.
func resumoTopo(e dados.Estado) (string, string) {
	suas, trav := 0, 0
	for _, i := range dados.MontarFeed(e) {
		switch i.Tipo {
		case dados.TipoDecidir, dados.TipoJuntar:
			suas++
		case dados.TipoTravado:
			trav++
		}
	}
	var ps []string
	if suas > 0 {
		ps = append(ps, fmt.Sprintf("%d precisa de você", suas))
	}
	if trav > 0 {
		ps = append(ps, fmt.Sprintf("%d %s", trav, map[bool]string{true: "travado", false: "travados"}[trav == 1]))
	}
	if n := trabalhando(e); n > 0 {
		ps = append(ps, fmt.Sprintf("%d trabalhando", n))
	}
	if len(ps) == 0 {
		ps = append(ps, "nada pedindo você")
	}
	alerta := ""
	if as := alertasMaquina(e); len(as) > 0 {
		alerta = "⚠ " + strings.Join(as, " · ")
	}
	return strings.Join(ps, " · "), alerta
}

// linhaTopo é o topo de uma linha só (janela baixa): resumo e alerta à esquerda, ↻ hora à direita.
func linhaTopo(e dados.Estado, w, quadro int, coletando bool, status string) string {
	stl := func(c lipgloss.Color) lipgloss.Style { return fundo(T.Mantle).Foreground(c) }
	resumo, alerta := resumoTopo(e)
	esq := stl(T.Texto).Bold(true).Render(resumo)
	if alerta != "" {
		esq += stl(T.Dim).Render(" · ") + stl(T.EstadoAviso()).Render(alerta)
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

// topoArea desenha o topo: gengar em tamanho cheio à esquerda e, ao lado, título, resumo, alerta
// (só se houver) e a linha de atalhos. Com img, reserva em branco a área da imagem.
func topoArea(e dados.Estado, largura, altura, quadro int, coletando bool, status string, img bool) ([]string, areaImg) {
	if largura <= 0 {
		return []string{""}, areaImg{}
	}
	f := Faixa(largura)
	if altura < AlturaSemMascote || f == "minimo" {
		return []string{espacos(1, T.Mantle) + linhaTopo(e, largura-2, quadro, coletando, status) + espacos(1, T.Mantle)}, areaImg{}
	}
	stl := func(c lipgloss.Color) lipgloss.Style { return fundo(T.Mantle).Foreground(c) }
	glifo, quando := "↻", "atualizado "+hora(e)
	if coletando {
		glifo, quando = []string{"◐", "◓", "◑", "◒"}[((quadro%4)+4)%4], "atualizando…"
	}
	resumo, alerta := resumoTopo(e)
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

	l1 := stl(T.Acento()).Bold(true).Render("fantasmas") + stl(T.Dim).Render("  "+glifo+" "+quando)
	l2 := stl(T.Texto).Render(resumo)
	bruto := []string{l1, "", l2}
	if alerta != "" {
		bruto = append(bruto, stl(T.EstadoAviso()).Render(alerta))
	}
	if status != "" {
		bruto = append(bruto, "", stl(T.Texto).Render(status))
	}
	if artLin < len(bruto) {
		bruto = []string{l1, l2}
		if alerta != "" {
			bruto = append(bruto, stl(T.EstadoAviso()).Render(alerta))
		}
		if status != "" {
			bruto = append(bruto, stl(T.Texto).Render(status))
		}
	}
	cardW := 0
	for i, l := range bruto {
		if lipgloss.Width(l) > maxCard {
			l = ansi.Truncate(l, maxCard, "…")
			bruto[i] = l
		}
		cardW = max(cardW, lipgloss.Width(l))
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

// NomeAba é o rótulo curto de cada gaveta na linha de abas.
var NomeAba = map[string]string{"agora": "Agora", "missoes": "Missões", "prs": "PRs", "maquina": "Máquina", "configuracoes": "Config"}

// Rodape devolve a linha de abas (nome mantido por compatibilidade).
func Rodape(ativa string, largura int) string {
	s, _ := rodapeAbas(ativa, largura)
	return s
}

// rodapeAbas é a linha de abas, centralizada: "1 Agora · 2 Missões · …", a ativa em acento e negrito,
// o dígito sempre no acento. Se couber, "? ajuda" em cinza no canto direito. Devolve também o
// intervalo [x0, x1) de cada aba, para o clique.
func rodapeAbas(ativa string, largura int) (string, [][2]int) {
	if largura <= 0 {
		return "", nil
	}
	const dica = "? ajuda"
	for _, curto := range []bool{false, true} {
		var partes []string
		var rel [][2]int
		x := 0
		for i, g := range Gavetas {
			on := g.ID == ativa
			if i > 0 {
				sep := " · "
				if curto {
					sep = "  "
				}
				partes = append(partes, st(T.Base, T.Dim).Render(sep))
				x += lipgloss.Width(sep)
			}
			nome := NomeAba[g.ID]
			if curto && !on {
				nome = ""
			}
			r := st(T.Base, T.Acento()).Render(g.Tecla)
			if nome != "" {
				cor := st(T.Base, T.Dim)
				if on {
					cor = st(T.Base, T.Acento()).Bold(true)
				}
				r += cor.Render(" " + nome)
			}
			w := lipgloss.Width(r)
			rel = append(rel, [2]int{x, x + w})
			x += w
			partes = append(partes, r)
		}
		if x > largura && !curto {
			continue
		}
		esq := max(0, (largura-x)/2)
		abas := make([][2]int, len(rel))
		for i, r := range rel {
			abas[i] = [2]int{r[0] + esq, r[1] + esq}
		}
		linha := espacos(esq, T.Base) + strings.Join(partes, "")
		if falta := largura - esq - x; falta >= len(dica)+2 {
			linha += espacos(falta-len(dica)-1, T.Base) + st(T.Base, T.Dim).Render(dica)
		}
		return encaixa(linha, largura, T.Base), abas
	}
	return "", nil
}
