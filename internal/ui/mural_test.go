package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func avisosEx() []dados.Aviso {
	h := time.Date(2026, 10, 30, 10, 12, 0, 0, time.Local)
	return []dados.Aviso{
		{At: h, Autor: "gengar-1030-0900", Texto: "vou mexer no delegate-spawn até 12h"},
		{At: h.Add(20 * time.Minute), Autor: "humano-com-nome-bem-comprido", Texto: "HML fora até 15h — não rodem homologação 🚀 nem suíte E2E contra o ambiente até eu avisar de novo aqui"},
	}
}

func TestMuralAusenteSemFaixa(t *testing.T) {
	e := dados.Estado{At: time.Now()}
	if s := textoAgora(e); strings.Contains(s, "mural") || strings.Contains(s, "◆") {
		t.Fatalf("sem aviso a faixa não aparece:\n%s", s)
	}
	e.Mural = avisosEx()
	s := textoAgora(e)
	if !strings.Contains(s, "mural") || !strings.Contains(s, "◆ gengar-1030-0900 · 10:12  vou mexer") {
		t.Fatalf("faixa faltando:\n%s", s)
	}
	if strings.Index(s, "mural") > strings.Index(s, "precisa de você") {
		t.Fatalf("mural vai antes de 'precisa de você':\n%s", s)
	}
}

func TestFaixaMuralCabe(t *testing.T) {
	for _, w := range larguras {
		for _, a := range avisosEx() {
			s := renderAviso(a, w, T.Base)
			if lipgloss.Width(s) > w {
				t.Fatalf("w=%d largura %d: %q", w, lipgloss.Width(s), ansi.Strip(s))
			}
		}
	}
}
