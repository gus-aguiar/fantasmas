package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	m.Run()
}

var larguras = []int{20, 30, 46, 56, 70, 96, 140}

func estadoEx() dados.Estado {
	return dados.Estado{
		At:        time.Date(2026, 10, 16, 9, 5, 0, 0, time.UTC),
		Esperando: []dados.Linha{{Chave: "c-pl-1/q1", Rotulo: "c-pl-1 · plano", Detalhe: "ação — válida…"}},
		Tasks:     []dados.Linha{{Chave: "t1", Rotulo: "ação", Detalhe: "working"}},
		Maquina:   []dados.Linha{{Chave: "trava", Detalhe: "mvn há 3 min · ação—válida", Extra: "aviso"}},
	}
}

func TestQuadroMascote(t *testing.T) {
	for i := 0; i < 16; i++ {
		q := QuadroMascote(i, T.Mantle)
		if len(q) != MascoteLinhas {
			t.Fatalf("quadro %d: %d linhas", i, len(q))
		}
		for _, l := range q {
			if lipgloss.Width(l) != MascoteColunas {
				t.Fatalf("quadro %d: largura %d", i, lipgloss.Width(l))
			}
		}
	}
	if !strings.Contains(ansi.Strip(QuadroMascote(7, T.Mantle)[1]), "─") {
		t.Fatal("quadro 7 sem olhos fechados")
	}
	if strings.Join(QuadroMascote(16, T.Mantle), "|") != strings.Join(QuadroMascote(0, T.Mantle), "|") {
		t.Fatal("quadro 16 != 0")
	}
	_ = QuadroMascote(-3, T.Mantle)
}

func TestCorteLarguraDeCelula(t *testing.T) {
	c := Corte("ação—válida", 5)
	if lipgloss.Width(c) != 5 || !strings.HasSuffix(c, "…") {
		t.Fatalf("corte %q", c)
	}
	if lipgloss.Width(Col("é", 3)) != 3 {
		t.Fatal("col")
	}
	if lipgloss.Width(Col("日本語日本語", 5)) > 5 {
		t.Fatal("largura dupla")
	}
}

func TestRodapeUmaLinhaSempre(t *testing.T) {
	for _, w := range larguras {
		for _, g := range Gavetas {
			r := Rodape(g.ID, w)
			if strings.Contains(r, "\n") || lipgloss.Width(r) > w {
				t.Fatalf("w=%d gaveta=%s largura=%d", w, g.ID, lipgloss.Width(r))
			}
		}
	}
}

func TestRodape56SoTeclaNasInativas(t *testing.T) {
	r := ansi.Strip(Rodape("agora", 56))
	if !strings.Contains(r, "2") || strings.Contains(r, "fantasminhas") {
		t.Fatalf("%q", r)
	}
}

func TestRodape96TeclaENome(t *testing.T) {
	if !strings.Contains(ansi.Strip(Rodape("agora", 96)), "2 fantasminhas") {
		t.Fatal("sem 2 fantasminhas")
	}
}

func TestTopoCabe(t *testing.T) {
	for _, w := range larguras {
		for _, col := range []bool{false, true} {
			for _, st := range []string{"", "entregue a c-pl-1: ação — válida com texto bem longo para estourar"} {
				for _, l := range strings.Split(Topo(estadoEx(), w, 3, col, st), "\n") {
					if lipgloss.Width(l) > w {
						t.Fatalf("w=%d linha %d: %q", w, lipgloss.Width(l), ansi.Strip(l))
					}
				}
			}
		}
	}
}

func TestTopoLinhas(t *testing.T) {
	if n := len(strings.Split(Topo(estadoEx(), 40, 0, false, ""), "\n")); n != 1 {
		t.Fatalf("minimo: %d linhas", n)
	}
	for _, w := range []int{46, 56, 70, 96} {
		alt := max(4, spriteTopo(w).Linhas)
		if n := len(strings.Split(Topo(estadoEx(), w, 0, false, ""), "\n")); n != alt {
			t.Fatalf("w=%d: %d linhas", w, n)
		}
	}
}

func TestTopoConteudo(t *testing.T) {
	s := ansi.Strip(Topo(estadoEx(), 96, 0, false, ""))
	for _, esperado := range []string{"atualizado 09:05", "r atualiza · p painel · ? ajuda · q sai", "pesados: mvn", "1 pra você"} {
		if !strings.Contains(s, esperado) {
			t.Fatalf("falta %q em %q", esperado, s)
		}
	}
	if !strings.Contains(ansi.Strip(Topo(estadoEx(), 96, 0, true, "")), "atualizando…") {
		t.Fatal("sem atualizando")
	}
	if !strings.Contains(ansi.Strip(Topo(estadoEx(), 96, 0, false, "entregue a x")), "entregue a x") {
		t.Fatal("sem status")
	}
}

func TestTopoSemEmoji(t *testing.T) {
	for _, w := range larguras {
		for _, r := range Topo(estadoEx(), w, 0, true, "") + Rodape("agora", w) {
			if r >= 0x1F300 && r <= 0x1FAFF {
				t.Fatalf("emoji %U", r)
			}
		}
	}
}
