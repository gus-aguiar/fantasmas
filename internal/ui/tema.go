package ui

import (
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

type Tokens struct {
	Base, Mantle, Surface, Sel              lipgloss.Color
	Borda                                   lipgloss.Color
	Texto, Sub, Dim                         lipgloss.Color
	Acao, Ok, Espera, Andamento, Erro, Info lipgloss.Color
	Chip, ChipTexto                         lipgloss.Color
	ChipVerde, ChipVerdeTexto               lipgloss.Color
	BarraCheia, BarraVazia                  lipgloss.Color
	CtxBaixo, CtxMedio, CtxAlto             lipgloss.Color
	MascoteCorpo, MascoteOlho, MascoteBoca  lipgloss.Color
}

var T = Tokens{
	Base: "#2c2244", Mantle: "#241b39", Surface: "#372a54", Sel: "#47366a",
	Borda: "#5f4890",
	Texto: "#f7f2ff", Sub: "#d8ccf0", Dim: "#a99ad2",
	Acao: "#c4a6ff", Ok: "#7ee7a8", Espera: "#ff8a5c", Andamento: "#ffd166", Erro: "#ff6fbf", Info: "#93b4ff",
	Chip: "#4c3580", ChipTexto: "#ede4ff",
	ChipVerde: "#1d5a4c", ChipVerdeTexto: "#b9f6dd",
	BarraCheia: "#c4a6ff", BarraVazia: "#5f4890",
	CtxBaixo: "#7ee7a8", CtxMedio: "#ffd166", CtxAlto: "#ff6fbf",
	MascoteCorpo: "#a66bff", MascoteOlho: "#ff3388", MascoteBoca: "#4b2a8a",
}

// corCtx escolhe a cor da barra de contexto: verde, amarelo e rosa conforme o uso sobe.
func corCtx(pct int) lipgloss.Color {
	switch {
	case pct >= 80:
		return T.CtxAlto
	case pct >= 50:
		return T.CtxMedio
	}
	return T.CtxBaixo
}

// corChipEtapa devolve fundo e texto do chip de cada etapa.
func corChipEtapa(etapa string) (lipgloss.Color, lipgloss.Color) {
	if etapa == "homologar" {
		return T.ChipVerde, T.ChipVerdeTexto
	}
	return T.Chip, T.ChipTexto
}

func rgb(c lipgloss.Color) (r, g, b float64) {
	s := string(c)
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	return float64(v>>16&0xff) / 255, float64(v>>8&0xff) / 255, float64(v&0xff) / 255
}

func luminancia(c lipgloss.Color) float64 {
	lin := func(x float64) float64 {
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, b := rgb(c)
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contraste devolve a razão WCAG 2.x entre duas cores.
func contraste(a, b lipgloss.Color) float64 {
	la, lb := luminancia(a), luminancia(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// matiz devolve matiz (graus) e saturação HSL de uma cor.
func matiz(c lipgloss.Color) (h, s float64) {
	r, g, b := rgb(c)
	mx, mn := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l := (mx + mn) / 2
	d := mx - mn
	if d == 0 {
		return 0, 0
	}
	s = d / (1 - math.Abs(2*l-1))
	switch mx {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s
}
