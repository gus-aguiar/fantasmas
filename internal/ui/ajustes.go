package ui

import (
	"fmt"
	"os"
	"strings"

	"fantasmas/internal/config"
	"fantasmas/internal/dados"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Gaveta configurações: o tema e as variáveis FANTASMAS_*, salvos fora do repo pela config.Loja.
// Valor de texto livre nunca aparece: nem na lista, nem no campo de edição, nem no toast.

// OpsAjustes é o que a gaveta configurações usa da config.Loja.
type OpsAjustes interface {
	Origem(chave string) config.Origem
	Valor(chave string) string
	Definir(chave, valor string) error
	Caminho() string
}

// Mascara ocupa o lugar de um valor de texto livre.
const Mascara = "••••••"

// editaAjuste é a barra de edição de um ajuste de texto livre.
type editaAjuste struct {
	chave string
	campo textinput.Model
}

// nomeAjuste é o nome que aparece na aba: a própria variável de ambiente.
func nomeAjuste(chave string) string { return chave }

// mostravel devolve o valor que pode ir à tela: escolhas fixas aparecem, texto livre vira a máscara.
func mostravel(o config.Opcao, valor string) string {
	if o.Livre() {
		return Mascara
	}
	return valor
}

func gavetaSettings(c *corpo, e dados.Estado, ops OpsAjustes, pedindoDD, atualizando, armadoAtu bool) {
	secaoAcoes(c, e, pedindoDD, atualizando, armadoAtu)
	c.branco()
	if ops == nil {
		c.branco()
		c.vazioIcone("◆", T.Espera, "configurações indisponíveis nesta execução")
		return
	}
	grupo := ""
	for _, o := range config.Opcoes {
		if o.Grupo != grupo {
			if grupo != "" {
				c.branco()
			}
			extra := ""
			if grupo == "" {
				extra = "enter troca ou edita"
			}
			grupo = o.Grupo
			c.secao(grupo, extra)
		}
		glifo, cor, estado, det := "○", T.Dim, o.Padrao, o.Descricao
		if estado == "" {
			estado = "padrão"
		}
		if o.Chave == config.ChaveFigura && len(o.Valores) > 1 {
			det += fmt.Sprintf(" · %d figuras", len(o.Valores))
		}
		switch ops.Origem(o.Chave) {
		case config.Arquivo:
			glifo, cor, estado = "●", T.Ok, mostravel(o, ops.Valor(o.Chave))
		case config.Ambiente:
			glifo, cor, estado = "◆", T.Info, mostravel(o, ops.Valor(o.Chave))
			det = "vem do shell · " + det
		}
		if o.Reabrir {
			det += " · vale ao reabrir"
		}
		c.linhaAjuste(alvo{ajuste: o.Chave}, glifo, cor, nomeAjuste(o.Chave), estado, det)
	}
	c.branco()
	caminho := ops.Caminho()
	if h, err := os.UserHomeDir(); err == nil && strings.HasPrefix(caminho, h+string(os.PathSeparator)) {
		caminho = "~" + caminho[len(h):]
	}
	c.texto(func(bg lipgloss.Color) string {
		return st(bg, T.Dim).Render("  ● salvo aqui · ◆ vem do shell e manda · " + Corte(caminho, max(8, c.largura-46)))
	})
}

// linhaAjuste: variável, valor e descrição, com colunas largas o bastante para os caminhos padrão.
func (c *corpo) linhaAjuste(a alvo, glifo string, cor lipgloss.Color, nome, valor, det string) {
	wNome, wValor := 24, 26
	if c.largura < 90 {
		wNome, wValor = 22, 18
	}
	wDet := c.largura - 4 - wNome - wValor
	if !c.passa(nome, valor, det) {
		return
	}
	c.item(a, func(bg lipgloss.Color) string {
		s := st(bg, cor).Render(glifo+" ") + st(bg, T.Texto).Render(Col(nome, wNome)) + st(bg, cor).Render(Col(valor, wValor))
		if wDet > 6 {
			s += st(bg, T.Dim).Render(Corte(det, wDet))
		}
		return s
	})
}

// acionarAjuste: escolha fixa avança para o próximo valor; texto livre abre a barra de edição.
func (m Model) acionarAjuste(chave string) (tea.Model, tea.Cmd) {
	o, ok := config.Achar(chave)
	ops := m.acoes.Ajustes
	if !ok || ops == nil {
		return m, nil
	}
	if ops.Origem(chave) == config.Ambiente {
		return m.mostrarToast(nomeAjuste(chave)+" vem do shell; tire de lá para mudar aqui", T.Espera)
	}
	if o.Livre() {
		ti := textinput.New()
		ti.Prompt = ""
		ti.Placeholder = "novo valor · vazio volta ao padrão (" + o.Padrao + ")"
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
		ti.TextStyle = st(T.Surface, T.Texto)
		ti.PlaceholderStyle = st(T.Surface, T.Dim)
		ti.Cursor.Style = lipgloss.NewStyle().Foreground(T.Acao)
		ti.Cursor.TextStyle = st(T.Surface, T.Texto)
		ti.Focus()
		m.ajuste = &editaAjuste{chave: chave, campo: ti}
		return m, textinput.Blink
	}
	if o.Chave == config.ChaveFigura && len(o.Valores) <= 1 {
		return m.mostrarToast("só há o gengar; ponha GIF ou PNG em "+m.acoes.DirFiguras+" e reabra", T.Espera)
	}
	return m.salvarAjuste(o, proximoValor(o, ops))
}

// proximoValor percorre padrão → valores → padrão; o Padrao da opção ("roxo", "gengar") é o próprio "".
func proximoValor(o config.Opcao, ops OpsAjustes) string {
	seq := []string{""}
	for _, v := range o.Valores {
		if v != o.Padrao {
			seq = append(seq, v)
		}
	}
	atual := ""
	if ops.Origem(o.Chave) == config.Arquivo {
		atual = ops.Valor(o.Chave)
	}
	for i, v := range seq {
		if v == atual {
			return seq[(i+1)%len(seq)]
		}
	}
	return seq[0]
}

func (m Model) salvarAjuste(o config.Opcao, valor string) (tea.Model, tea.Cmd) {
	if err := m.acoes.Ajustes.Definir(o.Chave, valor); err != nil {
		return m.mostrarToast(err.Error(), T.Erro)
	}
	switch o.Chave {
	case config.ChaveTema:
		if err := UsarTema(valor); err != nil {
			return m.mostrarToast(err.Error(), T.Erro)
		}
	case config.ChaveFigura:
		if os.Getenv("FANTASMAS_SPRITE") != "" {
			return m.mostrarToast("figura salva; o sprite definido manda sobre ela", T.Espera)
		}
		if err := UsarFigura(valor, m.acoes.DirFiguras); err != nil {
			return m.mostrarToast(err.Error(), T.Erro)
		}
		m = m.recarregarGrafico()
	}
	msg := nomeAjuste(o.Chave) + ": "
	switch {
	case valor == "" && o.Padrao != "":
		msg += o.Padrao
	case valor == "":
		msg += "padrão"
	default:
		msg += mostravel(o, valor)
	}
	if o.Reabrir {
		msg += " · vale ao reabrir"
	}
	return m.mostrarToast(msg, T.Ok)
}

// recarregarGrafico remonta a imagem kitty com a figura atual; o sprite em células já relê sozinho.
func (m Model) recarregarGrafico() Model {
	if m.img == nil {
		return m
	}
	if g, err := m.img.Recarregado(); err == nil {
		m.img, m.imgPos = g, areaImg{}
		return m.sincImagem(true)
	}
	return m
}

func (m Model) teclaAjuste(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.ajuste = nil
		return m, nil
	case "enter":
		e := m.ajuste
		m.ajuste = nil
		o, _ := config.Achar(e.chave)
		return m.salvarAjuste(o, strings.TrimSpace(e.campo.Value()))
	}
	e := *m.ajuste
	ti, cmd := e.campo.Update(msg)
	e.campo = ti
	m.ajuste = &e
	return m, cmd
}

func (m Model) linhaAjuste(largura int) string {
	e := m.ajuste
	esq := st(T.Surface, T.Acao).Bold(true).Render(" " + nomeAjuste(e.chave) + " ")
	dir := st(T.Surface, T.Dim).Render("enter salva · esc cancela ")
	w := max(1, largura-lipgloss.Width(esq)-lipgloss.Width(dir)-1)
	ti := e.campo
	ti.Width = max(1, w-1)
	meio := encaixa(ti.View(), w, T.Surface)
	return encaixa(esq+meio+espacos(1, T.Surface)+dir, largura, T.Surface)
}
