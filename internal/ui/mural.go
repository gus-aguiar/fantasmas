package ui

import (
	"github.com/charmbracelet/lipgloss"

	"fantasmas/internal/dados"
)

// renderAviso desenha "◆ <autor> · HH:MM  <texto>" em no máximo largura células.
func renderAviso(a dados.Aviso, largura int, bg lipgloss.Color) string {
	if largura <= 0 {
		return ""
	}
	wAutor := 22
	if largura < 56 {
		wAutor = 12
	}
	if largura < 30 {
		wAutor = max(0, largura/3)
	}
	cab := Corte(a.Autor, wAutor) + " · " + a.At.Format("15:04")
	if wAutor == 0 {
		cab = a.At.Format("15:04")
	}
	s := st(bg, T.Info).Render("◆ ") + st(bg, T.Texto).Bold(true).Render(cab) + espacos(2, bg)
	if resto := largura - lipgloss.Width(s); resto > 0 {
		s += st(bg, T.Sub).Render(Corte(a.Texto, resto))
	} else {
		s = Corte(s, largura)
	}
	return s
}
