package ui

import (
	"fmt"
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
	// MaxSemBusca: até tantos comandos a paleta não mostra o campo de busca.
	MaxSemBusca = 5
)

// cores da estética dos fantasminhas do painel do delegate
var (
	corFtCorpo   = lipgloss.Color("#a66bff")
	corFtOlho    = lipgloss.Color("#ff3355")
	corFtRodando = lipgloss.Color("#67d4f0")
)

func larguraPaleta(largura int) int { return max(24, min(LarguraPaleta, largura-4)) }

// ComBotoes abre a paleta de comandos (modo quadro).
func (m Model) ComBotoes() Model { return m.abrirBotoes() }

func (m Model) abrirBotoes() Model {
	m.selGaveta, m.sel, m.botoes, m.armadoCmd, m.buscaCmd = m.sel, 0, true, "", nil
	if len(dados.ComandosHerdr) > MaxSemBusca {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = "buscar comando"
		ti.TextStyle = st(T.Surface, T.Texto)
		ti.PlaceholderStyle = st(T.Surface, T.Dim)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acento())
		ti.Cursor.TextStyle = st(T.Surface, T.Texto)
		ti.Focus()
		m.buscaCmd = &ti
	}
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

// teclaBotoes: com a paleta aberta, as teclas ficam com ela. 1-N dispara o comando N (2× confirma);
// com a busca vazia (ou sem busca), b g j k q seguem como atalho.
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
		if len(k) == 1 && k[0] >= '1' && k[0] <= '9' {
			if i := int(k[0] - '1'); i < len(m.alvosPaleta()) {
				if i == m.sel {
					m.armadoCmd = armadoCmd
				}
				m.sel = i
				return m.enterCmd()
			}
			return m, nil
		}
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

// estadoFt é o estado do fantasminha de um comando: parado, rodando ou terminou.
type estadoFt int

const (
	ftParado estadoFt = iota
	ftRodando
	ftTerminou
)

// itemPaleta é uma linha da paleta: nome, descrição, estado à direita e o alvo do disparo.
type itemPaleta struct {
	nome, desc, busca, estado string
	ft                        estadoFt
	cor                       lipgloss.Color
	a                         alvo
}

var nomeComando = map[string]string{
	"dia-abrir":    "Abrir o dia",
	"dia-fechar":   "Fechar o dia",
	"saude-multas": "Saúde de multas",
}

// itensPaleta monta os comandos do Herdr com o estado de agora; a lógica de disparo é a de sempre.
func itensPaleta(_ dados.Estado, ec estadoComandos) []itemPaleta {
	var out []itemPaleta
	for _, k := range dados.ComandosHerdr {
		nome, ok := nomeComando[k.ID]
		if !ok {
			nome = k.Rotulo
		}
		it := itemPaleta{nome: nome, desc: k.Rotulo + " · " + k.Descricao, busca: k.Rotulo + " " + k.Descricao, cor: T.Dim, a: alvo{comando: k.ID}}
		if d, ok := ec.disparos[k.ID]; ok {
			it.estado, it.ft, it.a.aba = rodandoHa(d, ec.agora), ftRodando, d.Aba
		}
		switch {
		case ec.disparando == k.ID:
			it.estado, it.ft = "abrindo…", ftRodando
		case ec.armado == k.ID:
			it.estado, it.cor = "enter de novo confirma", T.EstadoAviso()
		}
		out = append(out, it)
	}
	return out
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
		if f == "" || strings.Contains(normaliza(it.nome+" "+it.busca), f) {
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

// miniFantasma é o fantasminha do cabeçalho, desenhado em células (3 linhas × 5 colunas).
func miniFantasma(bg lipgloss.Color) []string {
	c := st(bg, corFtCorpo)
	olho := lipgloss.NewStyle().Background(corFtCorpo).Foreground(corFtOlho)
	return []string{
		c.Render("▗▄▄▄▖"),
		c.Render("█") + olho.Render("▘") + c.Render("█") + olho.Render("▘") + c.Render("█"),
		c.Render("▀▝▀▘▀"),
	}
}

// caixaPaleta desenha a caixa: fantasminha e título, comandos com glifo de estado e rodapé.
func (m Model) caixaPaleta(w int) []string {
	bg := T.Surface
	const pad = 3
	iw := w - 2 - 2*pad
	borda := st(bg, corFtCorpo)
	lado := func(s string, fundo lipgloss.Color) string {
		return borda.Render("│") + espacos(pad, bg) + encaixa(s, iw, fundo) + espacos(pad, bg) + borda.Render("│")
	}
	vazio := lado("", bg)
	out := []string{borda.Render("╭" + strings.Repeat("─", w-2) + "╮"), vazio}
	ft := miniFantasma(bg)
	itens := m.itensFiltrados()
	sub := "1–" + fmt.Sprint(len(itens)) + " dispara · enter 2× confirma"
	cab := []string{"", st(bg, T.Texto).Bold(true).Render("Comandos"), st(bg, T.Dim).Render(sub)}
	for i := range ft {
		out = append(out, lado(ft[i]+espacos(2, bg)+cab[i], bg))
	}
	out = append(out, vazio)
	if m.buscaCmd != nil {
		ti := *m.buscaCmd
		ti.Width = max(1, iw-3)
		out = append(out, lado(st(bg, T.Acento()).Render("› ")+ti.View(), bg), vazio)
	}
	for i, it := range itens {
		fb := bg
		sel := i == m.sel
		if sel {
			fb = T.Sel
		}
		cg := T.Dim
		switch it.ft {
		case ftRodando:
			cg = corFtRodando
		case ftTerminou:
			cg = T.Ok
		}
		esq := st(fb, cg).Render(" ᗣ ") + st(fb, T.Dim).Render(fmt.Sprint(i+1)+" ")
		ns := st(fb, T.Texto)
		if sel {
			ns = ns.Bold(true)
		}
		cor := it.cor
		if it.ft == ftRodando && cor == T.Dim {
			cor = T.Secundario()
		}
		dir := ""
		if it.estado != "" {
			dir = st(fb, cor).Render(Corte(it.estado, max(4, iw-lipgloss.Width(esq)-lipgloss.Width(it.nome)-3)) + " ")
		}
		nome := ns.Render(Corte(it.nome, max(1, iw-lipgloss.Width(esq)-1)))
		meio := iw - lipgloss.Width(esq) - lipgloss.Width(nome) - lipgloss.Width(dir)
		s := esq + nome
		if meio >= 1 {
			s += espacos(meio, fb) + dir
		}
		out = append(out, lado(s, fb))
		if sel && it.desc != "" {
			out = append(out, lado(espacos(5, bg)+st(bg, T.Dim).Render(Corte(it.desc, max(1, iw-5))), bg))
		}
	}
	if len(itens) == 0 {
		out = append(out, lado(st(bg, T.Secundario()).Render("nada bate com “"+strings.TrimSpace(m.textoBusca())+"”"), bg))
	}
	rodape := "esc fecha"
	if a, ok := m.selecionado(); ok && a.aba != "" {
		rodape = "g vai à aba · esc fecha"
	}
	out = append(out, vazio, lado(st(bg, T.Dim).Render(rodape), bg), vazio)
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
