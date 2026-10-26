package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const margem = 2

func (m Model) ler(p string) ([]byte, error) {
	if m.acoes.LerArquivo != nil {
		return m.acoes.LerArquivo(p)
	}
	return os.ReadFile(p)
}

func (m Model) naRaiz(p string) string {
	if filepath.IsAbs(p) || m.estado.Raiz == "" {
		return p
	}
	return filepath.Join(m.estado.Raiz, p)
}

// relativo tira a raiz do Spark do texto do erro para o status mostrar caminho curto.
func (m Model) relativo(err error) error {
	if err == nil || m.estado.Raiz == "" {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), strings.TrimRight(m.estado.Raiz, "/")+"/", ""))
}

// abrirDetalhe lê a pergunta (questions.json#id) ou o relatório (<pasta>/report.json) e abre a tela.
func (m Model) abrirDetalhe(chave, tipo string) Model {
	var a *dados.Aberto
	var err error
	if tipo == "relatorio" {
		var b []byte
		if b, err = m.ler(filepath.Join(chave, "report.json")); err == nil {
			a, err = dados.RelatorioDe(chave, b)
		} else if os.IsNotExist(err) {
			md := filepath.Join(chave, "report.md")
			if b, err = m.ler(md); err == nil {
				a = dados.RelatorioMdDe(chave, md, string(b))
			} else if os.IsNotExist(err) {
				m.status = "não abri: a task não tem report.json nem report.md"
				return m
			}
		}
	} else {
		arquivo, id := dados.Separar(chave)
		var b []byte
		if b, err = m.ler(arquivo); err == nil {
			a, err = dados.PerguntaDe(chave, b, id)
		}
	}
	if err != nil {
		return m.erro("não abri: ", m.relativo(err))
	}
	if a == nil {
		m.status = "não abri: a pergunta não está mais no arquivo"
		return m
	}
	if p := dados.PlanoDe(a.Documentos); p != "" {
		a.PreviaDe = p
		if b, err := m.ler(m.naRaiz(p)); err == nil {
			a.Previa = dados.PreviaDe(string(b), 24)
		} else {
			a.Previa = []string{"(arquivo não encontrado nesta máquina)"}
		}
	}
	m.aberto, m.opcao, m.rolagem, m.diagrama, m.rascunho, m.armado, m.focoTexto, m.filtro = a, 0, 0, false, nil, 0, false, nil
	if tipo == "relatorio" {
		m = m.marcarVisto(chave)
	}
	for i, o := range a.Opcoes {
		if o.Padrao {
			m.opcao = i
			break
		}
	}
	return m
}

func (m Model) fecharDetalhe() Model {
	m.aberto, m.rascunho, m.rolagem, m.opcao, m.diagrama, m.armado, m.focoTexto = nil, nil, 0, 0, false, 0, false
	return m
}

func (m Model) urlDetalhe() string {
	switch a := m.aberto; {
	case len(a.Links) > 0:
		return a.Links[0].URL
	case a.Relatorio != "":
		return m.naRaiz(a.Relatorio)
	}
	return ""
}

// pedeConfirmacao marca as respostas de efeito grande: liberar o merge e recusar uma ação.
func pedeConfirmacao(a *dados.Aberto, o dados.Opcao) bool {
	return a.Merge != nil && (o.ID == "liberar" || o.Label == "Liberar merge") ||
		a.Kind == "action" && (o.ID == "nao" || o.Label == "Não vou fazer")
}

// escolher age na opção i; a que pede confirmação só responde se já estava armada.
func (m Model) escolher(i, armado int) (Model, tea.Cmd) {
	a := m.aberto
	if i < 0 || i >= len(a.Opcoes) {
		return m, nil
	}
	m.opcao = i
	o := a.Opcoes[i]
	if pedeConfirmacao(a, o) && armado != i+1 {
		m.armado = i + 1
		return m, nil
	}
	if o.Livre {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = "diga o quê"
		ti.TextStyle = st(T.Base, T.Texto)
		ti.PlaceholderStyle = st(T.Base, T.Dim)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acao)
		ti.Focus()
		m.rascunho = &ti
		return m, textinput.Blink
	}
	return m.responder(o.Label)
}

func (m Model) responder(texto string) (Model, tea.Cmd) {
	if m.acoes.Responder == nil {
		m.status = "não respondi: sem como responder nesta máquina"
		return m, nil
	}
	f, uid, chave := m.acoes.Responder, m.aberto.Uid, m.aberto.Chave
	if m.enviando == chave {
		return m, nil
	}
	texto = strings.Join(strings.Fields(texto), " ")
	m.status, m.enviando = "enviando…", chave
	return m, func() tea.Msg {
		e, err := f(uid, texto)
		return respondeuMsg{Entrega: e, error: err, chave: chave}
	}
}

func (m Model) respondeu(r respondeuMsg) Model {
	m.status = statusDe(r)
	if m.enviando == r.chave {
		m.enviando = ""
	}
	if r.error != nil {
		return m
	}
	if m.aberto != nil && m.aberto.Chave == r.chave {
		m = m.fecharDetalhe()
	}
	var resto []dados.Linha
	for _, l := range m.estado.Esperando {
		if l.Chave != r.chave {
			resto = append(resto, l)
		}
	}
	m.estado.Esperando = resto
	return m.mover(0)
}

func (m Model) alturaCorpo() int {
	topo, _ := m.linhasTopo(m.larguraEfetiva())
	return max(1, m.alturaEfetiva()-1-len(topo))
}

func (m Model) rolar(d int) Model {
	_, conteudo, area := m.layoutDetalhe(m.larguraEfetiva(), m.alturaCorpo())
	m.rolagem = min(max(0, m.rolagem+d), max(0, len(conteudo)-area))
	return m
}

func (m Model) teclaDetalhe(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	a := m.aberto
	if m.rascunho != nil {
		switch msg.String() {
		case "esc":
			m.rascunho = nil
			return m, nil
		case "enter":
			txt := strings.TrimSpace(m.rascunho.Value())
			if txt == "" {
				m.status = "escreva o quê; vazio não envia"
				return m, nil
			}
			return m.responder(a.Opcoes[m.opcao].Label + ": " + txt)
		case "ctrl+c":
			return m, tea.Quit
		}
		ti, cmd := m.rascunho.Update(msg)
		m.rascunho = &ti
		return m, cmd
	}
	pagina := max(1, m.alturaCorpo()-len(m.fixoDetalhe(m.larguraEfetiva()))-2)
	armado := m.armado
	m.armado = 0
	k := msg.String()
	switch k {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.ajuda = true
	case "esc":
		if armado != 0 {
			return m, nil
		}
		return m.fecharDetalhe(), nil
	case "tab":
		m.focoTexto = !m.focoTexto
	case "up", "k":
		if m.focoTexto {
			return m.rolar(-1), nil
		}
		m.opcao = max(0, m.opcao-1)
	case "down", "j":
		if m.focoTexto {
			return m.rolar(1), nil
		}
		m.opcao = min(len(a.Opcoes)-1, m.opcao+1)
	case "pgdown", " ":
		return m.rolar(pagina), nil
	case "pgup":
		return m.rolar(-pagina), nil
	case "d":
		if a.Diagrama != "" {
			m.diagrama = !m.diagrama
		}
	case "o":
		return m.abrirURL(m.urlDetalhe()), nil
	case "g":
		return m.irAba(), nil
	case "enter":
		if m.focoTexto {
			return m, nil
		}
		return m.escolher(m.opcao, armado)
	}
	if len(k) == 1 && k[0] >= '1' && k[0] <= '9' {
		return m.escolher(int(k[0]-'1'), armado)
	}
	return m, nil
}

// ---- desenho ----

func tx(fg lipgloss.Color, s string) string { return st(T.Base, fg).Render(s) }

// quebra divide por palavras (e corta palavra maior que a largura) sem estilo.
func quebra(s string, w int) []string {
	if w < 1 {
		w = 1
	}
	var out []string
	for _, l := range strings.Split(lipgloss.NewStyle().Width(w).Render(s), "\n") {
		out = append(out, strings.TrimRight(l, " "))
	}
	return out
}

// paragrafo quebra o texto; "- item" vira "• item" com recuo pendurado.
func paragrafo(s string, w int, fg lipgloss.Color) []string {
	pre, cont := "", ""
	if strings.HasPrefix(s, "- ") {
		s, pre, cont = strings.TrimPrefix(s, "- "), "• ", "  "
	}
	var out []string
	for i, l := range quebra(s, w-lipgloss.Width(pre)) {
		p := cont
		if i == 0 {
			p = pre
		}
		out = append(out, tx(T.Dim, p)+tx(fg, l))
	}
	return out
}

func rotuloSecao(s string) string { return st(T.Base, T.Dim).Bold(true).Render(strings.ToUpper(s)) }

// caixa desenha um bloco de borda arredondada com largura total w.
func caixa(linhas []string, w int, borda lipgloss.Color) []string {
	inner := w - 4
	var b strings.Builder
	for i, l := range linhas {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(encaixa(l, inner, T.Base))
	}
	s := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borda).BorderBackground(T.Base).
		Background(T.Base).Padding(0, 1).Render(b.String())
	return strings.Split(s, "\n")
}

// par desenha "chave  valor" com o valor quebrado e recuado sob ele.
func par(k, v string, kw, w int, fg lipgloss.Color) []string {
	var out []string
	for i, l := range quebra(v, w-kw) {
		c := strings.Repeat(" ", kw)
		if i == 0 {
			c = Col(k, kw)
		}
		out = append(out, tx(T.Dim, c)+tx(fg, l))
	}
	return out
}

type bloco struct{ ls []string }

func (b *bloco) add(l ...string) { b.ls = append(b.ls, l...) }
func (b *bloco) branco()         { b.ls = append(b.ls, "") }

func (b *bloco) secao(titulo string, ls []string) {
	if len(ls) == 0 {
		return
	}
	b.add(rotuloSecao(titulo))
	b.add(ls...)
	b.branco()
}

var nomeKind = map[string]string{"gate": "gate", "action": "ação", "question": "decisão", "report": "relatório"}

func (m Model) cabecalho(a *dados.Aberto, w int) []string {
	tipo := nomeKind[a.Kind]
	if tipo == "" {
		tipo = "decisão"
	}
	cor := T.Espera
	if a.Tipo == "relatorio" {
		tipo, cor = "concluído", T.Ok
		if a.Status == "blocked" {
			tipo, cor = "travado", T.Erro
		}
	}
	if a.Merge != nil {
		tipo = "juntar"
	}
	chips := st(T.Base, cor).Bold(true).Render(strings.ToUpper(tipo))
	if a.Peso != "" {
		cp := T.Dim
		if a.Peso == "alto" {
			cp = T.Espera
		}
		chips += tx(T.Dim, "  ·  ") + tx(cp, "peso "+a.Peso)
	}
	var meta []string
	for _, s := range []string{a.Crew, a.Key, a.Area} {
		if s != "" {
			meta = append(meta, s)
		}
	}
	var out []string
	aba, url := m.atual()
	if b := botoes(T.Base, aba != "", url != ""); b == "" {
		out = append(out, chips)
	} else if lipgloss.Width(chips)+2+lipgloss.Width(b) <= w {
		out = append(out, chips+espacos(w-lipgloss.Width(chips)-lipgloss.Width(b), T.Base)+b)
	} else {
		out = append(out, chips, b)
	}
	for _, l := range quebra(juntar(meta), w) {
		out = append(out, tx(T.Sub, l))
	}
	out = append(out, "")
	for _, l := range quebra(a.Titulo, w) {
		out = append(out, st(T.Base, T.Texto).Bold(true).Render(l))
	}
	return append(out, "")
}

func (m Model) cartaoJuntar(a *dados.Aberto, w int) []string {
	mg := a.Merge
	inner := w - 4
	pr := fmt.Sprintf("%s#%d", mg.Repo, mg.PR)
	var ls []string
	for _, l := range quebra(pr, inner) {
		ls = append(ls, st(T.Base, T.Texto).Bold(true).Render(l))
	}
	ls = append(ls, "")
	kw := 8
	if mg.Base != "" {
		ls = append(ls, par("base", mg.Base, kw, inner, T.Texto)...)
	}
	if mg.Spark != "" {
		ls = append(ls, par("Spark", mg.Spark, kw, inner, T.Texto)...)
	}
	if mg.Checks != "" {
		ls = append(ls, par("checks", mg.Checks, kw, inner, T.Texto)...)
	}
	borda := T.Sel
	for _, t := range m.estado.Travados {
		if t.Repo == mg.Repo && t.Number == mg.PR {
			cor := T.Andamento
			switch dados.TomDoPR(t.Motivo) {
			case "ok":
				cor, borda = T.Ok, T.Ok
			case "aviso":
				cor, borda = T.Espera, T.Espera
			}
			ls = append(ls, par("agora", "● "+t.Motivo, kw, inner, cor)...)
		}
	}
	return caixa(ls, w, borda)
}

func corCompara(status string) lipgloss.Color {
	switch status {
	case "ok":
		return T.Ok
	case "diff":
		return T.Espera
	}
	return T.Dim
}

func (m Model) tabelaCompara(a *dados.Aberto, w int) []string {
	cab := append([]string{}, a.ComparaCab...)
	for len(cab) < 4 {
		cab = append(cab, []string{"Ponto", "Protótipo / pedido", "Hoje", ""}[len(cab)])
	}
	var out []string
	if w < 66 {
		for _, r := range a.Compara {
			l := tx(corCompara(r.Status), "● ") + st(T.Base, T.Texto).Bold(true).Render(r.Item)
			if r.Tag != "" {
				l += tx(corCompara(r.Status), "  "+r.Tag)
			}
			if lipgloss.Width(l) > w {
				out = append(out, paragrafo(r.Item, w, T.Texto)...)
				out = append(out, tx(corCompara(r.Status), "  "+r.Tag))
			} else {
				out = append(out, l)
			}
			for _, kv := range [][2]string{{cab[1], r.A}, {cab[2], r.B}} {
				for _, q := range quebra(kv[0]+": "+kv[1], w-2) {
					out = append(out, "  "+tx(T.Sub, q))
				}
			}
			out = append(out, "")
		}
		return out[:max(0, len(out)-1)]
	}
	tw := lipgloss.Width(cab[3])
	for _, r := range a.Compara {
		tw = max(tw, lipgloss.Width(r.Tag))
	}
	tw = min(max(tw, 4), 14)
	resto := w - tw - 6
	iw := resto * 30 / 100
	cw := (resto - iw) / 2
	bw := resto - iw - cw
	ws := []int{iw, cw, bw, tw}
	linha := func(cels []string, cores []lipgloss.Color, negrito bool) []string {
		cols := make([][]string, 4)
		n := 0
		for i, c := range cels {
			cols[i] = quebra(c, ws[i])
			n = max(n, len(cols[i]))
		}
		var ls []string
		for k := 0; k < n; k++ {
			var s string
			for i := range cels {
				v := ""
				if k < len(cols[i]) {
					v = cols[i][k]
				}
				if i > 0 {
					s += tx(T.Dim, "  ")
				}
				s += st(T.Base, cores[i]).Bold(negrito).Render(Col(v, ws[i]))
			}
			ls = append(ls, s)
		}
		return ls
	}
	out = append(out, linha(cab[:4], []lipgloss.Color{T.Dim, T.Dim, T.Dim, T.Dim}, true)...)
	out = append(out, tx(T.Sel, strings.Repeat("─", w)))
	for _, r := range a.Compara {
		out = append(out, linha([]string{r.Item, r.A, r.B, r.Tag}, []lipgloss.Color{T.Texto, T.Sub, T.Sub, corCompara(r.Status)}, false)...)
	}
	return out
}

func (m Model) conteudoDetalhe(w int) []string {
	a := m.aberto
	cw := w - 2*margem
	var b bloco
	b.add(m.cabecalho(a, cw)...)
	if a.Merge != nil {
		b.add(m.cartaoJuntar(a, cw)...)
		b.branco()
	}
	para := func(ls []string, fg lipgloss.Color) []string {
		var out []string
		for _, l := range ls {
			out = append(out, paragrafo(l, cw, fg)...)
		}
		return out
	}
	if a.Tipo == "relatorio" {
		var res []string
		for _, s := range a.Resumo {
			res = append(res, paragrafo("- "+s, cw, T.Texto)...)
		}
		b.secao("resultado", res)
	} else {
		b.secao("contexto", para(a.Contexto, T.Texto))
	}
	var links []string
	for _, l := range a.Links {
		for i, q := range quebra(l.Label, cw-2) {
			p := "  "
			if i == 0 {
				p = "↗ "
			}
			links = append(links, tx(T.Info, p+q))
		}
		for _, q := range quebra(l.URL, cw-2) {
			links = append(links, tx(T.Dim, "  "+q))
		}
	}
	b.secao("links", links)
	if a.Relatorio != "" {
		var r []string
		for _, q := range quebra(a.Relatorio, cw) {
			r = append(r, tx(T.Sub, q))
		}
		b.secao("relatório completo", r)
	}
	b.secao("entenda em detalhe", para(a.Explicacao, T.Sub))
	if a.Diagrama != "" {
		if !m.diagrama {
			b.add(tx(T.Acao, "◆ ") + tx(T.Sub, "diagrama") + tx(T.Dim, "  (d para ver)"))
			b.branco()
		} else {
			var src []string
			for _, l := range strings.Split(a.Diagrama, "\n") {
				for _, q := range quebra(l, cw-4) {
					src = append(src, tx(T.Sub, q))
				}
			}
			ls := caixa(src, cw, T.Sel)
			for _, q := range quebra(a.LegendaDiagrama, cw) {
				if q != "" {
					ls = append(ls, tx(T.Dim, q))
				}
			}
			b.add(rotuloSecao("diagrama") + tx(T.Dim, "  (d recolhe · fonte Mermaid)"))
			b.add(ls...)
			b.branco()
		}
	}
	if len(a.Compara) > 0 {
		t := a.ComparaTitulo
		if t == "" {
			t = "lado a lado"
		}
		b.secao(t, m.tabelaCompara(a, cw))
	}
	if len(a.Comandos) > 0 {
		var cmds []string
		for _, c := range a.Comandos {
			for i, q := range quebra(c, cw-6) {
				p := "  "
				if i == 0 {
					p = "$ "
				}
				cmds = append(cmds, tx(T.Dim, p)+tx(T.Texto, q))
			}
		}
		b.add(rotuloSecao("comandos para você rodar"))
		b.add(caixa(cmds, cw, T.Sel)...)
		b.branco()
	}
	rec := func(titulo, texto string, borda lipgloss.Color) {
		if texto == "" {
			return
		}
		ls := []string{rotuloSecao(titulo)}
		ls = append(ls, paragrafo(texto, cw-4, T.Texto)...)
		b.add(caixa(ls, cw, borda)...)
		b.branco()
	}
	recCrew := a.Recomendacao
	for _, o := range a.Opcoes {
		if o.ID != "" && o.ID == recCrew {
			recCrew = o.Label
		}
	}
	rec("recomendação do fantasminha", recCrew, T.Sel)
	rec("recomendação do coordenador", a.Coordenador, T.Acao)
	if a.Destrava != "" {
		b.secao("o que destrava", paragrafo(a.Destrava, cw, T.Texto))
	}
	if a.PreviaDe != "" {
		var pv []string
		for _, l := range a.Previa {
			for _, q := range quebra(l, cw-4) {
				pv = append(pv, tx(T.Sub, q))
			}
		}
		var tit []string
		for _, q := range quebra(a.PreviaDe, cw-lipgloss.Width("PRÉVIA  ")) {
			tit = append(tit, q)
		}
		b.add(rotuloSecao("prévia") + tx(T.Dim, "  "+tit[0]))
		for _, q := range tit[1:] {
			b.add(tx(T.Dim, "        "+q))
		}
		b.add(caixa(pv, cw, T.Sel)...)
		b.branco()
	}
	var outros []string
	for _, d := range a.Documentos {
		if d != a.PreviaDe {
			outros = append(outros, paragrafo("- "+d, cw, T.Sub)...)
		}
	}
	b.secao("documentos citados", outros)
	out := make([]string, len(b.ls))
	for i, l := range b.ls {
		out[i] = encaixa(espacos(margem, T.Base)+l, w, T.Base)
	}
	return out
}

// fixoDetalhe são as opções, o campo de texto e a dica: ficam presos embaixo.
func (m Model) fixoDetalhe(w int) []string {
	ls, _ := m.fixoComDono(w)
	return ls
}

// fixoComDono devolve o fixo e, por linha, a opção dona (ou -1), para o clique.
func (m Model) fixoComDono(w int) ([]string, []int) {
	a := m.aberto
	cw := w - 2*margem
	sep := tx(T.Sel, strings.Repeat("─", cw))
	if m.focoTexto {
		r := "↑↓ rolam o texto · tab volta às opções "
		sep = tx(T.Acao, r) + tx(T.Sel, strings.Repeat("─", max(0, cw-lipgloss.Width(r))))
	}
	ls := []string{sep}
	dono := []int{-1}
	lw := 0
	for _, o := range a.Opcoes {
		lw = max(lw, lipgloss.Width(o.Label)+lipgloss.Width(marcaPadrao(o)))
	}
	estreito := w < 70 || lw > cw*2/5
	for i, o := range a.Opcoes {
		sel := i == m.opcao
		armada := m.armado == i+1
		bg, fgL, fgE, fgM, fgN := T.Base, T.Texto, T.Dim, T.Dim, T.Dim
		seta := "  "
		switch {
		case armada:
			bg, fgL, fgE, fgM, fgN, seta = T.Espera, T.Base, T.Mantle, T.Mantle, T.Mantle, "! "
		case sel && m.focoTexto:
			bg, fgM, fgN, seta = T.Sel, T.Sub, T.Sub, "› "
		case sel:
			bg, fgL, fgE, fgM, fgN, seta = T.Acao, T.Base, T.Mantle, T.Mantle, T.Mantle, "› "
		}
		num := "  "
		if i < 9 {
			num = fmt.Sprintf("%d ", i+1)
		}
		pre := st(bg, fgL).Bold(true).Render(seta) + st(bg, fgN).Render(num)
		cab := pre + st(bg, fgL).Bold(true).Render(o.Label) + st(bg, fgM).Render(marcaPadrao(o))
		var linhas []string
		if estreito || o.Effect == "" {
			for j, q := range quebra(o.Label+marcaPadrao(o), cw-4) {
				if j == 0 && lipgloss.Width(cab) <= cw {
					linhas = append(linhas, cab)
					break
				}
				if j == 0 {
					linhas = append(linhas, pre+st(bg, fgL).Bold(true).Render(q))
					continue
				}
				linhas = append(linhas, st(bg, fgL).Render("    "+q))
			}
			for _, q := range quebra(o.Effect, cw-6) {
				if q != "" {
					linhas = append(linhas, st(bg, fgE).Render("      "+q))
				}
			}
		} else {
			cab += espacos(2+lw-lipgloss.Width(o.Label+marcaPadrao(o)), bg)
			ew := cw - 6 - lw
			for j, q := range quebra(o.Effect, ew) {
				if j == 0 {
					linhas = append(linhas, cab+st(bg, fgE).Render(q))
				} else {
					linhas = append(linhas, espacos(6+lw, bg)+st(bg, fgE).Render(q))
				}
			}
		}
		if armada {
			for _, q := range quebra("Enter de novo para confirmar · Esc cancela", cw-4) {
				linhas = append(linhas, espacos(4, bg)+st(bg, T.Base).Bold(true).Render(q))
			}
		}
		for _, l := range linhas {
			ls = append(ls, encaixa(l, cw, bg))
			dono = append(dono, i)
		}
	}
	if m.rascunho != nil {
		o := a.Opcoes[m.opcao]
		pre := st(T.Base, T.Acao).Bold(true).Render(o.Label + ": ")
		ti := *m.rascunho
		ti.Width = max(1, cw-lipgloss.Width(pre)-2)
		ls = append(ls, "", pre+ti.View())
	}
	ls = append(ls, "", m.dica(cw))
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = encaixa(espacos(margem, T.Base)+l, w, T.Base)
	}
	for len(dono) < len(out) {
		dono = append(dono, -1)
	}
	return out, dono
}

func marcaPadrao(o dados.Opcao) string {
	if o.Padrao {
		return "  padrão"
	}
	return ""
}

func (m Model) dica(w int) string {
	var vs []string
	switch {
	case m.rascunho != nil:
		vs = []string{"enter envia · esc cancela", "enter · esc"}
	case m.armado != 0:
		vs = []string{"enter confirma · qualquer outra tecla cancela", "enter confirma · esc cancela"}
	default:
		d := ""
		if m.aberto.Diagrama != "" {
			d = " · d diagrama"
		}
		vs = []string{
			"↑↓ escolhe · 1-9 ou enter responde · tab foco · pgdn rola" + d + " · esc volta · ? ajuda",
			"↑↓ escolhe · 1-9 responde · tab foco" + d + " · esc volta",
			"↑↓ · 1-9 · tab · esc · ? ajuda",
			"↑↓ · enter · esc",
		}
	}
	for _, v := range vs {
		if lipgloss.Width(v) <= w {
			return tx(T.Dim, v)
		}
	}
	return tx(T.Dim, Corte(vs[len(vs)-1], w))
}

// layoutDetalhe devolve o fixo, o conteúdo inteiro e quantas linhas dele cabem.
func (m Model) layoutDetalhe(w, h int) (fixo, conteudo []string, area int) {
	fixo = m.fixoDetalhe(w)
	conteudo = m.conteudoDetalhe(w)
	return fixo, conteudo, h - len(fixo)
}

func (m Model) detalheLinhas(w, h int) ([]string, []int) {
	fixo, dFixo := m.fixoComDono(w)
	conteudo := m.conteudoDetalhe(w)
	area := h - len(fixo)
	if area < 1 {
		k := max(0, len(fixo)-h)
		return fixo[k:], dFixo[k:]
	}
	rol := min(max(0, m.rolagem), max(0, len(conteudo)-area))
	vis := append([]string{}, conteudo[rol:min(len(conteudo), rol+area)]...)
	for len(vis) < area {
		vis = append(vis, espacos(w, T.Base))
	}
	if rol+area < len(conteudo) {
		vis[area-1] = encaixa(espacos(margem, T.Base)+tx(T.Acao, "↓ ")+tx(T.Dim, fmt.Sprintf("mais %d linhas (pgdn)", len(conteudo)-rol-area)), w, T.Base)
	}
	if rol > 0 && area > 1 {
		vis[0] = encaixa(espacos(margem, T.Base)+tx(T.Acao, "↑ ")+tx(T.Dim, "acima (pgup)"), w, T.Base)
	}
	dono := make([]int, area, area+len(dFixo))
	for i := range dono {
		dono[i] = -1
	}
	return append(vis, fixo...), append(dono, dFixo...)
}

// marcarVisto tira o "ler" do feed; falha do cache só vale para esta sessão.
func (m Model) marcarVisto(pasta string) Model {
	var k string
	if m.acoes.MarcarVisto != nil {
		k, _ = m.acoes.MarcarVisto(pasta)
	} else {
		k = dados.ChaveVisto(pasta)
	}
	if k == "" {
		return m
	}
	m.vistos = juntaVistos(m.vistos, map[string]bool{k: true})
	m.estado.Vistos = juntaVistos(m.estado.Vistos, m.vistos)
	return m
}

func juntaVistos(a, b map[string]bool) map[string]bool {
	if len(b) == 0 {
		return a
	}
	v := make(map[string]bool, len(a)+len(b))
	for x := range a {
		v[x] = true
	}
	for x := range b {
		v[x] = true
	}
	return v
}
