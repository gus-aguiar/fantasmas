package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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

// seloFeed distingue, na lista "precisa de você", decisão de liberação.
var seloFeed = map[dados.TipoFeed]string{dados.TipoDecidir: "decisão", dados.TipoJuntar: "liberação"}

func (c *corpo) itemFeed(i dados.ItemFeed, a alvo) {
	cor := corTipo[i.Tipo]
	verbo := i.Tipo.String()
	if s, ok := seloFeed[i.Tipo]; ok {
		verbo = s
	}
	if c.largura >= 70 {
		c.item(a, func(bg lipgloss.Color) string {
			return st(bg, cor).Render("▌ "+Col(verbo, 10)) + st(bg, T.Texto).Bold(true).Render(i.Sujeito) + st(bg, T.Sub).Render("  "+i.Texto)
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

// caidoAntigo: caído de mais de um dia vai para "antigos", a um clique.
func caidoAntigo(x dados.Caido, agora dados.Estado) bool {
	return !x.Em.IsZero() && !agora.At.IsZero() && agora.At.Sub(x.Em) > 24*time.Hour
}

func gavetaAgora(c *corpo, e dados.Estado, altura int) {
	var praVoce, travados, acomp []dados.ItemFeed
	for _, i := range dados.MontarFeed(e) {
		if !c.passa(i.Tipo.String(), seloFeed[i.Tipo], i.Sujeito, i.Texto, i.URL) {
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
	var vivas, parados []dados.LinhaViva
	for _, l := range e.Linhas {
		if l.Estado != dados.Trabalhando && l.Estado != dados.Parado || !c.passa(l.Nome, l.Etapa, l.Agora) {
			continue
		}
		if l.Estado == dados.Trabalhando {
			vivas = append(vivas, l)
		} else {
			parados = append(parados, l)
		}
	}
	var recentes, velhos []dados.Caido
	for _, x := range e.Caidos {
		if caidoAntigo(x, e) {
			velhos = append(velhos, x)
		} else {
			recentes = append(recentes, x)
		}
	}
	porItem := 1
	if c.largura < 70 {
		porItem = 2
	}
	reservadas := 1 + len(vivas) + len(parados) + len(recentes)
	if len(travados) > 0 {
		reservadas += 3
	}
	if len(acomp) > 0 {
		reservadas += 3
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
	espaco := func() {
		if len(c.linhas) > 0 {
			c.branco()
		}
	}

	usados := 0
	if len(praVoce) > 0 || c.filtro == "" {
		c.secaoN("precisa de você", len(praVoce), "")
		if len(praVoce) == 0 {
			c.vazioIcone("✓", T.Ok, "nada pedindo você agora")
		} else {
			usados += lista(praVoce, cabem)
		}
	}
	if len(travados) > 0 {
		espaco()
		c.secaoN("travados", len(travados), "")
		usados += lista(travados, cabem-usados)
	}

	// fantasminhas: trabalhando primeiro, depois os parados e os caídos recentes
	tarefas := len(e.Linhas) == 0 && len(e.Tasks) > 0
	if len(vivas)+len(parados)+len(recentes) > 0 || tarefas {
		espaco()
		extra := ""
		if len(recentes) > 0 {
			extra = "R 2× retoma caído"
		}
		c.secao("fantasminhas", extra)
		for _, l := range vivas {
			c.linhaViva(l)
		}
		for _, l := range parados {
			c.linhaViva(l)
		}
		if tarefas {
			linhasTasks(c, e)
		}
		for _, x := range recentes {
			linhaCaido(c, x)
		}
	}

	if len(e.Mural) > 0 && c.filtro == "" {
		espaco()
		c.secao("mural", "")
		for _, a := range e.Mural {
			a := a
			w := c.largura
			c.texto(func(bg lipgloss.Color) string { return renderAviso(a, w, bg) })
		}
	}

	if len(acomp) > 0 {
		espaco()
		c.secaoN("acompanhando", len(acomp), "")
		lista(acomp, cabem-usados)
	}

	var antigas []dados.ItemFeed
	for _, i := range dados.MontarAntigas(e) {
		if c.casa(i.Sujeito, i.Texto) {
			antigas = append(antigas, i)
		}
	}
	if n := len(antigas) + len(velhos); n > 0 {
		espaco()
		dica := "enter abre"
		if c.antigasAbertas {
			dica = "enter recolhe"
		}
		aberta := c.antigasAbertas
		c.item(alvo{antigas: true}, func(bg lipgloss.Color) string {
			seta := "▸ "
			if aberta {
				seta = "▾ "
			}
			return st(bg, T.Acao).Render(seta) + st(bg, T.Sub).Bold(true).Render(fmt.Sprintf("antigos (%d)", n)) + st(bg, T.Dim).Render("  "+dica)
		})
		if c.antigasAbertas {
			for _, i := range antigas {
				c.passa(i.Sujeito, i.Texto)
				c.itemFeed(i, alvo{pergunta: i.Pergunta.Chave, arquivar: i.Pergunta.Chave})
			}
			for _, x := range velhos {
				linhaCaido(c, x)
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

// linhaCaido é um fantasminha fora do Herdr que dá para retomar com R.
func linhaCaido(c *corpo, x dados.Caido) {
	sid := x.SessionID
	if len(sid) > 8 {
		sid = sid[:8]
	}
	det := "sessão " + sid
	if sid == "" {
		det = "sem session_id"
	}
	if !x.Em.IsZero() {
		det += " · " + x.Em.Format("02/01 15:04")
	}
	c.linhaLista(alvo{retomar: x.Nome}, "↻", T.Acao, x.Nome, "caído", T.Espera, det+" · pane "+x.Pane)
}

// linhasTasks é a lista de fantasminhas sem linha viva (fonte sem transcripts): só os não entregues.
func linhasTasks(c *corpo, e dados.Estado) {
	for _, x := range e.Tasks {
		if reDone.MatchString(x.Detalhe) {
			continue
		}
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
		c.linhaLista(alvo{aba: x.Extra}, glifo, cor, x.Rotulo, nome, cor, juntar(resto))
	}
}

// gavetaPRs junta PRs e CI: uma linha por PR com o estado do PR, o CI e o título; o resumo do CI vai
// no título da seção.
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
	c.secaoN("PRs", len(e.Travados), resumoCI(dados.ResumoDoCI(e.Travados, c.rascunhos)))
	if n := rascunhosDe(e); n > 0 {
		msg := fmt.Sprintf("%d rascunhos ocultos · enter ou d mostra", ocultos)
		if c.rascunhos {
			msg = fmt.Sprintf("%d rascunhos à mostra · d esconde", n)
		}
		c.item(alvo{alterna: "rascunhos"}, func(bg lipgloss.Color) string { return st(bg, T.Acao).Render("▸ ") + st(bg, T.Dim).Render(msg) })
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
		gci, cci, tci := colunaCI(x)
		c.linhaPR(alvo{url: x.URL}, glifo, cor, nomePR(x), x.Motivo, gci, cci, tci, titulo)
	}
}

func nomePR(x dados.Travado) string {
	return fmt.Sprintf("%s #%d", strings.Replace(x.Repo, "Prolog", "", 1), x.Number)
}

// resumoCI é o extra do título dos PRs: só os estados que existem, vermelho primeiro.
func resumoCI(r dados.ResumoCI) string {
	var ps []string
	if n := len(r.Falhos); n > 0 {
		ps = append(ps, fmt.Sprintf("✗ %d", n))
	}
	if n := len(r.Rodando); n > 0 {
		ps = append(ps, fmt.Sprintf("◐ %d", n))
	}
	if r.Verdes > 0 {
		ps = append(ps, fmt.Sprintf("✓ %d", r.Verdes))
	}
	if r.SemCI > 0 {
		ps = append(ps, fmt.Sprintf("– %d", r.SemCI))
	}
	if len(ps) == 0 {
		return ""
	}
	return "CI: " + strings.Join(ps, " · ")
}

// colunaCI é o glifo, a cor e o texto do CI de um PR; vermelho leva os checks que falharam.
func colunaCI(x dados.Travado) (string, lipgloss.Color, string) {
	switch x.CI {
	case "FAILURE", "ERROR":
		t := "vermelho"
		if len(x.Falhos) > 0 {
			t = strings.Join(x.Falhos, ", ")
		}
		return "✗", T.Erro, t
	case "PENDING", "EXPECTED":
		return "◐", T.Andamento, "rodando"
	case "SUCCESS":
		return "✓", T.Ok, "verde"
	}
	return "–", T.Dim, "sem CI"
}

// linhaPR: glifo, PR, estado do PR, CI e título, em colunas que encolhem com a largura.
func (c *corpo) linhaPR(a alvo, glifo string, cor lipgloss.Color, nome, estado, gci string, cci lipgloss.Color, tci, titulo string) {
	wNome, wEstado, wCI := 22, 22, 18
	switch {
	case c.largura < 70:
		wNome, wEstado, wCI = 18, 12, 0
	case c.largura < 110:
		wNome, wEstado, wCI = 20, 16, 12
	}
	// tela estreita demais: o nome encolhe primeiro e o estado fica com o que sobrar
	wNome = min(wNome, max(1, c.largura-2))
	wEstado = min(wEstado, max(0, c.largura-2-wNome))
	wTit := c.largura - 2 - wNome - wEstado - wCI
	if !c.passa(nome, estado, tci, titulo, a.url) {
		return
	}
	// col deixa sempre um espaço entre as colunas, mesmo com o texto cortado
	col := func(s string, n int) string {
		if n <= 1 {
			return Col(s, n)
		}
		return Col(s, n-1) + " "
	}
	c.item(a, func(bg lipgloss.Color) string {
		s := st(bg, cor).Render(glifo+" ") + st(bg, T.Texto).Render(col(nome, wNome)) + st(bg, cor).Render(col(estado, wEstado))
		if wCI > 0 {
			s += st(bg, cci).Render(col(gci+" "+tci, wCI))
		}
		if wTit > 6 {
			s += st(bg, T.Dim).Render(Corte(titulo, wTit))
		}
		return s
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

// gavetaMaquina é só monitoramento: a saúde do computador, relida a cada coleta (r atualiza na hora).
func gavetaMaquina(c *corpo, e dados.Estado) {
	c.secao("máquina", "")
	if len(e.Maquina) == 0 {
		c.vazioIcone("◆", T.Espera, "sem leitura da máquina", "o maquina.sh não respondeu")
	}
	for _, l := range e.Maquina {
		cor, glifo := T.Ok, "●"
		switch l.Extra {
		case "erro":
			cor, glifo = T.Erro, "✗"
		case "aviso":
			cor, glifo = T.Espera, "◆"
		case "neutro":
			cor, glifo = T.Dim, "○"
		}
		c.linhaLista(alvo{}, glifo, cor, l.Rotulo, "", cor, l.Detalhe)
	}
}

// gavetaWorktrees lista onde cada branch está aberto no disco: o Spark e os repos dentro dele
// (coleta.sh worktrees); enter abre a pasta.
func gavetaWorktrees(c *corpo, e dados.Estado) {
	c.secao("worktrees", "enter abre a pasta")
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
