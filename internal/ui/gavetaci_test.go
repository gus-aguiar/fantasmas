package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// corpoPRs monta só a gaveta PRs (agora com o CI junto), sem a descrição da aba.
func corpoPRs(e dados.Estado, w int) (*corpo, []string) {
	c := &corpo{largura: w}
	gavetaPRs(c, e)
	var out []string
	for _, l := range c.linhas {
		out = append(out, ansi.Strip(l.desenha(T.Base)))
	}
	return c, out
}

func TestPRsComCITudoVerde(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{{Repo: "PrologWebReact", Number: 1, CI: "SUCCESS", Motivo: "aprovado · mergear"}, {Repo: "PrologFlutter", Number: 2, CI: "SUCCESS", Motivo: "aprovado · mergear"}}}
	_, ls := corpoPRs(e, 120)
	if !strings.HasPrefix(ls[0], "PRs 2  CI: ✓ 2") {
		t.Errorf("título: %q", ls[0])
	}
	for _, l := range ls[2:] {
		if !strings.Contains(l, "✓ verde") {
			t.Errorf("linha sem CI verde: %q", l)
		}
	}
}

func TestPRsSemPR(t *testing.T) {
	_, ls := corpoPRs(dados.Estado{}, 70)
	// título, branco e a caixa do vazio: borda, 2 de folga, texto, 2 de folga, borda
	if len(ls) != 9 || ls[0] != "PRs" || !strings.Contains(ls[2], "╭") || !strings.Contains(ls[5], "○  nenhum PR aberto") || !strings.Contains(ls[8], "╯") {
		t.Errorf("%q", ls)
	}
}

func TestPRsComCIVermelhoRodandoEVerdes(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{
		{Repo: "PrologWebReact", Number: 3650, URL: "u/3650", CI: "FAILURE", Falhos: []string{"lint", "vitest"}, Motivo: "CI vermelho"},
		{Repo: "PrologWebService", Number: 812, URL: "u/812", CI: "PENDING", Motivo: "aguardando revisão"},
		{Repo: "PrologFlutter", Number: 3, CI: "SUCCESS", Motivo: "aprovado · mergear"},
		{Repo: "PrologE2E", Number: 4, Motivo: "aguardando revisão"},
	}}
	c, ls := corpoPRs(e, 120)
	if !strings.HasPrefix(ls[0], "PRs 4  CI: ✗ 1 · ◐ 1 · ✓ 1 · – 1") {
		t.Errorf("título: %q", ls[0])
	}
	quer := map[string]string{"WebReact #3650": "✗ lint, vitest", "WebService #812": "◐ rodando", "Flutter #3": "✓ verde", "E2E #4": "– sem CI"}
	for nome, ci := range quer {
		achou := false
		for _, l := range ls {
			if strings.Contains(l, nome) && strings.Contains(l, ci) {
				achou = true
			}
		}
		if !achou {
			t.Errorf("%s sem %q:\n%s", nome, ci, strings.Join(ls, "\n"))
		}
	}
	if c.alvos[0].url != "u/3650" {
		t.Errorf("enter deveria abrir o PR: %+v", c.alvos[0])
	}
}

func TestPRsComCICabe(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{
		{Repo: "PrologWebReact", Number: 3650, CI: "FAILURE", Falhos: []string{"lint-com-nome-muito-comprido", "vitest-e-os-testes-de-integracao-do-modulo"}, Motivo: "CI vermelho", Title: "um título bem comprido para estourar a largura da gaveta em telas estreitas"},
		{Repo: "PrologWebService", Number: 812, CI: "PENDING", Motivo: "aguardando revisão"},
		{Repo: "PrologFlutter", Number: 3, CI: "SUCCESS"},
		{Repo: "PrologE2E", Number: 4},
	}}
	for _, w := range larguras {
		_, ls := corpoPRs(e, w)
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

func TestPRsColunasNuncaEncostam(t *testing.T) {
	e := dados.Estado{Travados: []dados.Travado{
		{Repo: "GlorpDesignSystem", Number: 605, CI: "SUCCESS", Motivo: "aprovado · mergear", Title: "Token de borda"},
		{Repo: "PrologWebService", Number: 4202, CI: "PENDING", Motivo: "aguardando revisão", Title: "Endpoint"},
	}}
	for _, w := range []int{90, 120} {
		_, ls := corpoPRs(e, w)
		for _, l := range ls[2:] {
			for _, colado := range []string{"#605aprovado", "#605a", "…◐", "…✓", "revisão◐", "mergear✓"} {
				if strings.Contains(l, colado) {
					t.Errorf("w=%d: colunas encostadas (%q): %q", w, colado, l)
				}
			}
		}
	}
}
