package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var agoraMissoes = time.Date(2026, 10, 30, 13, 0, 0, 0, time.UTC)

func modeloMissoes() Model {
	m := modelo(EstadoDemo(agoraMissoes), &registro{}).ComGaveta("missoes")
	m.largura, m.altura = 140, 42
	return m
}

func TestGavetaMissoesEhAPrimeira(t *testing.T) {
	if Gavetas[0].ID != "missoes" || Gavetas[0].Tecla != "1" {
		t.Fatalf("primeira=%+v", Gavetas[0])
	}
}

func TestCartaEmTodasAsLarguras(t *testing.T) {
	for w := 30; w <= 180; w += 7 {
		m := modeloMissoes()
		m.largura = w
		out := m.Render(w, 42)
		for i, l := range strings.Split(out, "\n") {
			if lw := lipgloss.Width(l); lw != w {
				t.Fatalf("w=%d linha %d com %d", w, i, lw)
			}
		}
		p := ansi.Strip(out)
		if !strings.Contains(p, "Frota") || !strings.Contains(p, "Genguinha") {
			t.Fatalf("w=%d sem missão ou principal:\n%s", w, p)
		}
		if w >= 96 {
			for _, s := range []string{"Frota por unidade", "g-frota", "PL-1203", "trabalhando", "PR vermelho", "gengar-1030-0900"} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d sem %q:\n%s", w, s, p)
				}
			}
		}
	}
}

func TestEnterFocaMissaoEFiltraGavetas(t *testing.T) {
	m := modeloMissoes()
	m = tecla(m, "enter")
	if m.missao == "" {
		t.Fatal("enter deveria focar a missão")
	}
	nome := nomeMissao(m.estado, m.missao)
	if Gavetas[m.gaveta].ID != "agora" {
		t.Fatalf("enter deveria ir ao agora, veio %s", Gavetas[m.gaveta].ID)
	}
	p := ansi.Strip(m.Render(140, 42))
	if !strings.Contains(p, "◆ "+nome) {
		t.Fatalf("sem selo:\n%s", p)
	}
	m = m.irGaveta(indiceGaveta("crew"))
	p = ansi.Strip(m.Render(140, 42))
	dentro, fora := 0, 0
	for _, l := range m.estado.Linhas {
		if strings.Contains(p, l.Nome) {
			if l.Missao == m.missao {
				dentro++
			} else {
				fora++
			}
		}
	}
	if dentro == 0 || fora != 0 {
		t.Fatalf("filtro não aplicado (dentro=%d fora=%d):\n%s", dentro, fora, p)
	}
	m = m.irGaveta(indiceGaveta("pipes"))
	for _, r := range FiltrarMissao(m.estado, m.missao).Runs {
		if r.Missao != m.missao {
			t.Fatalf("run de outra missão: %+v", r)
		}
	}
}

func TestEscEZeroLimpamMissao(t *testing.T) {
	for _, k := range []string{"esc", "0"} {
		m := tecla(modeloMissoes(), "enter")
		if m.missao == "" {
			t.Fatal("sem foco")
		}
		m = tecla(m, k)
		if m.missao != "" {
			t.Fatalf("%s não limpou", k)
		}
		if strings.Contains(ansi.Strip(m.Render(140, 42)), "◆ Frota") {
			t.Fatalf("%s: selo ficou", k)
		}
	}
}

func TestGVaiAAbaDoGengarDaMissao(t *testing.T) {
	r := &registro{}
	m := modelo(EstadoDemo(agoraMissoes), r).ComGaveta("missoes")
	m.largura, m.altura = 140, 42
	tecla(m, "g")
	if len(r.abas) != 1 || !strings.HasPrefix(r.abas[0], "w15:") {
		t.Fatalf("abas=%v", r.abas)
	}
}

func TestCorEstavelPorSlug(t *testing.T) {
	if corMissao("frota-por-unidade") != corMissao("frota-por-unidade") {
		t.Fatal("instável")
	}
	vistas := map[lipgloss.Color]bool{}
	for _, s := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "frota", "multas"} {
		vistas[corMissao(s)] = true
	}
	if len(vistas) < 4 {
		t.Fatalf("pouca variação: %v", vistas)
	}
	ms := EstadoDemo(agoraMissoes).Missoes
	cores := map[lipgloss.Color]string{}
	for _, x := range ms {
		if o, ok := cores[corMissao(x.Slug)]; ok {
			t.Errorf("demo: %s e %s com a mesma cor", o, x.Slug)
		}
		cores[corMissao(x.Slug)] = x.Slug
	}
}

func TestMissoesSemChaveMostraMensagem(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComGaveta("missoes")
	if out := ansi.Strip(m.Render(96, 30)); !strings.Contains(out, "missões: ligue FANTASMAS_FONTE=ddb") {
		t.Fatalf("sem mensagem:\n%s", out)
	}
}

func TestFiltroPorWorkspaceSemDdb(t *testing.T) {
	e := dados.Estado{At: agoraMissoes, Linhas: []dados.LinhaViva{
		{Nome: "c-a", Workspace: "w15", Estado: dados.Trabalhando},
		{Nome: "c-b", Workspace: "w1", Estado: dados.Trabalhando},
	}}
	f := filtrarPor(e, "", "w15")
	if len(f.Linhas) != 1 || f.Linhas[0].Nome != "c-a" {
		t.Fatalf("linhas=%+v", f.Linhas)
	}
}

// Valores calculados em Python com o fnv1a32 do painel (painel.py: utf-8, 32 bits, ACENTOS[h % 6]).
func TestCorMissaoIdenticaAoPainel(t *testing.T) {
	casos := map[string]string{
		"multas":             "#6ee7b7", // fnv1a32 = 2612968891, índice 1
		"economia-de-multas": "#f59ad8", // fnv1a32 = 743490299, índice 5
		"frota-por-unidade":  "#67d4f0", // fnv1a32 = 2777837278, índice 4
	}
	for slug, esperada := range casos {
		if got := string(corMissao(slug)); got != esperada {
			t.Errorf("corMissao(%q) = %s, painel usa %s", slug, got, esperada)
		}
	}
	painel := []string{"#b69cff", "#6ee7b7", "#fbbf5a", "#ff8f80", "#67d4f0", "#f59ad8"}
	for i, c := range T.Acentos {
		if string(c) != painel[i] {
			t.Errorf("Acentos[%d] = %s, painel usa %s", i, c, painel[i])
		}
	}
}
