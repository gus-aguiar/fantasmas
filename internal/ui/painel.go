package ui

import (
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// LarguraPaleta é a largura da caixa de comandos; em janela estreita ela encolhe.
const (
	LarguraPaleta = 58
	DicaPaleta    = "ctrl+p comandos"
)

func larguraPaleta(largura int) int { return max(24, min(LarguraPaleta, largura-4)) }

// ComBotoes abre a paleta de comandos (modo quadro).
func (m Model) ComBotoes() Model { return m.abrirBotoes() }

func (m Model) abrirBotoes() Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "buscar comando"
	ti.TextStyle = st(T.Surface, T.Texto)
	ti.PlaceholderStyle = st(T.Surface, T.Dim)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acento())
	ti.Cursor.TextStyle = st(T.Surface, T.Texto)
	ti.Focus()
	m.selGaveta, m.sel, m.botoes, m.armadoCmd, m.buscaCmd = m.sel, 0, true, "", &ti
	return m
}

func (m Model) fecharBotoes() Model {
	m.botoes, m.sel, m.armadoCmd, m.buscaCmd = false, m.selGaveta, "", nil
	return m
}

func (m Model) textoBusca() string {
	if m.buscaCmd == nil {
		return ""
	}
	return m.buscaCmd.Value()
}

// teclaBotoes: com a paleta aberta, as teclas ficam com ela. Com a busca vazia, b g j k q
// seguem como atalho; qualquer outra letra (ou estas, depois da primeira) vai para a busca.
func (m Model) teclaBotoes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	armadoCmd := m.armadoCmd
	m.armadoCmd = ""
	vazia := m.textoBusca() == ""
	switch k {
	case "esc", "ctrl+p":
		return m.fecharBotoes(), nil
	case "ctrl+c":
		return m, tea.Quit
	case "up", "ctrl+k":
		return m.mover(-1), nil
	case "down", "ctrl+j", "tab":
		return m.mover(1), nil
	case "enter":
		m.armadoCmd = armadoCmd
		return m.enterCmd()
	}
	if vazia {
		switch k {
		case "b":
			return m.fecharBotoes(), nil
		case "q":
			return m, tea.Quit
		case "k":
			return m.mover(-1), nil
		case "j":
			return m.mover(1), nil
		case "g":
			return m.irAba(), nil
		}
	}
	if m.buscaCmd == nil {
		return m, nil
	}
	ti, cmd := m.buscaCmd.Update(msg)
	m.buscaCmd = &ti
	m.sel = 0
	return m.mover(0), cmd
}

// itemPaleta é uma linha da paleta: grupo, nome, estado à direita e o alvo do disparo.
type itemPaleta struct {
	grupo, nome, busca, estado string
	cor                        lipgloss.Color
	a                          alvo
}

var nomeComando = map[string][2]string{
	"dia-abrir":    {"DIA", "Abrir o dia"},
	"dia-fechar":   {"DIA", "Fechar o dia"},
	"saude-multas": {"MULTAS", "Saúde de multas"},
}

// itensPaleta monta os comandos com o estado de agora; a lógica de disparo é a de sempre.
func itensPaleta(e dados.Estado, ec estadoComandos) []itemPaleta {
	var out []itemPaleta
	confirma := func(id string) bool { return ec.armado == id }
	for _, k := range dados.ComandosHerdr {
		gn, ok := nomeComando[k.ID]
		if !ok {
			gn = [2]string{"SISTEMA", k.Rotulo}
		}
		it := itemPaleta{grupo: gn[0], nome: gn[1], busca: k.Rotulo + " " + k.Descricao, estado: k.Rotulo, cor: T.Secundario(), a: alvo{comando: k.ID}}
		if d, ok := ec.disparos[k.ID]; ok {
			it.estado, it.cor, it.a.aba = rodandoHa(d, ec.agora), T.Andamento, d.Aba
		}
		switch {
		case ec.disparando == k.ID:
			it.estado, it.cor = "abrindo…", T.Andamento
		case confirma(k.ID):
			it.estado, it.cor = "enter de novo confirma", T.EstadoAviso()
		}
		out = append(out, it)
	}
	semTecla := func(s string) string {
		if i := strings.LastIndex(s, " · "); i >= 0 {
			return s[:i]
		}
		return s
	}
	junta := func(a, b string) string {
		if a == "" {
			return b
		}
		return a + " · " + b
	}
	_, cor, est, det := botaoDaydreaming(e.Daydreamer, ec.pedindoDD)
	dd := itemPaleta{grupo: "SISTEMA", nome: "DayDreaming", busca: "daydreaming sonho", estado: junta(est, semTecla(det)), cor: cor, a: alvo{comando: cmdDaydreaming}}
	if cor == T.Dim || cor == T.Ok {
		dd.cor = T.Secundario()
	}
	_, cor, est, det = botaoMCP(e.MCP)
	mcp := itemPaleta{grupo: "SISTEMA", nome: "Reconectar MCPs", busca: "mcp reconectar", estado: junta(est, semTecla(det)), cor: cor, a: alvo{comando: cmdMCP}}
	if cor == T.Dim || cor == T.Ok {
		mcp.cor = T.Secundario()
	}
	atu := itemPaleta{grupo: "SISTEMA", nome: "Atualizar app", busca: "atualizar app fantasmas", estado: "git pull, build e reabre", cor: T.Secundario(), a: alvo{comando: cmdAtualizar}}
	if ec.atualizando {
		atu.estado, atu.cor = "atualizando…", T.Andamento
	}
	for _, it := range []*itemPaleta{&dd, &mcp, &atu} {
		if confirma(it.a.comando) || (it.a.comando == cmdAtualizar && ec.armadoAtu) {
			it.estado, it.cor = "enter de novo confirma", T.EstadoAviso()
		}
	}
	return append(out, dd, mcp, atu)
}

func (m Model) estadoCmds() estadoComandos {
	return estadoComandos{disparos: m.disparos, armado: m.armadoCmd, disparando: m.disparando,
		agora: m.agora(), pedindoDD: m.pedindoDD, atualizando: m.atualizando, armadoAtu: m.armadoAtu}
}

// itensFiltrados aplica a busca da paleta.
func (m Model) itensFiltrados() []itemPaleta {
	f := normaliza(m.textoBusca())
	var out []itemPaleta
	for _, it := range itensPaleta(m.estado, m.estadoCmds()) {
		if f == "" || strings.Contains(normaliza(it.grupo+" "+it.nome+" "+it.busca), f) {
			out = append(out, it)
		}
	}
	return out
}

func (m Model) alvosPaleta() []alvo {
	var as []alvo
	for _, it := range m.itensFiltrados() {
		as = append(as, it.a)
	}
	return as
}

// caixaPaleta desenha a caixa centralizável: título na borda, busca, grupos e rodapé.
func (m Model) caixaPaleta(w int) []string {
	bg := T.Surface
	iw := w - 2 - 4 // bordas e padding 2 de cada lado
	borda := st(bg, T.Acento())
	lado := func(s string, fundo lipgloss.Color) string {
		return borda.Render("│") + espacos(2, fundo) + encaixa(s, iw, fundo) + espacos(2, fundo) + borda.Render("│")
	}
	vazio := lado("", bg)
	titulo := " Comandos "
	topo := borda.Render("╭─") + st(bg, T.Acento()).Bold(true).Render(titulo) + borda.Render(strings.Repeat("─", max(0, w-3-lipgloss.Width(titulo)))+"╮")
	out := []string{topo, vazio}
	busca := st(bg, T.Acento()).Render("› ")
	if m.buscaCmd != nil {
		ti := *m.buscaCmd
		ti.Width = max(1, iw-3)
		busca += ti.View()
	}
	out = append(out, lado(busca, bg), vazio)
	itens := m.itensFiltrados()
	grupo := ""
	for i, it := range itens {
		if it.grupo != grupo {
			if grupo != "" {
				out = append(out, vazio)
			}
			grupo = it.grupo
			out = append(out, lado(st(bg, T.Secundario()).Faint(true).Bold(true).Render(it.grupo), bg))
		}
		fb := bg
		marca := espacos(2, fb)
		if i == m.sel {
			fb = T.Sel
			marca = st(fb, T.Acento()).Render("▌ ")
		}
		dir := ""
		if it.estado != "" {
			dir = st(fb, it.cor).Render(Corte(it.estado, max(4, iw-2-lipgloss.Width(it.nome)-3)))
		}
		nome := st(fb, T.Texto).Render(Corte(it.nome, iw-2))
		meio := iw - 2 - lipgloss.Width(nome) - lipgloss.Width(dir)
		s := marca + nome + espacos(meio, fb) + dir
		if meio < 1 {
			s = marca + nome
		}
		out = append(out, lado(s, fb))
	}
	if len(itens) == 0 {
		out = append(out, lado(st(bg, T.Secundario()).Render("nada bate com “"+strings.TrimSpace(m.textoBusca())+"”"), bg))
	}
	rodape := "enter executa · esc fecha"
	if a, ok := m.selecionado(); ok && a.aba != "" {
		rodape = "enter executa · g vai à aba · esc fecha"
	}
	out = append(out, vazio, lado(st(bg, T.Secundario()).Render(rodape), bg), vazio)
	out = append(out, borda.Render("╰"+strings.Repeat("─", w-2)+"╯"))
	return out
}

// cortaCols devolve as colunas [a, b) de um texto sem cor, completando com espaço.
func cortaCols(s string, a, b int) string {
	var sb strings.Builder
	x := 0
	for _, r := range s {
		cw := lipgloss.Width(string(r))
		if x >= a && x+cw <= b {
			sb.WriteRune(r)
		} else if x < b && x+cw > a {
			sb.WriteString(strings.Repeat(" ", min(x+cw, b)-max(x, a)))
		}
		x += cw
	}
	if x < b {
		sb.WriteString(strings.Repeat(" ", b-max(x, a)))
	}
	return sb.String()
}

// sobreporPaleta esmaece a tela e põe a caixa no centro (lipgloss.Place não sobrepõe na v1).
func (m Model) sobreporPaleta(tela []string, largura, altura int) []string {
	w := larguraPaleta(largura)
	cx := m.caixaPaleta(w)
	if len(cx) > altura {
		cx = cx[:altura]
	}
	x0 := max(0, (largura-w)/2)
	y0 := max(0, (altura-len(cx))/2)
	apaga := lipgloss.NewStyle().Background(T.Base).Foreground(T.Secundario()).Faint(true)
	out := make([]string, len(tela))
	for i, l := range tela {
		plano := ansi.Strip(l)
		if i < y0 || i >= y0+len(cx) {
			out[i] = apaga.Render(cortaCols(plano, 0, largura))
			continue
		}
		out[i] = apaga.Render(cortaCols(plano, 0, x0)) + cx[i-y0] + apaga.Render(cortaCols(plano, x0+w, largura))
	}
	return out
}
