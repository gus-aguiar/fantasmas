package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestModoDemoTiraONomeDoCoordenador(t *testing.T) {
	if p := ansi.Strip(modeloMissoes().Render(140, 42)); !strings.Contains(p, "✦ GENGAR") {
		t.Fatalf("fora do demo o rótulo devia ser GENGAR:\n%s", p)
	}
	ModoDemo()
	defer func() { nomeCoordenador = "gengar" }()
	p := ansi.Strip(modeloMissoes().Render(140, 42))
	if strings.Contains(strings.ToLower(p), "gengar") || !strings.Contains(p, "✦ COORDENADOR") {
		t.Fatalf("no demo não devia aparecer gengar:\n%s", p)
	}
	if rotuloFigura(FiguraPadrao) != "fantasma" || rotuloFigura("entei") != "entei" {
		t.Fatalf("figura: %q %q", rotuloFigura(FiguraPadrao), rotuloFigura("entei"))
	}
}
