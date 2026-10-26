package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Corte limita s a n células de largura, com … no fim quando cortou.
func Corte(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	return ansi.Truncate(s, n, "…")
}

// Col corta e completa com espaços até exatamente n células.
func Col(s string, n int) string {
	if n <= 0 {
		return ""
	}
	s = Corte(s, n)
	if w := lipgloss.Width(s); w < n {
		s += strings.Repeat(" ", n-w)
	}
	return s
}
