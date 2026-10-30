package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type runnerPipes struct {
	mu    sync.Mutex
	calls []string
}

func (r *runnerPipes) Run(_ context.Context, argv ...string) (string, error) {
	k := strings.Join(argv, " ")
	r.mu.Lock()
	r.calls = append(r.calls, k)
	r.mu.Unlock()
	switch {
	case strings.HasSuffix(k, "pipeline list --json"):
		return `[{"name":"implementar","arquivo":"/s/implementar.yaml","version":1,"inputs":["card","repos"],"stages":["implementar"],"erros":[]},
{"name":"bugfix","arquivo":"/s/bugfix.yaml","version":1,"inputs":["card"],"stages":["debug","corrigir"],"erros":[]}]`, nil
	case strings.Contains(k, " run new "):
		return "7 2026-10-30-PL-9\n", nil
	}
	return "", nil
}

func estadoPipes() dados.Estado {
	e := fixture()
	e.PipesLigado = true
	e.Runs = EstadoDemo(time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)).Runs
	e.Proximo = map[string]string{"2026-10-30-PL-1203-homolog": "lançaria (homolog, skill homolog)"}
	return e
}

func modeloPipes(r *runnerPipes) Model {
	e := estadoPipes()
	a := Acoes{Pipes: &dados.Pipes{R: r, Raiz: "/r", Script: "/r/ddb.py"}}
	m := Novo(func(context.Context) dados.Estado { return e }, a, time.Second).ComEstado(e).ComGaveta("pipes")
	m.largura, m.altura = 140, 42
	return m
}

func teclaMsg(m Model, msg tea.KeyMsg) Model {
	n, _ := m.Update(msg)
	return n.(Model)
}

func digitaPipe(m Model, s string) Model {
	return teclaMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

func TestGavetaPipesAntesDePRs(t *testing.T) {
	ids := strings.Join(IDsGavetas(), ",")
	if ids != "missoes,agora,crew,pipes,prs,ci,ambiente,maquina,mais" {
		t.Fatalf("ordem=%s", ids)
	}
	if Gavetas[indiceGaveta("pipes")].Tecla != "4" || Gavetas[indiceGaveta("mais")].Tecla != "9" {
		t.Fatal("atalhos não deslocados")
	}
}

func TestPipesSemChaveMostraMensagem(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComGaveta("pipes")
	if out := ansi.Strip(m.Render(96, 30)); !strings.Contains(out, "pipes: ligue FANTASMAS_FONTE=ddb") {
		t.Fatalf("sem mensagem:\n%s", out)
	}
}

func TestPipesLinhaEmTodasLarguras(t *testing.T) {
	for _, w := range larguras {
		m := modeloPipes(&runnerPipes{})
		out := m.Render(w, 30)
		ls := strings.Split(out, "\n")
		if len(ls) != 30 {
			t.Fatalf("w=%d: %d linhas", w, len(ls))
		}
		for i, l := range ls {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
		p := ansi.Strip(out)
		if w >= 46 && !strings.Contains(p, "✓") {
			t.Errorf("w=%d: sem chip done:\n%s", w, p)
		}
		if w >= 96 {
			for _, s := range []string{"implementar-homologar", "espera você: 1", "▲", "●", "○"} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d: falta %q:\n%s", w, s, p)
				}
			}
		}
	}
}

func TestPipesDetalheEmTodasLarguras(t *testing.T) {
	for _, w := range larguras {
		m := modeloPipes(&runnerPipes{})
		m.largura = w
		m = tecla(m, "enter")
		if m.pipe == 0 {
			t.Fatalf("w=%d: enter não abriu o detalhe", w)
		}
		out := m.Render(w, 40)
		for i, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
		p := ansi.Strip(out)
		if w >= 70 {
			for _, s := range []string{"implementar", "homologar", "após implementar", "gate:plano", "lançaria", BotaoAba} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d: falta %q:\n%s", w, s, p)
				}
			}
		}
		m = tecla(m, "esc")
		if m.pipe != 0 {
			t.Fatalf("w=%d: esc não fechou", w)
		}
	}
}

func TestPipesASTPedemConfirmacao(t *testing.T) {
	for _, c := range []struct{ k, acao string }{{"A", "advance"}, {"S", "skip"}, {"T", "retry"}} {
		r := &runnerPipes{}
		m := tecla(modeloPipes(r), "enter")
		m = tecla(m, c.k)
		if len(r.calls) != 0 {
			t.Fatalf("%s: rodou sem confirmar: %v", c.k, r.calls)
		}
		if !strings.Contains(m.status, c.k+" de novo") {
			t.Fatalf("%s: status=%q", c.k, m.status)
		}
		m = tecla(m, "j")
		m = tecla(m, c.k)
		if len(r.calls) != 0 {
			t.Fatalf("%s: mover a seleção deveria desarmar: %v", c.k, r.calls)
		}
		m = tecla(m, c.k)
		id := m.runAberto().ID
		quer := "python3 /r/ddb.py --raiz /r stage " + fmt.Sprint(id) + ":homologar " + c.acao
		if len(r.calls) == 0 || r.calls[0] != quer {
			t.Fatalf("%s: calls=%v quer %q", c.k, r.calls, quer)
		}
	}
}

func TestPipesFormularioMontaComando(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(modeloPipes(r), "n")
	if m.form == nil {
		t.Fatal("n não abriu o formulário")
	}
	if p := ansi.Strip(m.Render(140, 42)); !strings.Contains(p, "implementar") || !strings.Contains(p, "bugfix") {
		t.Fatalf("lista de pipelines ausente:\n%s", p)
	}
	m = tecla(m, "enter")
	m = digitaPipe(m, "PL-9")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "WS React")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "Filtro")
	m = tecla(m, "enter")
	quer := "python3 /r/ddb.py --raiz /r run new implementar --input card=PL-9 --input repos=WS React --title Filtro"
	if r.calls[len(r.calls)-1] != quer {
		t.Fatalf("calls=%v", r.calls)
	}
	if m.form != nil {
		t.Fatal("formulário não fechou depois de lançar")
	}
	if !strings.Contains(m.toast, "run 7") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestPipesFormularioRecusaEntradaVazia(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(tecla(modeloPipes(r), "n"), "enter")
	m = tecla(m, "enter")
	m = tecla(m, "enter")
	m = tecla(m, "enter")
	for _, c := range r.calls {
		if strings.Contains(c, "run new") {
			t.Fatalf("lançou com entrada vazia: %v", r.calls)
		}
	}
	if m.form == nil {
		t.Fatal("formulário fechou")
	}
}

func TestPipesNMaiusculoPedeConfirmacaoEChamaGengar(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(modeloPipes(r), "N")
	if m.form != nil || !strings.Contains(m.status, "N de novo") {
		t.Fatalf("primeiro N deveria só armar: form=%v status=%q", m.form, m.status)
	}
	m = tecla(m, "N")
	if m.form == nil {
		t.Fatal("segundo N não abriu o formulário")
	}
	m = tecla(m, "enter")
	m = digitaPipe(m, "PL-9")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "WS React")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = tecla(m, "enter")
	n := len(r.calls)
	if n < 2 {
		t.Fatalf("calls=%v", r.calls)
	}
	if !strings.Contains(r.calls[n-2], " run new implementar ") {
		t.Fatalf("run new ausente: %v", r.calls)
	}
	if quer := "bash /r/scripts/delegate/gengar-spawn.sh 7 --assunto PL-9"; r.calls[n-1] != quer {
		t.Fatalf("calls=%v quer %q", r.calls, quer)
	}
	if !strings.Contains(m.toast, "gengar") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestPipesNMinusculoNaoChamaGengar(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(modeloPipes(r), "n")
	m = tecla(m, "enter")
	m = digitaPipe(m, "PL-9")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "WS")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = tecla(m, "enter")
	for _, c := range r.calls {
		if strings.Contains(c, "gengar-spawn") {
			t.Fatalf("n não deveria abrir gengar: %v", r.calls)
		}
	}
}

func TestPipesNMoverDesarma(t *testing.T) {
	m := tecla(modeloPipes(&runnerPipes{}), "N")
	m = tecla(m, "j")
	m = tecla(m, "N")
	if m.form != nil {
		t.Fatal("mover entre os dois N deveria desarmar")
	}
}

func TestPipesLinhaMostraGengarEProntidao(t *testing.T) {
	for _, w := range larguras {
		m := modeloPipes(&runnerPipes{})
		p := ansi.Strip(m.Render(w, 40))
		if w >= 46 {
			for _, s := range []string{"g-multas", "aguardando prontidão", "g-filtro-frota", "✓ pronto"} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d: falta %q:\n%s", w, s, p)
				}
			}
		}
		for i, l := range strings.Split(m.Render(w, 40), "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
	}
}
