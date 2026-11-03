package ui

import (
	"strings"
	"testing"
	"time"
	"unicode"

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
	if !strings.Contains(r, "3") || strings.Contains(r, "fantasminhas") {
		t.Fatalf("%q", r)
	}
}

func TestRodape110TeclaENome(t *testing.T) {
	if !strings.Contains(ansi.Strip(Rodape("agora", 110)), "2 fantasminhas") {
		t.Fatal("sem 2 fantasminhas")
	}
}

func TestRodape96SoNomes(t *testing.T) {
	r := ansi.Strip(Rodape("agora", 96))
	if !strings.Contains(r, "fantasminhas") || !strings.Contains(r, "configurações") {
		t.Fatalf("%q", r)
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
	for _, esperado := range []string{"atualizado 09:05", "r atualiza · p painel · ? ajuda · q sai", "pesados: mvn", "1 precisa de você"} {
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

func TestTopoSemTravaForaDoSpark(t *testing.T) {
	e := estadoEx()
	e.Maquina = []dados.Linha{{Chave: "memoria", Rotulo: "memória", Detalhe: "13 de 31 GB", Extra: "ok"}}
	if out := ansi.Strip(Topo(e, 96, 0, false, "")); !strings.Contains(out, "sem fila de pesados fora do Spark") {
		t.Fatalf("sem trava deveria dizer que está fora do Spark:\n%s", out)
	}
	e.Maquina = nil
	if out := ansi.Strip(Topo(e, 96, 0, false, "")); !strings.Contains(out, "fila de pesados livre") {
		t.Fatalf("antes da primeira leitura segue o livre:\n%s", out)
	}
}

func TestTodaGavetaTemDescricao(t *testing.T) {
	for i, g := range Gavetas {
		r := []rune(g.Descricao)
		if len(r) == 0 || !unicode.IsUpper(r[0]) || strings.HasSuffix(g.Descricao, ".") || len(r) > 48 {
			t.Errorf("%s: descrição %q deveria ser curta, com maiúscula e sem ponto", g.ID, g.Descricao)
			continue
		}
		m := modelo(fixture(), &registro{}).irGaveta(i)
		achou := false
		for _, l := range strings.Split(render(m, 140, 40), "\n") {
			if strings.Contains(l, "› ") && strings.Contains(l, g.Descricao) {
				achou = true
			}
		}
		if !achou {
			t.Errorf("%s: descrição não aparece ao lado de um título de seção:\n%s", g.ID, render(m, 140, 40))
		}
	}
}

func TestVazioEmCaixa(t *testing.T) {
	for _, w := range []int{30, 46, 96, 140} {
		c := &corpo{largura: w}
		c.vazioIcone("✓", T.Ok, "nada pedindo você agora", "segunda linha")
		var ls []string
		for _, l := range c.linhas {
			ls = append(ls, ansi.Strip(l.desenha(T.Base)))
		}
		// borda, 2 de folga, 2 linhas de texto, 2 de folga, borda
		if len(ls) != 8 || !strings.HasPrefix(strings.TrimSpace(ls[0]), "╭") || !strings.HasSuffix(strings.TrimSpace(ls[7]), "╯") {
			t.Fatalf("w=%d: %q", w, ls)
		}
		vazia := func(l string) bool { return strings.TrimSpace(strings.Trim(strings.TrimSpace(l), "│")) == "" }
		if !strings.Contains(ls[3], "✓  nada pedindo") || !vazia(ls[1]) || !vazia(ls[2]) || !vazia(ls[5]) || !vazia(ls[6]) {
			t.Fatalf("w=%d: ícone, texto ou folga: %q", w, ls)
		}
		// largura toda: 1 coluna de margem de cada lado
		if lipgloss.Width(ls[0]) != w-1 || !strings.HasPrefix(ls[0], " ╭") {
			t.Fatalf("w=%d: caixa deveria ocupar a largura toda: %q", w, ls[0])
		}
		for i, l := range ls {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d estoura: %q", w, i, l)
			}
		}
		// centrada: as sobras dos dois lados diferem em no máximo 1
		esq := len(ls[0]) - len(strings.TrimLeft(ls[0], " "))
		if dir := w - lipgloss.Width(ls[0]); esq-dir > 1 || dir-esq > 1 {
			t.Fatalf("w=%d: caixa fora do centro (%d × %d)", w, esq, dir)
		}
	}
}

func TestDescricaoSemItalicoEApagada(t *testing.T) {
	antes := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(antes)
	if T.Dim == T.Sub {
		t.Fatal("Dim e Sub iguais: a descrição não se separa do título")
	}
	c := &corpo{largura: 120, desc: "Seus PRs abertos e o que falta"}
	c.secaoN("PRs", 7, "")
	l := c.linhas[0].desenha(T.Base)
	if want := st(T.Base, T.Dim).Render("  Seus PRs abertos e o que falta"); !strings.Contains(l, want) {
		t.Fatalf("descrição deveria vir só em Dim, sem itálico: %q", l)
	}
	if !strings.Contains(l, st(T.Base, T.Sub).Bold(true).Render("PRs (7)")) {
		t.Fatalf("título fora do estilo: %q", l)
	}
}
