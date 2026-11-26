package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/x/ansi"
)

var agoraFila = time.Date(2026, 11, 26, 15, 0, 0, 0, time.UTC)

func modeloFila(responder func(chave, resposta string) (dados.Entrega, error)) Model {
	m := modelo(EstadoDemo(agoraFila), &registro{})
	m.acoes.Responder = responder
	m.largura, m.altura = 120, 40
	return m
}

func telaFila(m Model) string { return ansi.Strip(m.Render(120, 40)) }

func TestFilaMostraUmaDecisaoComRecomendacaoEVision(t *testing.T) {
	p := telaFila(modeloFila(nil))
	for _, s := range []string{"ESPERANDO VOCÊ 2", "Aprovar o plano de migração", "recomenda Aprovar o plano",
		"vision    migração de dados só roda com plano aprovado", "+1 na fila  c-vt-1", "y 2× aprova · s pula · enter responder · g ir à aba"} {
		if !strings.Contains(p, s) {
			t.Errorf("sem %q:\n%s", s, p)
		}
	}
	if strings.Contains(p, "Trocar o Testcontainers") {
		t.Errorf("só a decisão do topo aparece inteira:\n%s", p)
	}
}

func TestFilaSPulaEMostraForaDoVision(t *testing.T) {
	m, _ := apertaTecla(modeloFila(nil), 's')
	p := telaFila(m)
	if !strings.Contains(p, "Trocar o Testcontainers") || !strings.Contains(p, "fora do vision") || !strings.Contains(p, "+1 na fila  c-vt-2") {
		t.Fatalf("depois do s, a segunda decisão vai ao topo:\n%s", p)
	}
	m, _ = apertaTecla(m, 's')
	if p := telaFila(m); !strings.Contains(p, "Aprovar o plano de migração") {
		t.Fatalf("a fila gira e a primeira volta ao topo:\n%s", p)
	}
}

func TestFilaYDuasVezesAprovaARecomendacao(t *testing.T) {
	var chave, resposta string
	m := modeloFila(func(c, r string) (dados.Entrega, error) {
		chave, resposta = c, r
		return dados.Entrega{Crew: "c-vt-2", Resposta: r, Entregue: true}, nil
	})
	m, cmd := apertaTecla(m, 'y')
	if cmd != nil || m.armadoFila != "t-espera/q1" || !strings.Contains(m.status, "y de novo aprova: Aprovar o plano") {
		t.Fatalf("o primeiro y só arma: armado=%q status=%q", m.armadoFila, m.status)
	}
	m, cmd = apertaTecla(m, 'y')
	if cmd == nil {
		t.Fatal("o segundo y devia responder")
	}
	n, _ := m.Update(cmd())
	m = n.(Model)
	if chave != "t-espera/q1" || resposta != "Aprovar o plano" {
		t.Fatalf("respondeu %q com %q", chave, resposta)
	}
	if p := telaFila(m); !strings.Contains(p, "ESPERANDO VOCÊ 1") || !strings.Contains(p, "Trocar o Testcontainers") {
		t.Fatalf("a decisão respondida sai e a próxima sobe:\n%s", p)
	}
}

func TestFilaOutraTeclaDesarmaEYSemRecomendacaoNaoResponde(t *testing.T) {
	chamou := false
	m := modeloFila(func(c, r string) (dados.Entrega, error) { chamou = true; return dados.Entrega{}, nil })
	m, _ = apertaTecla(m, 'y')
	m, _ = apertaTecla(m, 'j')
	if m.armadoFila != "" {
		t.Fatal("outra tecla desarma o y")
	}
	m.estado.Recs = nil
	m, _ = apertaTecla(m, 'y')
	m, cmd := apertaTecla(m, 'y')
	if cmd != nil || chamou || !strings.Contains(m.status, "nada para aprovar") {
		t.Fatalf("sem recomendação aprovável o y não responde: status=%q", m.status)
	}
	if p := telaFila(m); !strings.Contains(p, "o agente não recomendou") || !strings.Contains(p, "fora do vision") {
		t.Fatalf("sem recomendação, a fila diz isso:\n%s", p)
	}
}

func TestPaineisDoDiaNaGavetaMissoes(t *testing.T) {
	m := modeloFila(nil).ComGaveta("missoes")
	p := ansi.Strip(m.Render(140, 50))
	for _, s := range []string{"FEITO HOJE 4", "d21-placar       2 PRs · 9 commits", "mtg-preco-deck   1 PR · 4 commits",
		"sem missão       1 PR · 1 commit", "LIÇÕES 3", "4× lint: import não usado", "→ rodar o lint antes do commit"} {
		if !strings.Contains(p, s) {
			t.Errorf("sem %q:\n%s", s, p)
		}
	}
	e := EstadoDemo(agoraFila)
	e.Mergeados, e.Licoes = nil, nil
	if p := ansi.Strip(modelo(e, &registro{}).ComGaveta("missoes").Render(140, 50)); strings.Contains(p, "FEITO HOJE") || strings.Contains(p, "LIÇÕES") {
		t.Fatalf("sem dado, sem seção:\n%s", p)
	}
}
