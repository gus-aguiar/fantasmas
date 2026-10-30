package ui

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// corMissao é o acento estável da missão, derivado do slug.
func corMissao(slug string) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(slug))
	return T.Acentos[h.Sum32()%uint32(len(T.Acentos))]
}

func missaoDe(e dados.Estado, slug string) (dados.Missao, bool) {
	for _, m := range e.Missoes {
		if m.Slug == slug {
			return m, true
		}
	}
	return dados.Missao{}, false
}

func nomeMissao(e dados.Estado, slug string) string {
	if m, ok := missaoDe(e, slug); ok && m.Nome != "" {
		return m.Nome
	}
	return slug
}

// FiltrarMissao devolve o estado com só o que pertence à missão (mais os avisos globais do mural).
func FiltrarMissao(e dados.Estado, slug string) dados.Estado {
	m, _ := missaoDe(e, slug)
	return filtrarPor(e, slug, m.Workspace)
}

func noWs(aba, ws string) bool { return ws != "" && strings.HasPrefix(aba, ws+":") }

func filtrarPor(e dados.Estado, slug, ws string) dados.Estado {
	m, _ := missaoDe(e, slug)
	da := func(missao, aba string) bool { return slug != "" && missao == slug || noWs(aba, ws) }
	f := e
	f.Tasks, f.Esperando, f.Linhas, f.Runs, f.Mural, f.Travados, f.Caidos = nil, nil, nil, nil, nil, nil, nil
	pastas := map[string]bool{}
	for _, t := range e.Tasks {
		if da(t.Missao, t.Extra) {
			f.Tasks = append(f.Tasks, t)
			pastas[t.Chave] = true
		}
	}
	for _, p := range e.Esperando {
		dono := false
		for k := range pastas {
			if strings.HasPrefix(p.Chave, k+"/") {
				dono = true
			}
		}
		if dono || slug != "" && p.Missao == slug {
			f.Esperando = append(f.Esperando, p)
		}
	}
	for _, l := range e.Linhas {
		if slug != "" && l.Missao == slug || ws != "" && l.Workspace == ws {
			f.Linhas = append(f.Linhas, l)
		}
	}
	for _, r := range e.Runs {
		if slug != "" && r.Missao == slug {
			f.Runs = append(f.Runs, r)
		}
	}
	for _, a := range e.Mural {
		if a.Missao == "" || a.Missao == slug {
			f.Mural = append(f.Mural, a)
		}
	}
	for _, p := range e.Travados {
		if dados.CitaCard(p.Title, m.Cards) {
			f.Travados = append(f.Travados, p)
		}
	}
	for _, c := range e.Caidos {
		if pastas[c.Pasta] || noWs(c.Pane, ws) {
			f.Caidos = append(f.Caidos, c)
		}
	}
	return f
}

// ComMissao abre o app com a missão em foco (usado pelo --uma-vez).
func (m Model) ComMissao(slug string) Model {
	m.missao = slug
	return m
}

// estadoVisto é o estado que as gavetas desenham: inteiro, ou só a missão em foco.
func (m Model) estadoVisto() dados.Estado {
	if m.missao == "" {
		return m.estado
	}
	return FiltrarMissao(m.estado, m.missao)
}

// seloMissao é a faixa "◆ <nome>" do topo enquanto uma missão está em foco.
func seloMissao(e dados.Estado, slug string, largura int) string {
	cor := corMissao(slug)
	s := fundo(T.Mantle).Foreground(cor).Bold(true).Render(" ◆ "+nomeMissao(e, slug)) +
		fundo(T.Mantle).Foreground(T.Dim).Render("   só esta missão · esc ou 0 mostra todas")
	return linhaFundo(s, largura, false)
}

type rankMissao int

const (
	missaoEspera rankMissao = iota
	missaoTrabalhando
	missaoParada
	missaoConcluida
)

func estadoMissao(m dados.Missao) rankMissao {
	c := m.Contagem
	switch {
	case m.Estado == "concluida":
		return missaoConcluida
	case c.Pendentes > 0:
		return missaoEspera
	case c.Trabalhando > 0 || c.PipesAtivos > 0:
		return missaoTrabalhando
	}
	return missaoParada
}

func glifoMissao(r rankMissao) (string, lipgloss.Color) {
	switch r {
	case missaoEspera:
		return "▲", T.Espera
	case missaoTrabalhando:
		return "●", T.Andamento
	case missaoConcluida:
		return "✓", T.Ok
	}
	return "◌", T.Dim
}

func haQuanto(t, agora time.Time) string {
	if t.IsZero() {
		return "sem atividade"
	}
	d := agora.Sub(t)
	switch {
	case d < time.Minute:
		return "agora há pouco"
	case d < time.Hour:
		return fmt.Sprintf("há %d min", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("há %dh", int(d.Hours()))
	}
	return fmt.Sprintf("há %d dias", int(d.Hours()/24))
}

// ladoALado junta esquerda e direita na largura w; sem espaço, a direita sai.
func ladoALado(esq, dir string, w int, bg lipgloss.Color) string {
	we, wd := lipgloss.Width(esq), lipgloss.Width(dir)
	if we+wd+2 > w {
		return ansi.Truncate(esq, w, "…")
	}
	return esq + espacos(w-we-wd, bg) + dir
}

func chip(bg, fg lipgloss.Color, s string) string {
	return lipgloss.NewStyle().Background(bg).Foreground(fg).Render(" " + s + " ")
}

func gavetaMissoes(c *corpo, e dados.Estado, foco string) {
	if !e.MissoesLigado {
		c.branco()
		c.vazio("missões: ligue FANTASMAS_FONTE=ddb")
		c.vazio("o filtro por workspace segue valendo pelo meta")
		return
	}
	w := c.largura - 2
	faixaPrincipal(c, e, w)
	c.branco()
	ms := append([]dados.Missao(nil), e.Missoes...)
	sort.SliceStable(ms, func(i, j int) bool { return estadoMissao(ms[i]) < estadoMissao(ms[j]) })
	abertas, suas := 0, 0
	for _, m := range ms {
		if m.Estado != "concluida" {
			abertas++
		}
		suas += m.Contagem.Pendentes
	}
	extra := fmt.Sprintf("%d abertas", abertas)
	if suas > 0 {
		extra += fmt.Sprintf(" · %d esperando você", suas)
	}
	c.secao("missões", extra+" · enter foca · g vai ao gengar")
	if len(ms) == 0 {
		c.vazio("nenhuma missão aberta · scripts/delegate/missao-abrir.sh")
		return
	}
	for i, m := range ms {
		if !c.passa(append([]string{m.Nome, m.Slug, m.Origem, m.OrigemRef, m.Gengar}, m.Cards...)...) {
			continue
		}
		if i > 0 {
			c.branco()
		}
		carta(c, e, m, m.Slug == foco, w)
	}
}

func abaDoCoordenador(e dados.Estado, nome string) string {
	if l, ok := e.Coordenadores[nome]; ok && l.Aba != "" {
		return l.Aba
	}
	return e.Agentes[nome].Aba
}

func faixaPrincipal(c *corpo, e dados.Estado, w int) {
	p := e.Principal
	nome := p.Nome
	if nome == "" {
		nome = "Genguinha"
	}
	l, vivo := e.Coordenadores[p.Gengar]
	c.texto(func(bg lipgloss.Color) string {
		b := faixaBg(bg)
		esq := st(b, T.Acao).Render("▌ ") + st(b, T.Acao).Bold(true).Render("✦ GENGAR") +
			st(b, T.Texto).Bold(true).Render("  "+nome) + st(b, T.Dim).Render("  coordena todas as missões")
		dir := st(b, T.Dim).Render(p.Workspace + " ")
		return ladoALado(esq, dir, w, b)
	})
	c.texto(func(bg lipgloss.Color) string {
		b := faixaBg(bg)
		return encaixa(st(b, T.Acao).Render("▌ ")+linhaGengar(p.Gengar, l, vivo, T.Acao, w-2, b), w, b)
	})
}

// faixaBg: a faixa do principal fica em Surface; selecionada, Sel.
func faixaBg(bg lipgloss.Color) lipgloss.Color {
	if bg == T.Sel {
		return bg
	}
	return T.Surface
}

// linhaGengar: nome do coordenador, ctx e o que ele faz agora (a mesma leitura da linha viva).
func linhaGengar(nome string, l dados.LinhaViva, vivo bool, cor lipgloss.Color, w int, bg lipgloss.Color) string {
	if nome == "" {
		return st(bg, T.Dim).Render("sem coordenador")
	}
	wn := 22
	if w < 60 {
		wn = 15
	}
	s := st(bg, cor).Bold(true).Render(Col(nome, wn)) + espacos(1, bg)
	if !vivo {
		return s + st(bg, T.Dim).Render("○ fora do Herdr")
	}
	g, cg := glifoLinha(l.Estado)
	if w >= 56 {
		b := barraCtxEm(l.CtxPct, bg)
		s += b + espacos(wBarraViva-lipgloss.Width(b), bg)
	}
	resto := w - lipgloss.Width(s) - 2
	if resto > 0 {
		s += st(bg, cg).Render(g+" ") + st(bg, corAgora(l.Estado)).Render(Corte(l.Agora, resto))
	}
	return s
}

func carta(c *corpo, e dados.Estado, m dados.Missao, emFoco bool, w int) {
	acento := corMissao(m.Slug)
	r := estadoMissao(m)
	g, cg := glifoMissao(r)
	a := alvo{missao: m.Slug, aba: abaDoCoordenador(e, m.Gengar)}
	if strings.HasPrefix(m.Prototipo, "http") {
		a.url = m.Prototipo
	}
	faixa := func(bg lipgloss.Color) string { return st(bg, acento).Render("▌ ") }
	largo := w >= 70
	l1 := func(bg lipgloss.Color) string {
		nomeCor := T.Texto
		if r == missaoConcluida {
			nomeCor = T.Sub
		}
		esq := faixa(bg) + st(bg, cg).Render(g+" ") + st(bg, nomeCor).Bold(true).Render(m.Nome) + espacos(2, bg)
		orig := m.Origem
		if m.OrigemRef != "" && m.OrigemRef != m.Origem {
			orig += " · " + m.OrigemRef
		}
		if orig != "" {
			esq += chip(acento, T.Mantle, orig)
		}
		if largo {
			for _, k := range m.Cards {
				esq += espacos(1, bg) + chip(T.Chip, T.ChipTexto, k)
			}
		}
		// sem alinhar à direita: o selecionado ganha os botões ali sem cortar a carta.
		if m.Workspace != "" {
			esq += st(bg, T.Dim).Render("  " + m.Workspace)
		}
		if emFoco {
			esq += st(bg, acento).Bold(true).Render("  ◆ em foco")
		}
		if lipgloss.Width(esq) > w {
			esq = ansi.Truncate(esq, w, "…")
		}
		return esq
	}
	l2 := func(bg lipgloss.Color) string {
		l, vivo := e.Coordenadores[m.Gengar]
		return encaixa(faixa(bg)+espacos(2, bg)+linhaGengar(m.Gengar, l, vivo, acento, w-4, bg), w, bg)
	}
	l3 := func(bg lipgloss.Color) string {
		return ladoALado(faixa(bg)+espacos(2, bg)+contagens(m.Contagem, bg, largo), st(bg, T.Dim).Render(haQuanto(m.UltimaAtividade, e.At)), w, bg)
	}
	c.item(a, l1, l2, l3)
}

func contagens(k dados.Contagem, bg lipgloss.Color, largo bool) string {
	type parte struct {
		n          int
		g, nome    string
		plural     string
		cor, corTx lipgloss.Color
	}
	pv := 0
	if k.PrsVermelhos != nil {
		pv = *k.PrsVermelhos
	}
	ps := []parte{
		{k.Trabalhando, "●", "trabalhando", "trabalhando", T.Andamento, T.Sub},
		{k.Parados, "◌", "parado", "parados", T.Dim, T.Sub},
		{k.Pendentes, "▲", "sua", "suas", T.Espera, T.Espera},
		{k.PipesAtivos, "⟳", "pipe", "pipes", T.Acao, T.Sub},
		{pv, "✗", "PR vermelho", "PRs vermelhos", T.Erro, T.Erro},
	}
	var out []string
	for _, p := range ps {
		if p.n == 0 {
			continue
		}
		nome := p.nome
		if p.n > 1 {
			nome = p.plural
		}
		s := st(bg, p.cor).Render(p.g) + st(bg, T.Texto).Bold(true).Render(fmt.Sprintf(" %d", p.n))
		if largo {
			s += st(bg, p.corTx).Render(" " + nome)
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return st(bg, T.Dim).Render("ninguém trabalhando agora")
	}
	return strings.Join(out, espacos(3, bg))
}
