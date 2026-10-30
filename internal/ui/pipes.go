package ui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OpsPipes são os comandos de pipeline do ddb que a gaveta pipes usa.
type OpsPipes interface {
	Listar() ([]dados.Pipeline, error)
	Novo(pipeline string, entradas [][2]string, titulo string) (string, error)
	Etapa(run int, etapa, acao string) error
	Validar(alvo string) (string, error)
}

const MsgPipesDesligado = "pipes: ligue FANTASMAS_FONTE=ddb"

type editouMsg struct {
	arquivo string
	err     error
}

// formPipe é o formulário do n: escolhe o pipeline e preenche as entradas.
type formPipe struct {
	pipelines []dados.Pipeline
	escolha   int
	passo     int
	campos    []textinput.Model
	nomes     []string
	foco      int
}

var glifoEtapa = map[string]string{
	"done": "✓", "running": "●", "needs_input": "▲", "failed": "✗", "pending": "○", "ready": "○", "skipped": "–",
}

var nomeEtapa = map[string]string{
	"done": "feita", "running": "rodando", "needs_input": "espera", "failed": "falhou", "pending": "na fila",
	"ready": "pronta", "skipped": "pulada",
}

func corEtapa(s string) lipgloss.Color {
	switch s {
	case "done":
		return T.Ok
	case "running":
		return T.Andamento
	case "needs_input":
		return T.Espera
	case "failed":
		return T.Erro
	case "ready":
		return T.Acao
	}
	return T.Dim
}

func glifoRun(r dados.Run) (string, lipgloss.Color) {
	switch {
	case r.EsperaVoce > 0:
		return "▲", T.Espera
	case r.Status == "done":
		return "✓", T.Ok
	case r.Status == "failed":
		return "✗", T.Erro
	case r.Status == "running":
		return "●", T.Andamento
	}
	return "◇", T.Acao
}

func idade(min int) string {
	switch {
	case min < 60:
		return fmt.Sprintf("%d min", min)
	case min < 48*60:
		return fmt.Sprintf("%dh", min/60)
	}
	return fmt.Sprintf("%dd", min/(24*60))
}

func nomeRun(r dados.Run) string {
	if r.Title != "" {
		return r.Title
	}
	if r.SlugBase != "" {
		return r.SlugBase
	}
	return fmt.Sprintf("run %d", r.ID)
}

// chips desenha as etapas como pílulas; curto mostra só o glifo.
func chips(bg lipgloss.Color, es []dados.Etapa, curto bool) string {
	var b strings.Builder
	for i, e := range es {
		g := glifoEtapa[e.Status]
		if g == "" {
			g = "○"
		}
		cor := corEtapa(e.Status)
		if curto {
			if i > 0 {
				b.WriteString(st(bg, T.Dim).Render(" "))
			}
			b.WriteString(st(bg, cor).Bold(true).Render(g))
			continue
		}
		if i > 0 {
			b.WriteString(st(bg, T.Borda).Render(" › "))
		}
		b.WriteString(st(T.Chip, cor).Bold(true).Render(" "+g) + st(T.Chip, T.ChipTexto).Render(" "+e.ID+" "))
	}
	return b.String()
}

func larguraChips(es []dados.Etapa) int {
	return lipgloss.Width(chips(T.Base, es, false))
}

func gavetaPipes(c *corpo, e dados.Estado) {
	if !e.PipesLigado {
		c.secao("pipes", "")
		c.branco()
		c.vazio(MsgPipesDesligado)
		return
	}
	espera := 0
	for _, r := range e.Runs {
		espera += r.EsperaVoce
	}
	extra := fmt.Sprintf("%d runs", len(e.Runs))
	if espera > 0 {
		extra += fmt.Sprintf(" · %d esperam você", espera)
	}
	if c.largura >= 70 {
		extra += " · n lança · e edita"
	}
	c.secao("pipes", extra)
	if len(e.Runs) == 0 {
		c.branco()
		c.vazio("nenhum run · n lança um pipeline")
		return
	}
	c.branco()
	for _, r := range e.Runs {
		if !c.passa(r.Pipeline, nomeRun(r), r.SlugBase, r.Status) {
			continue
		}
		c.linhaRun(r)
	}
}

func (c *corpo) linhaRun(r dados.Run) {
	w := c.largura - 2
	g, cg := glifoRun(r)
	nome := nomeRun(r)
	quando := idade(r.IdadeMin)
	esp := ""
	if r.EsperaVoce > 0 {
		esp = fmt.Sprintf("espera você: %d", r.EsperaVoce)
	}
	a := alvo{pipe: r.ID}
	dir := func(bg lipgloss.Color) string {
		s := st(bg, T.Dim).Render(quando)
		if esp != "" {
			s += st(bg, T.Dim).Render("  ") + st(bg, T.Espera).Bold(true).Render(esp)
		}
		return s
	}
	wc := larguraChips(r.Etapas)
	if w >= 94 {
		wp := 24
		wDir := lipgloss.Width(quando) + 2 + len([]rune(esp)) + 2
		wNome := min(36, max(10, w-2-wp-wc-wDir-2))
		c.item(a, func(bg lipgloss.Color) string {
			esq := st(bg, cg).Bold(true).Render(g+" ") + st(bg, T.Info).Render(Col(r.Pipeline, wp)) +
				st(bg, T.Texto).Bold(true).Render(Col(nome, wNome)) + espacos(2, bg) + chips(bg, r.Etapas, false)
			d := dir(bg)
			return esq + espacos(max(1, w-lipgloss.Width(esq)-lipgloss.Width(d)), bg) + d
		})
		c.branco()
		return
	}
	curto := wc > w-4
	c.item(a, func(bg lipgloss.Color) string {
		esq := st(bg, cg).Bold(true).Render(g+" ") + st(bg, T.Texto).Bold(true).Render(Corte(nome, max(4, w-2-14)))
		d := ""
		if esp != "" && w >= 46 {
			d = st(bg, T.Espera).Bold(true).Render(esp)
		} else if r.EsperaVoce > 0 {
			d = st(bg, T.Espera).Bold(true).Render(fmt.Sprintf("▲%d", r.EsperaVoce))
		}
		return esq + espacos(max(1, w-lipgloss.Width(esq)-lipgloss.Width(d)), bg) + d
	}, func(bg lipgloss.Color) string {
		s := espacos(2, bg)
		if w >= 46 {
			s += st(bg, T.Info).Render(Corte(r.Pipeline, 22)) + espacos(2, bg)
		}
		s += chips(bg, r.Etapas, curto) + espacos(2, bg) + st(bg, T.Dim).Render(quando)
		return s
	})
	c.branco()
}

// ---- detalhe ----

func (m Model) runAberto() dados.Run {
	for _, r := range m.estado.Runs {
		if r.ID == m.pipe {
			return r
		}
	}
	return dados.Run{}
}

func (m Model) abrirPipe(id int) Model {
	m.pipe, m.pipeSel, m.armadoPipe, m.rolagem = id, 0, "", 0
	for i, e := range m.runAberto().Etapas {
		if e.Status != "done" && e.Status != "skipped" {
			m.pipeSel = i
			break
		}
	}
	return m
}

func (m Model) fecharPipe() Model {
	m.pipe, m.armadoPipe, m.form = 0, "", nil
	return m
}

var reasonPT = []struct{ pre, txt string }{
	{"gate:", "aguarda o gate "}, {"question:", "perguntas abertas: "}, {"precondition:", "pré-condição falhou: "},
	{"spawn:", "spawn falhou: "},
}

// motivo traduz o reason do needs_input; o valor cru vai junto, em dim.
func motivo(r string) string {
	switch r {
	case "gone":
		return "o agente sumiu sem relatório"
	case "veto":
		return "veto entregue; aguardando o fantasminha"
	case "falhou":
		return "falhou e esgotou o retry"
	}
	for _, x := range reasonPT {
		if strings.HasPrefix(r, x.pre) {
			return x.txt + strings.TrimPrefix(r, x.pre)
		}
	}
	return r
}

type linhaPipe struct {
	f    func(bg lipgloss.Color) string
	dono int
}

func (m Model) conteudoPipe(w int) []linhaPipe {
	r := m.runAberto()
	var ls []linhaPipe
	add := func(dono int, f func(bg lipgloss.Color) string) { ls = append(ls, linhaPipe{f, dono}) }
	cw := w - 2*margem
	if r.ID == 0 {
		add(-1, func(bg lipgloss.Color) string { return st(bg, T.Dim).Render("o run sumiu do ddb · esc volta") })
		return ls
	}
	g, cg := glifoRun(r)
	add(-1, func(bg lipgloss.Color) string {
		return st(bg, cg).Bold(true).Render(g+" ") + st(bg, T.Texto).Bold(true).Render(Corte(nomeRun(r), max(4, cw-2)))
	})
	add(-1, func(bg lipgloss.Color) string {
		s := st(bg, T.Info).Render(fmt.Sprintf("run %d · %s", r.ID, r.Pipeline)) + st(bg, T.Dim).Render("  ·  ") +
			st(bg, corEtapa(map[string]string{"done": "done", "failed": "failed", "running": "running"}[r.Status])).Render(r.Status) +
			st(bg, T.Dim).Render("  ·  "+idade(r.IdadeMin))
		if r.Coordinator != "" {
			s += st(bg, T.Dim).Render("  ·  gengar " + r.Coordinator)
		}
		return s
	})
	if len(r.Inputs) > 0 {
		ks := make([]string, 0, len(r.Inputs))
		for k := range r.Inputs {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		var ps []string
		for _, k := range ks {
			ps = append(ps, k+"="+r.Inputs[k])
		}
		ent := strings.Join(ps, "  ")
		add(-1, func(bg lipgloss.Color) string {
			return st(bg, T.Dim).Render("entradas  ") + st(bg, T.Sub).Render(Corte(ent, max(4, cw-10)))
		})
	}
	add(-1, func(bg lipgloss.Color) string { return "" })
	add(-1, func(bg lipgloss.Color) string { return st(bg, T.Dim).Bold(true).Render("ETAPAS") })
	for i, e := range r.Etapas {
		i, e := i, e
		cor := corEtapa(e.Status)
		gl := glifoEtapa[e.Status]
		sel := i == m.pipeSel
		add(i, func(bg lipgloss.Color) string {
			bar := st(bg, T.Base).Render(" ")
			if sel {
				bar = st(bg, T.Acao).Bold(true).Render("▌")
			}
			s := bar + st(bg, cor).Bold(true).Render(" "+gl+" ") + st(bg, T.Texto).Bold(true).Render(e.ID) +
				st(bg, cor).Render("  "+nomeEtapa[e.Status])
			det := e.Role
			if e.Skill != "" {
				det += " · " + e.Skill
			}
			if e.Attempt > 0 {
				det += fmt.Sprintf(" · tentativa %d", e.Attempt+1)
			}
			return s + st(bg, T.Dim).Render("  "+det)
		})
		ind := func(bg lipgloss.Color) string { return st(bg, T.Base).Render(" ") + espacos(4, bg) }
		dep := "início"
		if len(e.After) > 0 {
			var as []string
			for _, x := range e.After {
				as = append(as, strings.TrimSuffix(x, ".done"))
			}
			dep = "após " + strings.Join(as, ", ")
		}
		var gs []string
		for _, x := range e.Gates {
			gs = append(gs, x.Item+" ("+x.By+")")
		}
		meta := dep
		if len(gs) > 0 {
			meta += " · gates " + strings.Join(gs, ", ")
		}
		add(i, func(bg lipgloss.Color) string {
			return ind(bg) + st(bg, T.Dim).Render(Corte(meta+" · "+e.Slug, max(4, cw-6)))
		})
		if e.Status == "needs_input" && e.Reason != "" {
			txt, cru := motivo(e.Reason), e.Reason
			add(i, func(bg lipgloss.Color) string {
				s := ind(bg) + st(bg, T.Espera).Bold(true).Render("▲ "+txt)
				return s + st(bg, T.Dim).Render(Corte("  "+cru, max(0, cw-lipgloss.Width(s)-2)))
			})
		}
		if e.Agent != "" || e.Tab != "" {
			ag := e.Agent
			if e.Gone {
				ag += " (gone)"
			}
			tab := e.Tab
			add(i, func(bg lipgloss.Color) string {
				s := ind(bg) + st(bg, T.Sub).Render(Corte("agente "+ag, max(4, cw-22)))
				if tab != "" {
					s += espacos(2, bg) + st(bg, T.Acao).Bold(sel).Render(BotaoAba)
				}
				return s
			})
		}
		if p := m.estado.Proximo[e.Slug]; p != "" {
			add(i, func(bg lipgloss.Color) string {
				return ind(bg) + st(bg, T.Info).Render("» próximo: ") + st(bg, T.Sub).Render(Corte(p, max(4, cw-17)))
			})
		}
		for _, u := range dados.LinksDoReport(e.Report) {
			u := u
			add(i, func(bg lipgloss.Color) string {
				return ind(bg) + st(bg, T.Info).Render("↗ "+Corte(u, max(4, cw-8)))
			})
		}
		add(-1, func(bg lipgloss.Color) string { return "" })
	}
	return ls
}

func (m Model) rodapePipe(w int) []string {
	cw := w - 2*margem
	k := func(t, d string) string {
		return st(T.Base, T.Acao).Bold(true).Render(t) + st(T.Base, T.Dim).Render(" "+d+"  ")
	}
	dica := k("↑↓", "etapa") + k("A", "avança") + k("S", "pula") + k("T", "tenta de novo") + k("g", "aba") + k("esc", "volta")
	if cw < 70 {
		dica = k("A", "avança") + k("S", "pula") + k("T", "retry") + k("esc", "")
	}
	out := []string{tx(T.Sel, strings.Repeat("─", max(1, cw)))}
	if m.armadoPipe != "" {
		out = append(out, st(T.Base, T.Espera).Bold(true).Render("▲ "+Corte(m.status, max(4, cw-2))))
	}
	return append(out, dica)
}

func (m Model) pipeLinhas(w, h int) ([]string, []int) {
	pad := func(s string, bg lipgloss.Color) string {
		return encaixa(espacos(margem, bg)+s, w, bg)
	}
	cont := m.conteudoPipe(w)
	rod := m.rodapePipe(w)
	area := max(1, h-len(rod))
	ini, fim := -1, -1
	for i, l := range cont {
		if l.dono == m.pipeSel {
			if ini < 0 {
				ini = i
			}
			fim = i
		}
	}
	desl := 0
	if fim >= area {
		desl = fim - area + 1
	}
	if ini >= 0 && ini < desl {
		desl = ini
	}
	var out []string
	var dono []int
	for i := desl; i < len(cont) && len(out) < area; i++ {
		bg := T.Base
		if cont[i].dono >= 0 && cont[i].dono == m.pipeSel {
			bg = T.Surface
		}
		out = append(out, pad(cont[i].f(bg), bg))
		dono = append(dono, cont[i].dono)
	}
	for len(out) < area {
		out = append(out, espacos(w, T.Base))
		dono = append(dono, -1)
	}
	for _, r := range rod {
		out = append(out, pad(r, T.Base))
		dono = append(dono, -1)
	}
	if len(out) > h {
		out, dono = out[len(out)-h:], dono[len(dono)-h:]
	}
	return out, dono
}

func (m Model) teclaPipe(k string) (tea.Model, tea.Cmd) {
	armado := m.armadoPipe
	m.armadoPipe = ""
	r := m.runAberto()
	switch k {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "left", "h":
		m.status = ""
		return m.fecharPipe(), nil
	case "up", "k":
		m.pipeSel = max(0, m.pipeSel-1)
	case "down", "j":
		m.pipeSel = min(max(0, len(r.Etapas)-1), m.pipeSel+1)
	case "g":
		return m.irAba(), nil
	case "o":
		if m.pipeSel < len(r.Etapas) {
			if ls := dados.LinksDoReport(r.Etapas[m.pipeSel].Report); len(ls) > 0 {
				return m.abrirURL(ls[0]), nil
			}
		}
	case "r":
		return m.coletar()
	case "A", "S", "T":
		return m.acaoEtapa(k, armado)
	}
	return m, nil
}

var acaoDaTecla = map[string]string{"A": "advance", "S": "skip", "T": "retry"}
var verboDaTecla = map[string]string{"A": "avançar", "S": "pular", "T": "tentar de novo"}

// acaoEtapa: a primeira tecla arma, a segunda (a mesma, na mesma etapa) executa.
func (m Model) acaoEtapa(k, armado string) (tea.Model, tea.Cmd) {
	r := m.runAberto()
	if m.pipeSel < 0 || m.pipeSel >= len(r.Etapas) {
		return m, nil
	}
	e := r.Etapas[m.pipeSel]
	chave := fmt.Sprintf("%s:%d:%s", k, r.ID, e.ID)
	if armado != chave {
		m.armadoPipe = chave
		m.status = k + " de novo confirma " + verboDaTecla[k] + " " + e.ID
		return m, nil
	}
	if m.acoes.Pipes == nil {
		return m.mostrarToast("sem ddb nesta máquina", T.Erro)
	}
	m.status = ""
	if err := m.acoes.Pipes.Etapa(r.ID, e.ID, acaoDaTecla[k]); err != nil {
		return m.mostrarToast("stage "+acaoDaTecla[k]+": "+err.Error(), T.Erro)
	}
	m2, toast := m.mostrarToast(fmt.Sprintf("run %d · %s: %s", r.ID, e.ID, acaoDaTecla[k]), T.Ok)
	m3, col := m2.coletar()
	return m3, tea.Batch(toast, col)
}

// ---- n: lançar ----

func (m Model) abrirForm() (tea.Model, tea.Cmd) {
	if !m.estado.PipesLigado {
		return m.mostrarToast(MsgPipesDesligado, T.Espera)
	}
	if m.acoes.Pipes == nil {
		return m.mostrarToast("sem ddb nesta máquina", T.Erro)
	}
	ps, err := m.acoes.Pipes.Listar()
	if err != nil {
		return m.mostrarToast("pipeline list: "+err.Error(), T.Erro)
	}
	if len(ps) == 0 {
		return m.mostrarToast("nenhum pipeline em shared/delegate/pipelines", T.Espera)
	}
	m.form = &formPipe{pipelines: ps}
	return m, nil
}

func novoCampo(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = placeholder
	ti.TextStyle = st(T.Surface, T.Texto)
	ti.PlaceholderStyle = st(T.Surface, T.Dim)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acao)
	ti.Cursor.TextStyle = st(T.Surface, T.Texto)
	return ti
}

func (f *formPipe) focar(i int) {
	n := len(f.campos)
	f.foco = ((i % n) + n) % n
	for j := range f.campos {
		if j == f.foco {
			f.campos[j].Focus()
		} else {
			f.campos[j].Blur()
		}
	}
}

func (m Model) teclaForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := *m.form
	k := msg.String()
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	if k == "esc" {
		m.form, m.status = nil, ""
		return m, nil
	}
	if f.passo == 0 {
		switch k {
		case "up", "k":
			f.escolha = max(0, f.escolha-1)
		case "down", "j":
			f.escolha = min(len(f.pipelines)-1, f.escolha+1)
		case "enter":
			p := f.pipelines[f.escolha]
			if len(p.Erros) > 0 {
				m.status = p.Name + " não valida; e edita o YAML"
				return m, nil
			}
			f.passo, f.nomes, f.campos = 1, nil, nil
			for _, in := range p.Inputs {
				f.nomes = append(f.nomes, in)
				f.campos = append(f.campos, novoCampo(in))
			}
			f.nomes = append(f.nomes, "título")
			f.campos = append(f.campos, novoCampo("opcional"))
			f.focar(0)
			m.form = &f
			return m, textinput.Blink
		}
		m.form = &f
		return m, nil
	}
	switch k {
	case "tab", "down":
		f.focar(f.foco + 1)
	case "shift+tab", "up":
		f.focar(f.foco - 1)
	case "enter":
		if f.foco < len(f.campos)-1 {
			f.focar(f.foco + 1)
			break
		}
		m.form = &f
		return m.lancar()
	default:
		ti, cmd := f.campos[f.foco].Update(msg)
		f.campos[f.foco] = ti
		m.form = &f
		return m, cmd
	}
	m.form = &f
	return m, nil
}

func (m Model) lancar() (tea.Model, tea.Cmd) {
	f := m.form
	p := f.pipelines[f.escolha]
	var ents [][2]string
	var falta []string
	for i, nome := range f.nomes[:len(f.nomes)-1] {
		v := strings.TrimSpace(f.campos[i].Value())
		if v == "" {
			falta = append(falta, nome)
		}
		ents = append(ents, [2]string{nome, v})
	}
	if len(falta) > 0 {
		m.status = "falta: " + strings.Join(falta, ", ")
		return m, nil
	}
	titulo := f.campos[len(f.campos)-1].Value()
	out, err := m.acoes.Pipes.Novo(p.Name, ents, titulo)
	if err != nil {
		m.status = "run new: " + err.Error()
		return m.mostrarToast(m.status, T.Erro)
	}
	m.form, m.status = nil, ""
	id, _, _ := strings.Cut(out, " ")
	m2, toast := m.mostrarToast("run "+id+" criado em intake ("+p.Name+") · o tick avisa o gengar", T.Ok)
	m3, col := m2.coletar()
	return m3, tea.Batch(toast, col)
}

func (m Model) formLinhas(w, h int) []string {
	f := m.form
	cw := w - 2*margem
	pad := func(s string, bg lipgloss.Color) string { return encaixa(espacos(margem, bg)+s, w, bg) }
	var out []string
	add := func(s string, bg lipgloss.Color) { out = append(out, pad(s, bg)) }
	add(st(T.Base, T.Acao).Render("› ")+st(T.Base, T.Texto).Bold(true).Render("lançar pipeline"), T.Base)
	add("", T.Base)
	if f.passo == 0 {
		add(st(T.Base, T.Dim).Bold(true).Render("PIPELINE"), T.Base)
		for i, p := range f.pipelines {
			bg := T.Base
			seta := "  "
			if i == f.escolha {
				bg, seta = T.Surface, "▸ "
			}
			s := st(bg, T.Acao).Bold(true).Render(seta) + st(bg, T.Texto).Bold(true).Render(Col(p.Name, 24))
			if len(p.Erros) > 0 {
				s += st(bg, T.Erro).Render("✗ " + Corte(p.Erros[0], max(4, cw-30)))
			} else if cw >= 56 {
				s += st(bg, T.Info).Render(Corte(strings.Join(p.Stages, " › "), 28)) +
					st(bg, T.Dim).Render("  entradas: "+Corte(strings.Join(p.Inputs, ", "), max(4, cw-60)))
			}
			add(s, bg)
		}
		add("", T.Base)
		add(st(T.Base, T.Dim).Render("↑↓ escolhe · enter segue · esc cancela"), T.Base)
	} else {
		p := f.pipelines[f.escolha]
		add(st(T.Base, T.Dim).Render("pipeline  ")+st(T.Base, T.Info).Bold(true).Render(p.Name)+
			st(T.Base, T.Dim).Render("  "+Corte(strings.Join(p.Stages, " › "), max(4, cw-30))), T.Base)
		add("", T.Base)
		for i, nome := range f.nomes {
			ti := f.campos[i]
			ti.Width = max(4, cw-18)
			cor := T.Sub
			if i == f.foco {
				cor = T.Acao
			}
			campo := encaixa(ti.View(), max(4, cw-16), T.Surface)
			add(st(T.Base, cor).Bold(i == f.foco).Render(Col(nome, 14))+espacos(1, T.Base)+campo, T.Base)
		}
		add("", T.Base)
		add(st(T.Base, T.Dim).Render("tab próximo campo · enter no último lança · esc cancela"), T.Base)
		if m.status != "" {
			add(st(T.Base, T.Espera).Bold(true).Render("▲ "+Corte(m.status, max(4, cw-2))), T.Base)
		}
		add(st(T.Base, T.Dim).Render("o run nasce em intake; o tick avisa o gengar geral"), T.Base)
	}
	for len(out) < h {
		out = append(out, espacos(w, T.Base))
	}
	return out[:h]
}

// ---- e: editar o YAML ----

func (m Model) editarPipeline() (tea.Model, tea.Cmd) {
	if !m.estado.PipesLigado {
		return m.mostrarToast(MsgPipesDesligado, T.Espera)
	}
	if m.acoes.Pipes == nil || m.acoes.Editar == nil {
		return m.mostrarToast("sem ddb ou sem FANTASMAS_EDITOR", T.Espera)
	}
	ps, err := m.acoes.Pipes.Listar()
	if err != nil || len(ps) == 0 {
		return m.mostrarToast("pipeline list falhou", T.Erro)
	}
	alvoP := ps[0]
	if a, ok := m.selecionado(); ok && a.pipe != 0 {
		for _, r := range m.estado.Runs {
			if r.ID != a.pipe {
				continue
			}
			for _, p := range ps {
				if p.Name == r.Pipeline {
					alvoP = p
				}
			}
		}
	}
	cmd := m.acoes.Editar(alvoP.Arquivo)
	if cmd == nil {
		return m.mostrarToast("ligue FANTASMAS_EDITOR para editar "+alvoP.Name, T.Espera)
	}
	arq := alvoP.Arquivo
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg { return editouMsg{arq, err} })
}

func (m Model) editou(msg editouMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m.mostrarToast("editor: "+msg.err.Error(), T.Erro)
	}
	out, err := m.acoes.Pipes.Validar(msg.arquivo)
	if err != nil {
		return m.mostrarToast("validate: "+err.Error(), T.Erro)
	}
	return m.mostrarToast(out, T.Ok)
}

// ComandoEditor monta o comando do FANTASMAS_EDITOR para o arquivo, ou nil sem editor.
func ComandoEditor(editor, arquivo string) *exec.Cmd {
	ed := strings.Fields(editor)
	if len(ed) == 0 {
		return nil
	}
	return exec.Command(ed[0], append(ed[1:], arquivo)...)
}

// ComPipe abre o detalhe do run (usado pelo --uma-vez --pipe).
func (m Model) ComPipe(id int) Model { return m.abrirPipe(id) }
