package ui

import (
	"fmt"
	"regexp"
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
)

var corTipo = map[dados.TipoFeed]lipgloss.Color{
	dados.TipoDecidir: T.Espera, dados.TipoJuntar: T.Ok, dados.TipoTravado: T.Erro, dados.TipoEsperar: T.Andamento, dados.TipoLer: T.Dim,
}

var reDone = regexp.MustCompile(`\bdone\b`)

func alvoDoFeed(i dados.ItemFeed) alvo {
	a := alvo{url: i.URL, aba: i.AbaHerdr}
	if i.Pergunta != nil {
		a.pergunta = i.Pergunta.Chave
	}
	if i.Tipo == dados.TipoLer {
		a.relatorio = i.Pasta
	}
	return a
}

func (c *corpo) itemFeed(i dados.ItemFeed, a alvo) {
	cor := corTipo[i.Tipo]
	verbo := i.Tipo.String()
	if c.largura >= 70 {
		c.item(a, func(bg lipgloss.Color) string {
			return st(bg, cor).Render("▌ "+Col(verbo, 8)) + st(bg, T.Texto).Bold(true).Render(i.Sujeito) + st(bg, T.Sub).Render("  "+i.Texto)
		})
		return
	}
	c.item(a,
		func(bg lipgloss.Color) string {
			return st(bg, cor).Render("▌ "+verbo+" ") + st(bg, T.Texto).Bold(true).Render(i.Sujeito)
		},
		func(bg lipgloss.Color) string {
			return st(bg, cor).Render("▌ ") + st(bg, T.Sub).Render(i.Texto)
		})
}

// envolve distribui os rótulos em linhas que cabem na largura.
func envolve(rotulos []string, largura int) [][]string {
	var out [][]string
	var atual []string
	w := 0
	for _, r := range rotulos {
		cel := 2 + lipgloss.Width(r) + 2
		if len(atual) > 0 && w+cel > largura {
			out = append(out, atual)
			atual, w = nil, 0
		}
		atual = append(atual, r)
		w += cel
	}
	if len(atual) > 0 {
		out = append(out, atual)
	}
	return out
}

func gavetaAgora(c *corpo, e dados.Estado, altura int) {
	var praVoce, travados, acomp []dados.ItemFeed
	for _, i := range dados.MontarFeed(e) {
		if !c.passa(i.Tipo.String(), i.Sujeito, i.Texto, i.URL) {
			continue
		}
		switch i.Tipo {
		case dados.TipoDecidir, dados.TipoJuntar:
			praVoce = append(praVoce, i)
		case dados.TipoTravado:
			travados = append(travados, i)
		default:
			acomp = append(acomp, i)
		}
	}
	var trab []string
	for _, t := range e.Tasks {
		if strings.Contains(t.Detalhe, "working") && c.casa(t.Rotulo) {
			r := t.Rotulo
			if c.largura < 70 {
				r = Corte(r, 18)
			}
			trab = append(trab, r)
		}
	}
	var vivas, parados []dados.LinhaViva
	for _, l := range e.Linhas {
		if l.Estado != dados.Trabalhando && l.Estado != dados.Parado || !c.passa(l.Nome, l.Etapa, l.Agora) {
			continue
		}
		switch l.Estado {
		case dados.Trabalhando:
			vivas = append(vivas, l)
		case dados.Parado:
			parados = append(parados, l)
		}
	}
	if len(e.Linhas) > 0 {
		trab = nil
	}
	grupos := envolve(trab, c.largura)
	porItem := 1
	if c.largura < 70 {
		porItem = 2
	}
	reservadas := 1
	if len(travados) > 0 {
		reservadas += 3
	}
	if len(trab) > 0 {
		reservadas += 2 + len(grupos)
	}
	if len(acomp) > 0 {
		reservadas += 3
	}
	if len(vivas) > 0 {
		reservadas += 2 + len(vivas)
	}
	if len(parados) > 0 {
		reservadas += 2 + len(parados)
	}
	cabem := max(2, (altura-reservadas)/porItem)

	lista := func(itens []dados.ItemFeed, cabe int) int {
		vis := itens[:min(len(itens), max(1, cabe))]
		for _, i := range vis {
			a := alvoDoFeed(i)
			if a.aba == "" && i.Pergunta != nil {
				a.aba = abaDaChave(e, i.Pergunta.Chave)
			}
			c.itemFeed(i, a)
		}
		maisNaFila(c, len(itens)-len(vis))
		return len(vis)
	}

	usados := 0
	if len(e.Mural) > 0 && c.filtro == "" {
		c.secao("mural", contagem(len(e.Mural)))
		for _, a := range e.Mural {
			a := a
			w := c.largura
			c.texto(func(bg lipgloss.Color) string { return renderAviso(a, w, bg) })
		}
		c.branco()
	}
	if len(praVoce) > 0 || c.filtro == "" {
		c.secao("precisa de você", contagem(len(praVoce)))
		if len(praVoce) == 0 {
			c.vazio("nada pedindo você agora ✓")
		} else {
			usados += lista(praVoce, cabem)
		}
	}

	if len(travados) > 0 {
		if len(c.linhas) > 0 {
			c.branco()
		}
		c.secao("travados", contagem(len(travados)))
		usados += lista(travados, cabem-usados)
	}

	if len(trab) > 0 {
		if len(c.linhas) > 0 {
			c.branco()
		}
		c.secao("trabalhando", fmt.Sprint(len(trab)))
		for _, g := range grupos {
			g := g
			c.texto(func(bg lipgloss.Color) string {
				var b strings.Builder
				for _, r := range g {
					b.WriteString(st(bg, T.Ok).Render("● ") + st(bg, T.Texto).Render(r) + espacos(2, bg))
				}
				return b.String()
			})
		}
	}

	raia := func(titulo string, ls []dados.LinhaViva) {
		if len(ls) == 0 {
			return
		}
		if len(c.linhas) > 0 {
			c.branco()
		}
		c.secao(titulo, fmt.Sprint(len(ls)))
		for _, l := range ls {
			c.linhaViva(l)
		}
	}
	raia("agora", vivas)
	raia("parados", parados)

	if len(acomp) > 0 {
		if len(c.linhas) > 0 {
			c.branco()
		}
		c.secao("acompanhando", fmt.Sprint(len(acomp)))
		lista(acomp, cabem-usados)
	}

	var antigas []dados.ItemFeed
	for _, i := range dados.MontarAntigas(e) {
		if c.casa(i.Sujeito, i.Texto) {
			antigas = append(antigas, i)
		}
	}
	if len(antigas) > 0 {
		if len(c.linhas) > 0 {
			c.branco()
		}
		dica := "a abre"
		if c.antigasAbertas {
			dica = "a recolhe"
		}
		n := len(antigas)
		aberta := c.antigasAbertas
		c.item(alvo{antigas: true}, func(bg lipgloss.Color) string {
			seta := "▸ "
			if aberta {
				seta = "▾ "
			}
			return st(bg, T.Acao).Render(seta) + st(bg, T.Sub).Bold(true).Render(fmt.Sprintf("antigas (%d)", n)) + st(bg, T.Dim).Render("  "+dica)
		})
		if c.antigasAbertas {
			for _, i := range antigas {
				c.passa(i.Sujeito, i.Texto)
				c.itemFeed(i, alvo{pergunta: i.Pergunta.Chave, arquivar: i.Pergunta.Chave})
			}
		}
	}

	if falta := dados.FaltaDe(e.Recursos); len(falta) > 0 && c.filtro == "" {
		c.branco()
		c.secao("falta nesta máquina", "")
		for _, f := range falta {
			f := f
			c.texto(func(bg lipgloss.Color) string { return st(bg, T.Dim).Render("○ " + f) })
		}
	}
}

func contagem(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprint(n)
}

func maisNaFila(c *corpo, n int) {
	if n > 0 {
		c.texto(func(bg lipgloss.Color) string { return st(bg, T.Dim).Render(fmt.Sprintf("  +%d na fila", n)) })
	}
}

var ptEstado = map[string]string{"working": "trabalhando", "idle": "parado", "done": "entregue", "blocked": "travado"}

func estadoTask(d string) string {
	partes := strings.Split(d, " · ")
	for _, k := range []string{"blocked", "working", "esperando", "idle", "done"} {
		for _, p := range partes {
			if strings.Contains(p, k) {
				return p
			}
		}
	}
	for _, p := range partes {
		if strings.Contains(p, "sem report") {
			return p
		}
	}
	return ""
}

func gavetaCrew(c *corpo, e dados.Estado) {
	if len(e.Linhas) > 0 {
		gavetaCrewViva(c, e)
		return
	}
	ativos := 0
	for _, t := range e.Tasks {
		if strings.Contains(t.Detalhe, "working") {
			ativos++
		}
	}
	c.secao("fantasminhas", fmt.Sprintf("%d ativos", ativos))
	n := len(e.Esperando)
	c.texto(func(bg lipgloss.Color) string {
		return st(bg, T.Dim).Render(fmt.Sprintf("%d pendente(s) no painel", n))
	})
	if len(e.Tasks) == 0 {
		c.vazio("nenhum fantasminha")
	}
	for _, x := range e.Tasks {
		cor, glifo := T.Dim, "○"
		switch {
		case strings.Contains(x.Detalhe, "blocked"):
			cor, glifo = T.Erro, "●"
		case strings.HasPrefix(x.Detalhe, "esperando"):
			cor, glifo = T.Espera, "●"
		case strings.Contains(x.Detalhe, "working"):
			cor, glifo = T.Ok, "●"
		}
		est := estadoTask(x.Detalhe)
		nome := est
		if pt, ok := ptEstado[est]; ok {
			nome = pt
		}
		var resto []string
		for _, p := range strings.Split(x.Detalhe, " · ") {
			if p != est {
				resto = append(resto, p)
			}
		}
		a := alvo{aba: x.Extra}
		if reDone.MatchString(x.Detalhe) && !strings.Contains(x.Detalhe, "sem report") {
			a.relatorio = x.Chave
		}
		c.linhaLista(a, glifo, cor, x.Rotulo, nome, cor, juntar(resto))
	}
}

func gavetaPRs(c *corpo, e dados.Estado) {
	var prs []dados.Travado
	ocultos := 0
	for _, x := range e.Travados {
		if x.Draft && !c.rascunhos {
			ocultos++
			continue
		}
		prs = append(prs, x)
	}
	c.secao("PRs", fmt.Sprintf("%d abertos", len(e.Travados)))
	if n := rascunhosDe(e); n > 0 {
		msg := fmt.Sprintf("%d rascunhos ocultos · d mostra", ocultos)
		if c.rascunhos {
			msg = fmt.Sprintf("%d rascunhos à mostra · d esconde", n)
		}
		c.texto(func(bg lipgloss.Color) string { return st(bg, T.Dim).Render(msg) })
	}
	if len(e.Travados) == 0 {
		c.vazio("nenhum PR aberto")
	} else if len(prs) == 0 {
		c.vazio("só rascunhos")
	}
	for _, x := range prs {
		cor, glifo := T.Andamento, "◐"
		switch dados.TomDoPR(x.Motivo) {
		case "ok":
			cor, glifo = T.Ok, "✓"
		case "aviso":
			cor, glifo = T.Erro, "✗"
		}
		titulo := regexp.MustCompile(`^\[[^\]]+\]\s*`).ReplaceAllString(x.Title, "")
		c.linhaLista(alvo{url: x.URL}, glifo, cor, fmt.Sprintf("%s #%d", strings.Replace(x.Repo, "Prolog", "", 1), x.Number), x.Motivo, cor, titulo)
	}
}

func gavetaCI(c *corpo, e dados.Estado) {
	c.secao("CI", "")
	if len(e.Travados) == 0 {
		c.vazio("sem PR aberto")
		return
	}
	r := dados.ResumoDoCI(e.Travados, c.rascunhos)
	nome := func(x dados.Travado) string {
		return fmt.Sprintf("%s #%d", strings.Replace(x.Repo, "Prolog", "", 1), x.Number)
	}
	for _, x := range r.Falhos {
		estado := "vermelho"
		if len(x.Falhos) > 0 {
			estado += " · " + strings.Join(x.Falhos, ", ")
		}
		c.linhaCI(alvo{url: x.URL + "/checks"}, "✗", T.Erro, nome(x), estado)
	}
	for _, x := range r.Rodando {
		c.linhaCI(alvo{url: x.URL + "/checks"}, "◐", T.Andamento, nome(x), "rodando")
	}
	if len(r.Falhos) == 0 && len(r.Rodando) == 0 {
		if r.Verdes > 0 {
			c.texto(func(bg lipgloss.Color) string { return st(bg, T.Ok).Render("✓ CI: tudo verde") })
		}
	} else if r.Verdes > 0 {
		msg := fmt.Sprintf("✓ %d PRs verdes", r.Verdes)
		if r.Verdes == 1 {
			msg = "✓ 1 PR verde"
		}
		c.texto(func(bg lipgloss.Color) string { return st(bg, T.Ok).Render(msg) })
	}
	if r.SemCI > 0 {
		msg := fmt.Sprintf("%d sem CI", r.SemCI)
		c.texto(func(bg lipgloss.Color) string { return st(bg, T.Dim).Render(msg) })
	}
}

// linhaCI: glifo, PR e estado numa linha só, cortada na largura.
func (c *corpo) linhaCI(a alvo, glifo string, cor lipgloss.Color, nome, estado string) {
	if !c.passa(nome, estado, a.url) {
		return
	}
	w := c.largura
	c.item(a, func(bg lipgloss.Color) string {
		resto := max(w-2-lipgloss.Width(nome)-2, 0)
		return st(bg, cor).Render(glifo+" ") + st(bg, T.Texto).Render(nome) + st(bg, cor).Render("  "+Corte(estado, resto))
	})
}

var rePorta = regexp.MustCompile(`:(\d{4})`)

func gavetaAmbiente(c *corpo, e dados.Estado) {
	c.secao("ambiente", "")
	if len(e.Servicos) == 0 {
		c.vazio("nenhum serviço conhecido")
	}
	for _, s := range e.Servicos {
		no := s.Rotulo != "parado"
		cor, glifo := T.Dim, "○"
		if no {
			cor, glifo = T.Ok, "●"
		}
		a := alvo{}
		if m := rePorta.FindStringSubmatch(s.Detalhe); m != nil && strings.Contains(s.Chave, "React") {
			a.url = "http://localhost:" + m[1]
		}
		c.linhaLista(a, glifo, cor, s.Chave, s.Rotulo, cor, s.Detalhe)
	}
}

func gavetaMaquina(c *corpo, e dados.Estado, pedindoDD bool) {
	c.secao("ações", "enter aciona")
	g, cor, est, det := botaoDaydreaming(e.Daydreamer, pedindoDD)
	c.linhaLista(alvo{acao: "daydreaming"}, g, cor, "atualizar DayDreaming", est, cor, det)
	g, cor, est, det = botaoMCP(e.MCP)
	c.linhaLista(alvo{acao: "mcp"}, g, cor, "reconectar MCPs", est, cor, det)
	if p := e.MCP; !p.Rodando && len(p.Sessoes) > 0 {
		for _, s := range p.Sessoes {
			cor, glifo := corTomMCP(s.Tom)
			c.linhaLista(alvo{}, "  "+glifo, cor, s.Nome, s.Tom, cor, s.Detalhe)
		}
	}
	c.branco()
	c.secao("máquina", "")
	if len(e.Maquina) == 0 {
		c.vazio("sem leitura da máquina")
	}
	for _, l := range e.Maquina {
		cor, glifo := T.Ok, "●"
		switch l.Extra {
		case "erro":
			cor, glifo = T.Erro, "✗"
		case "aviso":
			cor, glifo = T.Espera, "◆"
		}
		c.linhaLista(alvo{}, glifo, cor, l.Rotulo, "", cor, l.Detalhe)
	}
}

func gavetaMais(c *corpo, e dados.Estado) {
	c.secao("worktrees", "")
	if len(e.Worktrees) == 0 {
		c.vazio("nenhum worktree")
	}
	for _, l := range e.Worktrees {
		repo, branch, _ := strings.Cut(l.Rotulo, " · ")
		c.linhaLista(alvo{url: l.Chave}, "○", T.Dim, repo, "", T.Dim, branch)
	}
}

func rascunhosDe(e dados.Estado) int {
	n := 0
	for _, x := range e.Travados {
		if x.Draft {
			n++
		}
	}
	return n
}

func botaoDaydreaming(d dados.Daydreamer, pedindo bool) (glifo string, cor lipgloss.Color, estado, detalhe string) {
	glifo, cor, estado = "○", T.Dim, "sem aba"
	switch {
	case d.Status == "working":
		glifo, cor, estado = "●", T.Andamento, "trabalhando"
	case d.Status != "":
		glifo, cor, estado = "●", T.Ok, "no ar"
	}
	if pedindo {
		glifo, cor, estado = "◐", T.Andamento, "pedindo…"
	}
	detalhe = "última —"
	if !d.Ultima.IsZero() {
		detalhe = "última " + d.Ultima.Format("02/01 15:04")
	}
	return glifo, cor, estado, detalhe + " · D"
}

func botaoMCP(p dados.ProgressoMCP) (glifo string, cor lipgloss.Color, estado, detalhe string) {
	switch {
	case p.Rodando:
		total := "?"
		if p.Total > 0 {
			total = fmt.Sprint(p.Total)
		}
		return "◐", T.Andamento, "reconectando…", fmt.Sprintf("%d/%s sessões · M", len(p.Sessoes), total)
	case p.Interrompida:
		return "✗", T.Erro, "interrompida", "sem '# fim' em 30 min · M de novo"
	case p.Inicio.IsZero():
		return "○", T.Dim, "", "todas as sessões, todos os workspaces · M"
	}
	return "✓", T.Ok, "pronto", resumoMCP(p) + " · " + p.Inicio.Format("02/01 15:04") + " · M"
}

// resumoMCP conta as sessões por tom, ex.: 3 ok · 1 pulada.
func resumoMCP(p dados.ProgressoMCP) string {
	if p.Nota != "" {
		return p.Nota
	}
	n := map[string]int{}
	for _, s := range p.Sessoes {
		n[s.Tom]++
	}
	var ps []string
	for _, t := range []string{"ok", "parcial", "falhou", "pulada"} {
		if n[t] > 0 {
			ps = append(ps, fmt.Sprintf("%d %s", n[t], t))
		}
	}
	if len(ps) == 0 {
		return "nenhuma sessão"
	}
	return juntar(ps)
}

func corTomMCP(tom string) (lipgloss.Color, string) {
	switch tom {
	case "ok":
		return T.Ok, "✓"
	case "parcial":
		return T.Espera, "◆"
	case "falhou":
		return T.Erro, "✗"
	}
	return T.Dim, "○"
}

func gavetaCrewViva(c *corpo, e dados.Estado) {
	ativos := 0
	for _, l := range e.Linhas {
		if l.Estado == dados.Trabalhando {
			ativos++
		}
	}
	c.secao("fantasminhas", fmt.Sprintf("%d ativos", ativos))
	n := len(e.Esperando)
	c.texto(func(bg lipgloss.Color) string {
		return st(bg, T.Dim).Render(fmt.Sprintf("%d pendente(s) no painel", n))
	})
	c.branco()
	if w := c.largura - 2; w >= 56 {
		c.texto(func(bg lipgloss.Color) string { return cabecalhoViva(w, bg) })
	}
	for _, l := range linhasPorEstado(e.Linhas) {
		if c.passa(l.Nome, l.Etapa, l.Agora) {
			c.linhaViva(l)
		}
	}
}
