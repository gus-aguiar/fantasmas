package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

func atalho(tecla, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(tecla), key.WithHelp(tecla, desc))
}

// mapaAtalhos é o rodapé curto do contexto (até 5); o completo continua no ?.
type mapaAtalhos []key.Binding

func (m mapaAtalhos) ShortHelp() []key.Binding  { return m }
func (m mapaAtalhos) FullHelp() [][]key.Binding { return [][]key.Binding{m} }

func (m Model) atalhosContexto() mapaAtalhos {
	ajuda := atalho("?", "ajuda")
	switch {
	case m.botoes:
		return mapaAtalhos{atalho("enter", "executa"), atalho("g", "ir à aba"), atalho("esc", "fecha")}
	case m.ajuda:
		return mapaAtalhos{atalho("?", "fecha")}
	case m.aberto != nil:
		return mapaAtalhos{atalho("esc", "volta"), atalho("enter", "responde"), atalho("o", "abrir"), atalho("g", "ir à aba"), ajuda}
	case m.pipe != 0 || m.form != nil:
		return mapaAtalhos{atalho("esc", "volta"), atalho("enter", "abre"), ajuda}
	}
	base := mapaAtalhos{atalho("ctrl+p", "comandos"), atalho("enter", "abrir"), atalho("g", "ir à aba")}
	switch Gavetas[m.gaveta].ID {
	case "agora":
		base = append(base, atalho("a", "antigos"))
	case "missoes":
		base = append(base, atalho("n", "sidequest"))
	case "prs":
		base = append(base, atalho("d", "rascunhos"))
	default:
		base = append(base, atalho("/", "filtra"))
	}
	return append(base, ajuda)
}

// rodapeAtalhos desenha o rodapé com bubbles/help: tecla e descrição em cinza.
func (m Model) rodapeAtalhos(largura int) string {
	h := help.New()
	h.ShortSeparator = " · "
	h.Styles.ShortKey = st(T.Mantle, T.Sub)
	h.Styles.ShortDesc = st(T.Mantle, T.Dim)
	h.Styles.ShortSeparator = st(T.Mantle, T.Dim)
	h.Styles.Ellipsis = st(T.Mantle, T.Dim)
	h.Width = max(1, largura-2)
	return linhaFundo(espacos(1, T.Mantle)+h.View(m.atalhosContexto()), largura, false)
}
