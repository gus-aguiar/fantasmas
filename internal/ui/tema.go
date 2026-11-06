package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

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
	// Acentos dão a cada missão uma cor estável (ver corMissao).
	Acentos [6]lipgloss.Color
}

// T é o tema em uso; UsarTema troca.
var T = temaRoxo

// NomesTemas na ordem da aba configurações; o primeiro é o padrão.
var NomesTemas = []string{"roxo", "orca", "terminal"}

var temas = map[string]*Tokens{"roxo": &temaRoxo, "orca": &temaOrca, "terminal": &temaTerminal}

// UsarTema troca as cores do app ("" = roxo). O sprite em células guarda o fundo, então o cache é limpo.
func UsarTema(nome string) error {
	if nome == "" {
		nome = NomesTemas[0]
	}
	t, ok := temas[nome]
	if !ok {
		return fmt.Errorf("tema %q desconhecido; válidos: %s", nome, strings.Join(NomesTemas, ", "))
	}
	T = *t
	limparSprites()
	return nil
}

// temaOrca: grafite neutro, para combinar com o terminal do Orca; acentos e mascote iguais aos do roxo.
var temaOrca = Tokens{
	Base: "#1f2125", Mantle: "#17181b", Surface: "#2a2d32", Sel: "#363a41",
	Borda: "#4b5059",
	Texto: "#eceef2", Sub: "#c4c8d0", Dim: "#8d939d",
	Acao: "#8fb4ff", Ok: "#7ee7a8", Espera: "#ff8a5c", Andamento: "#ffd166", Erro: "#ff7a9e", Info: "#7fd0f0",
	Chip: "#30384a", ChipTexto: "#e3e8f2",
	ChipVerde: "#1d4a3f", ChipVerdeTexto: "#b9f6dd",
	BarraCheia: "#8fb4ff", BarraVazia: "#4b5059",
	CtxBaixo: "#7ee7a8", CtxMedio: "#ffd166", CtxAlto: "#ff7a9e",
	MascoteCorpo: "#a66bff", MascoteOlho: "#ff3388", MascoteBoca: "#4b2a8a",
	Acentos: temaRoxo.Acentos,
}

// temaTerminal não pinta fundo ("" = cor padrão do terminal) e usa a paleta ANSI do próprio terminal.
// Seleção e chips usam cinzas escuros do 256 (232–255), que quase todo terminal mantém fixos:
// sobre o ANSI 8 (cinza médio) o texto colorido some.
var temaTerminal = Tokens{
	Base: "", Mantle: "", Surface: "", Sel: "237",
	Borda: "8",
	Texto: "", Sub: "", Dim: "7",
	Acao: "5", Ok: "2", Espera: "3", Andamento: "4", Erro: "1", Info: "6",
	Chip: "238", ChipTexto: "",
	ChipVerde: "2", ChipVerdeTexto: "0",
	BarraCheia: "5", BarraVazia: "8",
	CtxBaixo: "2", CtxMedio: "3", CtxAlto: "1",
	MascoteCorpo: "5", MascoteOlho: "1", MascoteBoca: "0",
	Acentos: [6]lipgloss.Color{"13", "10", "11", "9", "14", "12"},
}

var temaRoxo = Tokens{
	Base: "#2c2244", Mantle: "#241b39", Surface: "#372a54", Sel: "#47366a",
	Borda: "#5f4890",
	Texto: "#f7f2ff", Sub: "#d8ccf0", Dim: "#a99ad2",
	Acao: "#c4a6ff", Ok: "#7ee7a8", Espera: "#ff8a5c", Andamento: "#ffd166", Erro: "#ff6fbf", Info: "#93b4ff",
	Chip: "#4c3580", ChipTexto: "#ede4ff",
	ChipVerde: "#1d5a4c", ChipVerdeTexto: "#b9f6dd",
	BarraCheia: "#c4a6ff", BarraVazia: "#5f4890",
	CtxBaixo: "#7ee7a8", CtxMedio: "#ffd166", CtxAlto: "#ff6fbf",
	MascoteCorpo: "#a66bff", MascoteOlho: "#ff3388", MascoteBoca: "#4b2a8a",
	// Os mesmos 6 acentos, na mesma ordem, do painel (ACENTOS em delegate-painel/painel.py): a cor da missão
	// é idêntica nos dois. Todos passam de 4.5:1 sobre Base e Sel.
	Acentos: [6]lipgloss.Color{"#b69cff", "#6ee7b7", "#fbbf5a", "#ff8f80", "#67d4f0", "#f59ad8"},
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

// Tokens semânticos: um acento (foco, aba ativa, seleção) e três estados. Cor só para estado e
// foco; texto secundário vai em Secundario. As views novas usam só estes.
func (t Tokens) Acento() lipgloss.Color      { return t.Acao }
func (t Tokens) EstadoOk() lipgloss.Color    { return t.Ok }
func (t Tokens) EstadoAviso() lipgloss.Color { return t.Espera }
func (t Tokens) EstadoErro() lipgloss.Color  { return t.Erro }
func (t Tokens) Secundario() lipgloss.Color  { return t.Dim }
