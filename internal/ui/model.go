package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Coleta func(ctx context.Context) dados.Estado

type Acoes struct {
	AbrirURL    func(string) error
	FocarAba    func(string) error
	Responder   func(chave, resposta string) (dados.Entrega, error)
	LerArquivo  func(string) ([]byte, error)
	Arquivar    func(chave string) error
	Retomar     func(c dados.Caido) error
	MarcarVisto func(pasta string) (string, error)
	// botões da gaveta máquina
	Daydreaming    func() (string, error)
	ReconectarMCPs func() error
	ProgressoMCP   func() dados.ProgressoMCP
}

type coletouMsg struct{ dados.Estado }
type tickColetaMsg struct{ gen int }
type daydreamingMsg struct {
	texto string
	err   error
}
type tickMCPMsg struct{}
type progressoMCPMsg struct{ dados.ProgressoMCP }
type tickQuadroMsg struct{}
type toastFimMsg struct{ gen int }
type respondeuMsg struct {
	dados.Entrega
	error
	chave string
}

const (
	passoQuadro = 180 * time.Millisecond
	passoMCP    = 2 * time.Second
	duraToast   = 4 * time.Second
)

type Model struct {
	coleta    Coleta
	acoes     Acoes
	intervalo time.Duration

	estado    dados.Estado
	coletando bool
	genColeta int
	quadro    int
	gaveta    int
	sel       int
	status    string
	largura   int
	altura    int
	aberto    *dados.Aberto
	rascunho  *textinput.Model
	opcao     int
	rolagem   int
	diagrama  bool
	enviando  string
	ajuda     bool
	toast     string
	toastCor  lipgloss.Color
	genToast  int
	filtro    *textinput.Model
	busca     string
	armado    int
	focoTexto bool
	img       *Grafico
	imgPos    areaImg
	imgQuadro int
	modoImg   string
	antigas   bool
	rascunhos bool
	armadoArq string
	armadoRet string
	vistos    map[string]bool
	pedindoDD bool
	vigiaMCP  bool
}

// Novo já nasce coletando: o Init dispara a primeira coleta.
func Novo(coleta Coleta, acoes Acoes, intervalo time.Duration) Model {
	return Model{coleta: coleta, acoes: acoes, intervalo: intervalo, coletando: true}
}

// ComEstado devolve o modelo com um estado já coletado (usado pelo --uma-vez).
func (m Model) ComEstado(e dados.Estado) Model {
	m.estado, m.coletando = e, false
	return m
}

// ComGaveta abre o app na gaveta pedida (id desconhecido cai na primeira).
func (m Model) ComGaveta(id string) Model {
	return m.irGaveta(indiceGaveta(id))
}

// ComGrafico liga o Gengar como imagem kitty no lugar do sprite em células.
func (m Model) ComGrafico(g *Grafico) Model {
	m.img = g
	return m
}

// ComModoImagem registra o modo do Gengar para a ajuda.
func (m Model) ComModoImagem(s string) Model {
	m.modoImg = s
	return m
}

// ComStatus abre o app já com uma mensagem na linha de status.
func (m Model) ComStatus(s string) Model {
	m.status = s
	return m
}

func (m Model) comecar() tea.Cmd {
	c := m.coleta
	return func() tea.Msg { return coletouMsg{c(context.Background())} }
}

// passoDe usa o delay do quadro atual do sprite; sem sprite, o passo fixo do mascote nativo.
func passoDe(quadro int) time.Duration {
	if sp := spriteFn(96); sp != nil {
		return sp.Delay(quadro)
	}
	return passoQuadro
}

func tickQuadro(quadro int) tea.Cmd {
	return tea.Tick(passoDe(quadro), func(time.Time) tea.Msg { return tickQuadroMsg{} })
}

func (m Model) Init() tea.Cmd { return tea.Batch(m.comecar(), tickQuadro(m.quadro)) }

func (m Model) coletar() (Model, tea.Cmd) {
	if m.coletando {
		return m, nil
	}
	m.coletando = true
	return m, m.comecar()
}

// GavetaValida diz se id é uma gaveta conhecida.
func GavetaValida(id string) bool {
	for _, g := range Gavetas {
		if g.ID == id {
			return true
		}
	}
	return false
}

// IDsGavetas lista os ids válidos, na ordem das gavetas.
func IDsGavetas() []string {
	ids := make([]string, len(Gavetas))
	for i, g := range Gavetas {
		ids[i] = g.ID
	}
	return ids
}

func indiceGaveta(id string) int {
	for i, g := range Gavetas {
		if g.ID == id {
			return i
		}
	}
	return 0
}

func (m Model) larguraEfetiva() int {
	if m.largura > 0 {
		return m.largura
	}
	return 96
}

func (m Model) alturaEfetiva() int {
	if m.altura > 0 {
		return m.altura
	}
	return 40
}

func (m Model) temBarra() bool {
	return m.aberto == nil && !m.ajuda && (m.filtro != nil || m.busca != "")
}

// alturaLista é a altura do corpo da gaveta, a mesma usada no desenho, para enter e clique baterem.
func (m Model) alturaLista() int {
	h := m.alturaCorpo()
	if m.temBarra() {
		h--
	}
	return max(1, h)
}

func (m Model) alvos() []alvo {
	return m.montarCorpo(m.larguraEfetiva(), m.alturaLista()).alvos
}

// atual devolve a aba do Herdr e a URL do item em foco (detalhe aberto ou seleção).
func (m Model) atual() (aba, url string) {
	if m.aberto != nil {
		return abaDaChave(m.estado, m.aberto.Chave), m.urlDetalhe()
	}
	if a, ok := m.selecionado(); ok {
		return a.aba, a.url
	}
	return "", ""
}

func (m Model) irAba() Model {
	aba, _ := m.atual()
	if aba == "" || m.acoes.FocarAba == nil {
		return m
	}
	return m.erro("não fui à aba: ", m.acoes.FocarAba(aba))
}

func abaDaChave(e dados.Estado, chave string) string {
	for _, t := range e.Tasks {
		if t.Extra != "" && t.Chave != "" && (chave == t.Chave || strings.HasPrefix(chave, t.Chave+"/")) {
			return t.Extra
		}
	}
	return ""
}

func (m Model) teclaFiltro(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.filtro, m.busca, m.sel = nil, "", 0
		return m, nil
	case "enter":
		m.filtro = nil
		return m, nil
	case "up":
		return m.mover(-1), nil
	case "down":
		return m.mover(1), nil
	}
	ti, cmd := m.filtro.Update(msg)
	m.filtro = &ti
	m.busca, m.sel = ti.Value(), 0
	return m.mover(0), cmd
}

func (m Model) abrirFiltro() (Model, tea.Cmd) {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "crew, repo, PR ou título"
	ti.SetValue(m.busca)
	ti.TextStyle = st(T.Surface, T.Texto)
	ti.PlaceholderStyle = st(T.Surface, T.Dim)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acao)
	ti.Cursor.TextStyle = st(T.Surface, T.Texto)
	ti.Focus()
	m.filtro = &ti
	return m, textinput.Blink
}

func (m Model) mover(d int) Model {
	n := len(m.alvos())
	m.sel += d
	if m.sel >= n {
		m.sel = n - 1
	}
	if m.sel < 0 {
		m.sel = 0
	}
	return m
}

func (m Model) irGaveta(i int) Model {
	n := len(Gavetas)
	m.gaveta, m.sel = ((i%n)+n)%n, 0
	return m
}

func (m Model) selecionado() (alvo, bool) {
	as := m.alvos()
	if m.sel < 0 || m.sel >= len(as) {
		return alvo{}, false
	}
	return as[m.sel], true
}

func (m Model) erro(prefixo string, err error) Model {
	if err != nil {
		m.status = prefixo + err.Error()
	}
	return m
}

// PainelURL é o painel do delegate servido pelo painel.py.
const PainelURL = "http://127.0.0.1:7480/"

func (m Model) abrirURL(u string) Model {
	if u == "" || m.acoes.AbrirURL == nil {
		return m
	}
	return m.erro("não abri: ", m.acoes.AbrirURL(u))
}

func (m Model) enter() Model {
	a, ok := m.selecionado()
	switch {
	case !ok:
		return m
	case a.antigas:
		m.antigas = !m.antigas
		return m
	case a.pergunta != "":
		return m.abrirDetalhe(a.pergunta, "pergunta")
	case a.relatorio != "":
		return m.abrirDetalhe(a.relatorio, "relatorio")
	}
	return m.abrirURL(a.url)
}

// enterCmd é o enter com efeito assíncrono: os botões da gaveta máquina.
func (m Model) enterCmd() (tea.Model, tea.Cmd) {
	if a, ok := m.selecionado(); ok {
		switch a.acao {
		case acaoDaydreaming:
			return m.daydreaming()
		case acaoMCP:
			return m.reconectarMCPs()
		}
	}
	return m.enter(), nil
}

const (
	acaoDaydreaming = "daydreaming"
	acaoMCP         = "mcp"
)

// daydreaming roda fora do loop da UI: o spawn pode levar ~2 min.
func (m Model) daydreaming() (tea.Model, tea.Cmd) {
	if m.pedindoDD || m.acoes.Daydreaming == nil {
		return m, nil
	}
	m.pedindoDD = true
	m.status = "pedindo rodada do DayDreaming…"
	f := m.acoes.Daydreaming
	return m, func() tea.Msg { t, err := f(); return daydreamingMsg{t, err} }
}

func (m Model) reconectarMCPs() (tea.Model, tea.Cmd) {
	if m.acoes.ReconectarMCPs == nil {
		return m, nil
	}
	if err := m.acoes.ReconectarMCPs(); err != nil {
		return m.mostrarToast("não reconectei: "+err.Error(), T.Erro)
	}
	m.estado.MCP = dados.ProgressoMCP{Inicio: time.Now(), Rodando: true}
	m, cmd := m.vigiarMCP()
	m2, toast := m.mostrarToast("reconectando MCPs em todas as sessões…", T.Andamento)
	return m2, tea.Batch(cmd, toast)
}

// vigiarMCP liga a releitura do TSV a cada 2 s enquanto a rodada corre.
func (m Model) vigiarMCP() (Model, tea.Cmd) {
	if m.vigiaMCP || m.acoes.ProgressoMCP == nil {
		return m, nil
	}
	m.vigiaMCP = true
	return m, tea.Tick(passoMCP, func(time.Time) tea.Msg { return tickMCPMsg{} })
}

func statusDe(r respondeuMsg) string {
	switch {
	case r.error != nil:
		return "não respondi: " + r.error.Error()
	case !r.Entregue:
		return "gravado; não entregue a " + r.Crew + ": " + r.Motivo
	}
	return "entregue a " + r.Crew + ": " + r.Resposta
}

func (m Model) mostrarToast(s string, cor lipgloss.Color) (Model, tea.Cmd) {
	m.genToast++
	m.toast, m.toastCor = s, cor
	g := m.genToast
	return m, tea.Tick(duraToast, func(time.Time) tea.Msg { return toastFimMsg{g} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	n, cmd := m.atualizar(msg)
	if nm, ok := n.(Model); ok && nm.img != nil {
		_, redim := msg.(tea.WindowSizeMsg)
		n = nm.sincImagem(redim)
	}
	return n, cmd
}

func (m Model) atualizar(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.largura, m.altura = msg.Width, msg.Height
	case coletouMsg:
		m.estado, m.coletando = msg.Estado, false
		m.estado.Vistos = juntaVistos(m.estado.Vistos, m.vistos)
		m = m.mover(0)
		if m.estado.MCP.Rodando {
			var cmd tea.Cmd
			m, cmd = m.vigiarMCP()
			if cmd != nil {
				m.genColeta++
				g := m.genColeta
				return m, tea.Batch(cmd, tea.Tick(m.intervalo, func(time.Time) tea.Msg { return tickColetaMsg{g} }))
			}
		}
		// cada coleta reinicia a cadeia: ticks de gerações antigas morrem ao chegar.
		m.genColeta++
		g := m.genColeta
		return m, tea.Tick(m.intervalo, func(time.Time) tea.Msg { return tickColetaMsg{g} })
	case tickColetaMsg:
		if msg.gen != m.genColeta || m.coletando {
			return m, nil
		}
		return m.coletar()
	case tickQuadroMsg:
		m.quadro++
		return m, tickQuadro(m.quadro)
	case daydreamingMsg:
		m.pedindoDD = false
		if msg.err != nil {
			m.status = "DayDreaming: " + msg.err.Error()
			return m.mostrarToast(m.status, T.Erro)
		}
		m.status = msg.texto
		m, toast := m.mostrarToast(msg.texto, T.Ok)
		m2, col := m.coletar()
		return m2, tea.Batch(toast, col)
	case tickMCPMsg:
		f := m.acoes.ProgressoMCP
		return m, func() tea.Msg { return progressoMCPMsg{f()} }
	case progressoMCPMsg:
		m.estado.MCP = msg.ProgressoMCP
		if msg.Rodando {
			return m, tea.Tick(passoMCP, func(time.Time) tea.Msg { return tickMCPMsg{} })
		}
		m.vigiaMCP = false
		return m.mostrarToast("MCPs: "+resumoMCP(msg.ProgressoMCP), T.Ok)
	case toastFimMsg:
		if msg.gen == m.genToast {
			m.toast = ""
		}
		return m, nil
	case respondeuMsg:
		m = m.respondeu(msg)
		cor := T.Ok
		switch {
		case msg.error != nil:
			cor = T.Erro
		case !msg.Entregue:
			cor = T.Espera
		}
		return m.mostrarToast(m.status, cor)
	case tea.MouseMsg:
		return m.mouse(msg)
	case tea.KeyMsg:
		if m.ajuda {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			m.ajuda = false
			return m, nil
		}
		if m.aberto != nil {
			return m.teclaDetalhe(msg)
		}
		if m.filtro != nil {
			return m.teclaFiltro(msg)
		}
		return m.tecla(msg.String())
	default:
		if m.rascunho != nil {
			ti, cmd := m.rascunho.Update(msg)
			m.rascunho = &ti
			return m, cmd
		}
		if m.filtro != nil {
			ti, cmd := m.filtro.Update(msg)
			m.filtro = &ti
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) mouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		d := 1
		if msg.Button == tea.MouseButtonWheelUp {
			d = -1
		}
		if m.aberto != nil {
			m.armado = 0
			return m.rolar(3 * d), nil
		}
		return m.mover(d), nil
	}
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return m, nil
	}
	if m.ajuda {
		m.ajuda = false
		return m, nil
	}
	q := m.montar(m.larguraEfetiva(), m.alturaEfetiva())
	armado, armadoArq := m.armado, m.armadoArq
	m.armado, m.armadoArq = 0, ""
	x, y := msg.X, msg.Y
	dentro := func(s faixaX) bool { return s.y == y && x >= s.x0 && x < s.x1 }
	switch {
	case dentro(q.atualizar):
		return m.coletar()
	case dentro(q.voltar):
		return m.fecharDetalhe(), nil
	case dentro(q.irAba):
		return m.irAba(), nil
	case dentro(q.arquivar):
		m.armadoArq = armadoArq
		return m.arquivar()
	case dentro(q.abrir):
		_, u := m.atual()
		return m.abrirURL(u), nil
	case y == len(q.linhas)-1:
		for i, a := range q.abas {
			if x >= a[0] && x < a[1] {
				return m.fecharDetalhe().irGaveta(i), nil
			}
		}
		return m, nil
	}
	r := y - q.topo
	if r < 0 || r >= len(q.dono) || q.dono[r] < 0 {
		return m, nil
	}
	if m.aberto != nil {
		if m.rascunho != nil {
			return m, nil
		}
		m.opcao, m.armado, m.focoTexto = q.dono[r], armado, false
		return m.teclaDetalhe(tea.KeyMsg{Type: tea.KeyEnter})
	}
	m.sel = q.dono[r]
	return m.enterCmd()
}

func (m Model) tecla(k string) (tea.Model, tea.Cmd) {
	armadoArq, armadoRet := m.armadoArq, m.armadoRet
	m.armadoArq, m.armadoRet = "", ""
	switch k {
	case "R":
		m.armadoRet = armadoRet
		return m.retomar()
	case "x":
		m.armadoArq = armadoArq
		return m.arquivar()
	case "a":
		if Gavetas[m.gaveta].ID == "agora" {
			m.antigas = !m.antigas
			return m.mover(0), nil
		}
	case "d":
		if Gavetas[m.gaveta].ID == "prs" {
			m.rascunhos = !m.rascunhos
			return m.mover(0), nil
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.ajuda = true
		return m, nil
	case "/":
		return m.abrirFiltro()
	case "esc":
		m.busca, m.sel = "", 0
		return m, nil
	case "r":
		return m.coletar()
	case "left", "h":
		return m.irGaveta(m.gaveta - 1), nil
	case "right", "l":
		return m.irGaveta(m.gaveta + 1), nil
	case "up", "k":
		return m.mover(-1), nil
	case "down", "j":
		return m.mover(1), nil
	case "enter":
		return m.enterCmd()
	case "D":
		return m.daydreaming()
	case "M":
		return m.reconectarMCPs()
	case "o":
		if a, ok := m.selecionado(); ok {
			return m.abrirURL(a.url), nil
		}
	case "g":
		return m.irAba(), nil
	case "p":
		return m.abrirURL(PainelURL), nil
	}
	if len(k) == 1 && k[0] >= '1' && int(k[0]-'1') < len(Gavetas) {
		return m.irGaveta(int(k[0] - '1')), nil
	}
	return m, nil
}

func (m Model) montarCorpo(largura, altura int) *corpo {
	c := &corpo{largura: largura, filtro: normaliza(m.busca), antigasAbertas: m.antigas, rascunhos: m.rascunhos}
	e := m.estado
	switch Gavetas[m.gaveta].ID {
	case "agora":
		gavetaAgora(c, e, altura)
	case "crew":
		gavetaCrew(c, e)
	case "prs":
		gavetaPRs(c, e)
	case "ci":
		gavetaCI(c, e)
	case "ambiente":
		gavetaAmbiente(c, e)
	case "maquina":
		gavetaMaquina(c, e, m.pedindoDD)
	case "mais":
		gavetaMais(c, e)
	}
	if c.filtro != "" && c.vis == 0 {
		c.branco()
		c.vazio("nada bate com “" + strings.TrimSpace(m.busca) + "”")
	}
	return c
}

func esqueleto(largura, altura int) []string {
	n := 8
	if largura < 70 {
		n = 4
	}
	var out []string
	for i := 0; i < n && i < altura; i++ {
		w := largura - 4 - (i%3)*6
		if w < 1 {
			w = 1
		}
		out = append(out, encaixa(st(T.Base, T.Sel).Render("▌ "+strings.Repeat("░", w)), largura, T.Base))
	}
	return out
}

// corpoLinhas devolve as linhas visíveis da gaveta e, para cada uma, o item dono (ou -1).
func (m Model) corpoLinhas(largura, altura int) ([]string, []int) {
	if m.estado.At.IsZero() {
		ls := esqueleto(largura, altura)
		dono := make([]int, len(ls))
		for i := range dono {
			dono[i] = -1
		}
		return ls, dono
	}
	c := m.montarCorpo(largura, altura)
	ini, fim := -1, -1
	for i, l := range c.linhas {
		if l.item == m.sel {
			if ini < 0 {
				ini = i
			}
			fim = i
		}
	}
	desl := 0
	if fim >= altura {
		desl = fim - altura + 1
	}
	if ini >= 0 && ini < desl {
		desl = ini
	}
	var out []string
	var dono []int
	for i := desl; i < len(c.linhas) && len(out) < altura; i++ {
		l := c.linhas[i]
		bg := T.Base
		sel := l.item >= 0 && l.item == m.sel
		if sel {
			bg = T.Sel
		}
		s := espacos(1, bg) + l.desenha(bg)
		if sel && i == ini && l.item < len(c.alvos) {
			a := c.alvos[l.item]
			if b := botoesItem(bg, a, m.armadoArq); b != "" && largura-lipgloss.Width(b)-3 >= 16 {
				s = encaixa(s, largura-lipgloss.Width(b)-3, bg) + espacos(2, bg) + b
			}
		}
		out = append(out, encaixa(s, largura, bg))
		dono = append(dono, l.item)
	}
	return out, dono
}

func (m Model) ajudaLinhas(largura int) []string {
	k := func(t, d string) string { return Col("  "+t, 16) + d }
	ls := []string{
		"teclado",
		k("↑ ↓  j k", "escolhe item ou opção"),
		k("enter", "abre · responde"),
		k("enter 2×", "confirma Liberar merge"),
		k("", "e Não vou fazer"),
		k("← → 1-7", "troca de gaveta"),
		k("/", "filtra a lista"),
		k("esc", "volta · limpa o filtro"),
		k("o", "abre o link (abrir ↗)"),
		k("g", "vai à aba (ir à aba →)"),
		k("p", "abre o painel do delegate"),
		k("a", "abre · recolhe as antigas"),
		k("x 2×", "arquiva a pergunta antiga"),
		k("R 2×", "retoma fantasminha caído (crew)"),
		k("d", "PRs: mostra · esconde rascunhos"),
		k("D", "pede rodada do DayDreaming"),
		k("M", "reconecta os MCPs de todas as sessões"),
		k("r", "atualiza agora"),
		k("q", "sai"),
		"",
		"no detalhe",
		k("1-9", "escolhe a opção N"),
		k("tab", "foco: texto ↔ opções"),
		k("pgup pgdn", "rola o texto"),
		k("d", "mostra o diagrama"),
		"",
		"mouse",
		k("clique", "no item: abre"),
		k("clique", "na opção: responde"),
		k("clique", "na aba: troca de gaveta"),
		k("clique", "↻ atualiza · ‹ voltar"),
		k("clique", "ir à aba → · abrir ↗"),
		k("clique", "antigas: abre · arquivar ✕"),
		k("roda", "rola a lista ou o texto"),
		"",
		"? fecha · qualquer tecla ou clique também",
	}
	if m.modoImg != "" {
		ls = append(ls, "", "gengar: "+m.modoImg)
	}
	out := make([]string, len(ls))
	for i, l := range ls {
		cor := T.Sub
		switch {
		case l == "teclado" || l == "mouse" || l == "no detalhe":
			out[i] = encaixa(espacos(2, T.Base)+rotuloSecao(l), largura, T.Base)
			continue
		case strings.HasPrefix(l, "?"):
			cor = T.Dim
		}
		out[i] = encaixa(espacos(2, T.Base)+st(T.Base, cor).Render(l), largura, T.Base)
	}
	return out
}

type faixaX struct{ y, x0, x1 int }

// tela é o quadro desenhado mais o mapa de cliques.
type tela struct {
	linhas    []string
	topo      int
	dono      []int
	abas      [][2]int
	atualizar faixaX
	voltar    faixaX
	irAba     faixaX
	abrir     faixaX
	arquivar  faixaX
	img       areaImg
}

func (m Model) linhasTopo(largura int) ([]string, areaImg) {
	status := Corte(m.status, largura)
	if m.aberto != nil || m.ajuda {
		return []string{TopoCompacto(m.estado, largura, m.quadro, m.coletando, status)}, areaImg{}
	}
	return topoArea(m.estado, largura, m.quadro, m.coletando, status, m.img != nil)
}

// achar devolve a faixa da primeira linha que contém s.
func achar(ls []string, s string, largura int) faixaX {
	for y, l := range ls {
		p := ansi.Strip(l)
		if i := strings.Index(p, s); i >= 0 {
			x := lipgloss.Width(p[:i])
			return faixaX{y, x, x + lipgloss.Width(s)}
		}
	}
	return faixaX{y: -1}
}

func (m Model) montar(largura, altura int) tela {
	var q tela
	q.atualizar, q.voltar, q.irAba, q.abrir, q.arquivar = faixaX{y: -1}, faixaX{y: -1}, faixaX{y: -1}, faixaX{y: -1}, faixaX{y: -1}
	if largura <= 0 || altura <= 0 {
		return q
	}
	topo, area := m.linhasTopo(largura)
	if len(topo) > altura-1 {
		topo = topo[:max(0, altura-1)]
	}
	q.topo = len(topo)
	if area.ok && area.y+area.r <= q.topo {
		q.img = area
	}
	q.atualizar = achar(topo, "↻ atualizado", largura)
	if m.aberto != nil || m.ajuda {
		q.voltar = achar(topo, Voltar, largura)
	}
	h := altura - 1 - len(topo)
	linhas := append([]string{}, topo...)
	if h > 0 {
		var corpo []string
		switch {
		case m.ajuda:
			corpo = m.ajudaLinhas(largura)
		case m.aberto != nil:
			corpo, q.dono = m.detalheLinhas(largura, h)
		default:
			hl := h
			if m.temBarra() && h > 1 {
				hl = h - 1
			}
			corpo, q.dono = m.corpoLinhas(largura, hl)
			for len(corpo) < hl {
				corpo = append(corpo, espacos(largura, T.Base))
			}
		}
		if len(corpo) > h {
			corpo = corpo[:h]
		}
		for len(corpo) < h {
			corpo = append(corpo, espacos(largura, T.Base))
		}
		barra := m.temBarra() && h > 1
		if barra {
			corpo[h-1] = m.linhaFiltro(largura)
		}
		if m.toast != "" {
			y := h - 1
			if barra {
				y = h - 2
			}
			corpo[y] = m.linhaToast(largura)
		}
		linhas = append(linhas, corpo...)
	}
	rod, abas := rodapeAbas(Gavetas[m.gaveta].ID, largura)
	q.abas = abas
	linhas = append(linhas, rod)
	for i, l := range linhas {
		if lipgloss.Width(l) > largura {
			linhas[i] = encaixa(l, largura, T.Base)
		}
	}
	q.linhas = linhas
	if !m.ajuda {
		q.irAba = achar(linhas, BotaoAba, largura)
		q.abrir = achar(linhas, BotaoAbrir, largura)
		if q.arquivar = achar(linhas, BotaoConfirmaArq, largura); q.arquivar.y < 0 {
			q.arquivar = achar(linhas, BotaoArquivar, largura)
		}
	}
	return q
}

const (
	BotaoAba         = "ir à aba →"
	BotaoAbrir       = "abrir ↗"
	BotaoArquivar    = "arquivar ✕"
	BotaoConfirmaArq = "confirma arquivar?"
)

// botoes desenha os botões de ação do item; sem aba ou URL, o botão não aparece.
func botoes(bg lipgloss.Color, aba, url bool) string {
	var ps []string
	if aba {
		ps = append(ps, st(bg, T.Acao).Render(BotaoAba))
	}
	if url {
		ps = append(ps, st(bg, T.Info).Render(BotaoAbrir))
	}
	return strings.Join(ps, espacos(2, bg))
}

func (m Model) linhaFiltro(largura int) string {
	c := m.montarCorpo(largura, m.alturaLista())
	conta := fmt.Sprintf("%d de %d", c.vis, c.total)
	if m.filtro == nil {
		conta += " · esc limpa"
	}
	dir := st(T.Surface, T.Dim).Render(conta + " ")
	esq := st(T.Surface, T.Acao).Bold(true).Render(" / ")
	w := max(1, largura-lipgloss.Width(esq)-lipgloss.Width(dir)-1)
	var meio string
	if m.filtro != nil {
		ti := *m.filtro
		ti.Width = max(1, w-1)
		meio = ti.View()
	} else {
		meio = st(T.Surface, T.Texto).Render(Corte(m.busca, w))
	}
	meio = encaixa(meio, w, T.Surface)
	return encaixa(esq+meio+espacos(1, T.Surface)+dir, largura, T.Surface)
}

var semAcento = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ü", "u", "ç", "c")

func normaliza(s string) string { return semAcento.Replace(strings.ToLower(strings.TrimSpace(s))) }

func (m Model) linhaToast(largura int) string {
	g := "●"
	if m.toastCor == T.Erro {
		g = "✗"
	}
	s := st(T.Surface, m.toastCor).Bold(true).Render(" "+g+" ") + st(T.Surface, T.Texto).Render(Corte(m.toast, largura-6)+" ")
	return encaixa(espacos(1, T.Base)+s, largura, T.Base)
}

// Render desenha o quadro com exatamente `altura` linhas, a última sendo o rodapé.
func (m Model) Render(largura, altura int) string {
	return strings.Join(m.montar(largura, altura).linhas, "\n")
}

func (m Model) View() string { return m.Render(m.largura, m.altura) }

func botoesItem(bg lipgloss.Color, a alvo, armadoArq string) string {
	b := botoes(bg, a.aba != "", a.url != "")
	if a.arquivar == "" {
		return b
	}
	arq := st(bg, T.Erro).Render(BotaoArquivar)
	if armadoArq == a.arquivar {
		arq = st(bg, T.Erro).Bold(true).Render(BotaoConfirmaArq)
	}
	if b == "" {
		return arq
	}
	return b + espacos(2, bg) + arq
}

// arquivar pede confirmação: o primeiro x (ou clique) arma, o segundo grava.
func (m Model) retomar() (tea.Model, tea.Cmd) {
	a, ok := m.selecionado()
	if !ok || a.retomar == "" {
		m.armadoRet = ""
		return m, nil
	}
	var c dados.Caido
	for _, x := range m.estado.Caidos {
		if x.Nome == a.retomar {
			c = x
		}
	}
	if c.SessionID == "" || c.Pane == "" {
		m.armadoRet = ""
		return m.mostrarToast("sem session_id; retome à mão", T.Espera)
	}
	if m.armadoRet != c.Nome {
		m.armadoRet = c.Nome
		m.status = "R de novo confirma retomar " + c.Nome
		return m, nil
	}
	m.armadoRet = ""
	if m.acoes.Retomar == nil {
		return m.mostrarToast("não retomei: sem como retomar nesta máquina", T.Erro)
	}
	if err := m.acoes.Retomar(c); err != nil {
		return m.mostrarToast("não retomei: "+err.Error(), T.Erro)
	}
	var resto []dados.Caido
	for _, x := range m.estado.Caidos {
		if x.Nome != c.Nome {
			resto = append(resto, x)
		}
	}
	m.estado.Caidos = resto
	m.status = "retomando " + c.Nome
	m = m.mover(0)
	return m.mostrarToast("retomando "+c.Nome+" no pane "+c.Pane, T.Ok)
}

func (m Model) arquivar() (tea.Model, tea.Cmd) {
	a, ok := m.selecionado()
	if !ok || a.arquivar == "" {
		m.armadoArq = ""
		return m, nil
	}
	if m.armadoArq != a.arquivar {
		m.armadoArq = a.arquivar
		m.status = "x de novo (ou clique) confirma arquivar"
		return m, nil
	}
	m.armadoArq = ""
	if m.acoes.Arquivar != nil {
		if err := m.acoes.Arquivar(a.arquivar); err != nil {
			return m.mostrarToast("não arquivei: "+err.Error(), T.Erro)
		}
	}
	var resto []dados.Linha
	for _, p := range m.estado.Esperando {
		if p.Chave != a.arquivar {
			resto = append(resto, p)
		}
	}
	m.estado.Esperando = resto
	m.status = "arquivada"
	m = m.mover(0)
	return m.mostrarToast("arquivada, sem aviso a ninguém", T.Ok)
}
