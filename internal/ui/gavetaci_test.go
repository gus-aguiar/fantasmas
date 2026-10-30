package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func corpoCI(e dados.Estado, w int) (*corpo, []string) {
	c := &corpo{largura: w}
	gavetaCI(c, e)
	var out []string
	for _, l := range c.linhas {
		out = append(out, ansi.Strip(l.desenha(T.Base)))
	}
	return c, out
}

func TestGavetaCITudoVerde(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{{Repo: "PrologWebReact", Number: 1, CI: "SUCCESS"}, {Repo: "PrologFlutter", Number: 2, CI: "SUCCESS"}}}
	_, ls := corpoCI(e, 70)
	if len(ls) != 2 || strings.TrimSpace(ls[1]) != "✓ CI: tudo verde" {
		t.Errorf("%q", ls)
	}
}

func TestGavetaCISemPR(t *testing.T) {
	_, ls := corpoCI(dados.Estado{}, 70)
	if len(ls) != 2 || !strings.Contains(ls[1], "sem PR aberto") {
		t.Errorf("%q", ls)
	}
}

func TestGavetaCIVermelhoRodandoEVerdes(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{
		{Repo: "PrologWebReact", Number: 3650, URL: "u/3650", CI: "FAILURE", Falhos: []string{"lint", "vitest"}},
		{Repo: "PrologWebService", Number: 812, URL: "u/812", CI: "PENDING"},
		{Repo: "PrologFlutter", Number: 3, CI: "SUCCESS"},
		{Repo: "PrologE2E", Number: 4},
	}}
	c, ls := corpoCI(e, 70)
	quer := []string{"✗ WebReact #3650  vermelho · lint, vitest", "◐ WebService #812  rodando", "✓ 1 PR verde", "1 sem CI"}
	if len(ls) != 5 {
		t.Fatalf("%q", ls)
	}
	for i, q := range quer {
		if strings.TrimSpace(ls[i+1]) != q {
			t.Errorf("linha %d: %q quer %q", i+1, ls[i+1], q)
		}
	}
	if c.alvos[0].url != "u/3650/checks" {
		t.Errorf("alvo %+v", c.alvos[0])
	}
}

func TestGavetaCICabe(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{
		{Repo: "PrologWebReact", Number: 3650, CI: "FAILURE", Falhos: []string{"lint-com-nome-muito-comprido", "vitest-e-os-testes-de-integracao-do-modulo"}},
		{Repo: "PrologWebService", Number: 812, CI: "PENDING"},
		{Repo: "PrologFlutter", Number: 3, CI: "SUCCESS"},
		{Repo: "PrologE2E", Number: 4},
	}}
	for _, w := range larguras {
		_, ls := corpoCI(e, w)
		for _, l := range ls {
			if lipgloss.Width(l) > w {
				t.Errorf("w=%d largura %d: %q", w, lipgloss.Width(l), l)
			}
		}
	}
}

func TestDemoVerdeSoTemVerdes(t *testing.T) {
	r := dados.ResumoDoCI(EstadoDemoVerde(time.Now()).Travados, true)
	if len(r.Falhos) != 0 || len(r.Rodando) != 0 || r.SemCI != 0 || r.Verdes == 0 {
		t.Errorf("%+v", r)
	}
	if r := dados.ResumoDoCI(EstadoDemo(time.Now()).Travados, false); len(r.Falhos) != 1 || len(r.Falhos[0].Falhos) != 2 {
		t.Errorf("demo vermelha: %+v", r)
	}
}
