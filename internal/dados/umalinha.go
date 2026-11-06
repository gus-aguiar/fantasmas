package dados

import "strings"

// UmaLinha junta quebras de linha e espaços repetidos num espaço só. Campo que a TUI desenha numa linha
// não pode trazer "\n": a altura da tela sai errada e o Bubble Tea desenha a gaveta em branco (06/11).
func UmaLinha(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
