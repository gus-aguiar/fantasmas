package ui

import "github.com/charmbracelet/lipgloss"

type Tokens struct {
	Base, Mantle, Surface, Sel              lipgloss.Color
	Texto, Sub, Dim                         lipgloss.Color
	Acao, Ok, Espera, Andamento, Erro, Info lipgloss.Color
}

var T = Tokens{
	Base: "#282a36", Mantle: "#21222c", Surface: "#343746", Sel: "#44475a",
	Texto: "#f8f8f2", Sub: "#bfbfd0", Dim: "#6272a4",
	Acao: "#bd93f9", Ok: "#50fa7b", Espera: "#ffb86c", Andamento: "#f1fa8c", Erro: "#ff5555", Info: "#8be9fd",
}
