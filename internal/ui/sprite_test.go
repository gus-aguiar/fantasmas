package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fantasmas/assets"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func gifSintetico(t *testing.T) []byte {
	pal := color.Palette{color.Transparent, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}}
	q1 := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	q2 := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	for y := 1; y < 3; y++ {
		for x := 1; x < 3; x++ {
			q1.SetColorIndex(x, y, 1)
			q2.SetColorIndex(x, y, 2)
		}
	}
	var buf bytes.Buffer
	g := &gif.GIF{Image: []*image.Paletted{q1, q2}, Delay: []int{5, 0}, Disposal: []byte{gif.DisposalBackground, gif.DisposalBackground}}
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSpriteSintetico(t *testing.T) {
	s, err := CarregarSprite(gifSintetico(t), 16, 8, T.Mantle)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Quadros) != 2 {
		t.Fatalf("quadros %d", len(s.Quadros))
	}
	if s.Colunas != 2 || s.Linhas != 1 {
		t.Fatalf("recorte %dx%d, quero 2x1", s.Colunas, s.Linhas)
	}
	if s.Delays[0] != 50*time.Millisecond || s.Delays[1] != 100*time.Millisecond {
		t.Fatalf("delays %v", s.Delays)
	}
	for _, q := range s.Quadros {
		for _, l := range q {
			if lipgloss.Width(l) != s.Colunas {
				t.Fatalf("largura %d", lipgloss.Width(l))
			}
		}
	}
	if s.Quadro(0)[0] == s.Quadro(1)[0] || s.Quadro(2)[0] != s.Quadro(0)[0] {
		t.Fatal("quadros não alternam")
	}
}

func TestSpriteReal(t *testing.T) {
	s, err := CarregarSprite(assets.Gengar, 24, 10, T.Mantle)
	if err != nil {
		t.Fatal(err)
	}
	if s.Colunas > 24 || s.Linhas > 10 || len(s.Quadros) < 2 {
		t.Fatalf("%dx%d, %d quadros", s.Colunas, s.Linhas, len(s.Quadros))
	}
	for _, q := range s.Quadros {
		if len(q) != s.Linhas {
			t.Fatalf("linhas %d", len(q))
		}
		for _, l := range q {
			if lipgloss.Width(l) != s.Colunas {
				t.Fatalf("largura %d", lipgloss.Width(l))
			}
		}
	}
}

func TestSpriteInvalidoUsaFallback(t *testing.T) {
	if _, err := CarregarSprite([]byte("nada"), 16, 8, T.Mantle); err == nil {
		t.Fatal("esperava erro")
	}
	antes := spriteFn
	defer func() { spriteFn = antes }()
	spriteFn = func(int) *Sprite { s, _ := CarregarSprite([]byte("nada"), 16, 8, T.Mantle); return s }
	topo := Topo(estadoEx(), 96, 0, false, "")
	if n := len(strings.Split(topo, "\n")); n != MascoteLinhas {
		t.Fatalf("fallback com %d linhas", n)
	}
	if !strings.Contains(topo, QuadroMascote(0, T.Mantle)[1]) || passoDe(0) != passoQuadro {
		t.Fatal("fallback não usou o mascote nativo")
	}
}

// gifOlho: 32x8 roxo; a primeira célula (4x8 px) tem só 4 px vermelhos (12,5%: vence com 0,10, perderia com 0,15).
func gifOlho(t *testing.T) []byte {
	roxo, verm := color.RGBA{110, 60, 150, 255}, color.RGBA{230, 20, 30, 255}
	pal := color.Palette{color.Transparent, roxo, verm}
	q := image.NewPaletted(image.Rect(0, 0, 32, 8), pal)
	for y := 0; y < 8; y++ {
		for x := 0; x < 32; x++ {
			q.SetColorIndex(x, y, 1)
		}
	}
	for _, p := range [][2]int{{1, 1}, {2, 1}, {1, 2}, {2, 2}} {
		q.SetColorIndex(p[0], p[1], 2)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{Image: []*image.Paletted{q}, Delay: []int{5}}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDestaqueVenceMaioria(t *testing.T) {
	s, err := CarregarSprite(gifOlho(t), 8, 1, T.Mantle)
	if err != nil {
		t.Fatal(err)
	}
	if s.Colunas != 8 || s.Linhas != 1 {
		t.Fatalf("%dx%d", s.Colunas, s.Linhas)
	}
	l := s.Quadro(0)[0]
	if !strings.Contains(l, "48;2;230;20;30m") || !strings.Contains(l, "48;2;110;60;150m") {
		t.Fatalf("célula do olho não ficou vermelha: %q", l)
	}
}

func TestCentraCadaQuadro(t *testing.T) {
	pal := color.Palette{color.Transparent, color.RGBA{255, 0, 0, 255}}
	var qs []*image.Paletted
	for _, x0 := range []int{0, 12} {
		q := image.NewPaletted(image.Rect(0, 0, 16, 4), pal)
		for y := 0; y < 4; y++ {
			for x := x0; x < x0+4; x++ {
				q.SetColorIndex(x, y, 1)
			}
		}
		qs = append(qs, q)
	}
	var buf bytes.Buffer
	g := &gif.GIF{Image: qs, Delay: []int{5, 5}, Disposal: []byte{gif.DisposalBackground, gif.DisposalBackground}}
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	s, err := CarregarSprite(buf.Bytes(), 16, 8, T.Mantle)
	if err != nil {
		t.Fatal(err)
	}
	if s.Colunas != 4 || s.Quadro(0)[0] != s.Quadro(1)[0] {
		t.Fatalf("quadro não centrado: %d colunas", s.Colunas)
	}
}

func TestUsarSpriteEFallback(t *testing.T) {
	antesT, antesE := gifTopo, gifEmbutido
	defer func() { gifTopo, gifEmbutido = antesT, antesE; limparSprites() }()
	dir := t.TempDir()
	bom, ruim := filepath.Join(dir, "a.gif"), filepath.Join(dir, "b.gif")
	os.WriteFile(bom, gifSintetico(t), 0o644)
	os.WriteFile(ruim, []byte("nada"), 0o644)

	if err := UsarSprite(bom); err != nil {
		t.Fatal(err)
	}
	if s := spriteTopo(96); s == nil || s.Colunas != 2 {
		t.Fatal("não carregou o gif do caminho")
	}
	if err := UsarSprite(ruim); err == nil {
		t.Fatal("esperava erro")
	}
	if err := UsarSprite(filepath.Join(dir, "nao-existe.gif")); err == nil {
		t.Fatal("esperava erro de leitura")
	}
	if s := spriteTopo(96); s == nil || s.Colunas < 10 {
		t.Fatal("erro não voltou ao embutido")
	}
	gifEmbutido = []byte("nada")
	if err := UsarSprite(ruim); err == nil {
		t.Fatal("esperava erro")
	}
	if spriteTopo(96) != nil || passoDe(0) != passoQuadro {
		t.Fatal("sem embutido devia cair no mascote nativo")
	}
}

func TestStatusInicial(t *testing.T) {
	m := Novo(nil, Acoes{}, time.Second).ComStatus("sprite: falhou")
	if !strings.Contains(ansi.Strip(strings.Join(soTopo(m, 96), "\n")), "sprite: falhou") {
		t.Fatal("erro do sprite não aparece no status")
	}
}
