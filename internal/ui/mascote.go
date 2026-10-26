package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const MascoteColunas, MascoteLinhas = 8, 4

const (
	corpoCor = lipgloss.Color("#a66bff")
	olhoCor  = lipgloss.Color("#ff3355")
	bocaCor  = lipgloss.Color("#4b2a8a")
)

var (
	corpoAberto = [3]string{".CCCCCC.", "CCECCECC", "CCCLRCCC"}
	corpoPisca  = [3]string{".CCCCCC.", "CC-CC-CC", "CCCLRCCC"}
	pes         = [2]string{"CC.CC.CC", "C.C..C.C"}
	olhos       = [16]int{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0}
)

func celula(c byte, fundo lipgloss.Color) string {
	corpo := lipgloss.NewStyle().Background(corpoCor)
	switch c {
	case 'C':
		return corpo.Render(" ")
	case 'E':
		return corpo.Foreground(olhoCor).Render("●")
	case '-':
		return corpo.Foreground(bocaCor).Render("─")
	case 'L':
		return corpo.Foreground(bocaCor).Render("╰")
	case 'R':
		return corpo.Foreground(bocaCor).Render("╯")
	}
	return lipgloss.NewStyle().Background(fundo).Render(" ")
}

// QuadroMascote devolve as 4 linhas (8 células) do quadro i; o ciclo tem 16 quadros.
func QuadroMascote(i int, fundo lipgloss.Color) []string {
	i = ((i % 16) + 16) % 16
	corpo := corpoAberto
	if olhos[i] == 1 {
		corpo = corpoPisca
	}
	linhas := []string{corpo[0], corpo[1], corpo[2], pes[i%2]}
	out := make([]string, len(linhas))
	for r, l := range linhas {
		var b strings.Builder
		for c := 0; c < len(l); c++ {
			b.WriteString(celula(l[c], fundo))
		}
		out[r] = b.String()
	}
	return out
}
