package ui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const mergeJSON = `{"repo":"PrologWebReact","pr":77,"url":"https://github.com/o/PrologWebReact/pull/77","base":"dev","spark":"ok","checks":"ok"}`

func fixture() dados.Estado {
	return dados.Estado{
		At: time.Date(2026, 10, 16, 9, 5, 0, 0, time.UTC),
		Esperando: []dados.Linha{
			{Chave: "c-pl-1/questions.json#plano", Rotulo: "c-pl-1 · plano", Detalhe: "aprovar o plano?"},
			{Chave: "c-pl-2/questions.json#gate", Rotulo: "c-pl-2 · gate", Detalhe: "mergear?", Merge: mergeJSON},
		},
		Travados: []dados.Travado{
			{Repo: "PrologWebService", Number: 1, Motivo: "conflito", Title: "[PL-1] conserta", URL: "https://github.com/o/PrologWebService/pull/1"},
			{Repo: "PrologWebService", Number: 2, Motivo: "aguardando revisão", Title: "a", URL: "https://github.com/o/PrologWebService/pull/2"},
			{Repo: "PrologWebService", Number: 3, Motivo: "aguardando revisão", Title: "b", URL: "https://github.com/o/PrologWebService/pull/3"},
			{Repo: "PrologWebService", Number: 4, Motivo: "aguardando revisão", Title: "c", URL: "https://github.com/o/PrologWebService/pull/4"},
			{Repo: "PrologWebService", Number: 5, Motivo: "aguardando revisão", Title: "d", URL: "https://github.com/o/PrologWebService/pull/5"},
		},
		Tasks: []dados.Linha{
			{Chave: "t-a", Rotulo: "c-pl-a", Detalhe: "working", Extra: "aba-a"},
			{Chave: "t-b", Rotulo: "c-pl-b", Detalhe: "working", Extra: "aba-b"},
			{Chave: "t-c", Rotulo: "c-pl-c", Detalhe: "working · ação"},
			{Chave: "t-d", Rotulo: "c-pl-d", Detalhe: "done"},
			{Chave: "t-e", Rotulo: "c-pl-e", Detalhe: "done"},
		},
		CI: []dados.Linha{
			{Chave: "https://ci/1", Rotulo: "PrologFlutter · ci", Detalhe: "completed/failure · dev"},
			{Chave: "https://ci/2", Rotulo: "PrologWebReact · build", Detalhe: "completed/success · dev"},
		},
		Maquina:   []dados.Linha{{Chave: "trava", Rotulo: "pesados", Detalhe: "livre", Extra: "ok"}, {Chave: "ram", Rotulo: "RAM", Detalhe: "3 GB livres", Extra: "aviso"}},
		Servicos:  []dados.Linha{{Chave: "React", Rotulo: "no ar", Detalhe: "http://localhost:3000"}, {Chave: "WS", Rotulo: "parado", Detalhe: ":8081"}},
		Worktrees: []dados.Linha{{Chave: "/tmp/wt", Rotulo: "PrologWebReact · feature/x-PL-1", Detalhe: "/tmp/wt"}},
		Recursos:  dados.Recursos{GH: true, Herdr: true, Delegate: true, Painel: "/p.py", Autor: "eu"},
	}
}

type registro struct{ urls, abas []string }

func modelo(e dados.Estado, r *registro) Model {
	a := Acoes{
		AbrirURL: func(u string) error { r.urls = append(r.urls, u); return nil },
		FocarAba: func(s string) error { r.abas = append(r.abas, s); return nil },
	}
	m := Novo(func(context.Context) dados.Estado { return e }, a, time.Second)
	return m.ComEstado(e)
}

func tecla(m Model, k string) Model {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	n, _ := m.Update(msg)
	return n.(Model)
}

func TestTodasAsGavetasCabem(t *testing.T) {
	for _, g := range Gavetas {
		for _, w := range []int{30, 46, 56, 96, 140} {
			for _, h := range []int{8, 40} {
				m := modelo(fixture(), &registro{})
				m.gaveta = indiceGaveta(g.ID)
				out := m.Render(w, h)
				ls := strings.Split(out, "\n")
				if len(ls) != h {
					t.Fatalf("%s %dx%d: %d linhas", g.ID, w, h, len(ls))
				}
				for i, l := range ls {
					if lipgloss.Width(l) > w {
						t.Fatalf("%s %dx%d linha %d larga: %d", g.ID, w, h, i, lipgloss.Width(l))
					}
				}
				if ls[h-1] != Rodape(g.ID, w) {
					t.Fatalf("%s %dx%d: última linha não é o rodapé", g.ID, w, h)
				}
			}
		}
	}
}

func TestEsqueletoAntesDaColeta(t *testing.T) {
	m := Novo(func(context.Context) dados.Estado { return dados.Estado{} }, Acoes{}, time.Second)
	out := m.Render(96, 30)
	if !strings.Contains(out, "░") || strings.Contains(out, "precisa de você") {
		t.Fatalf("esqueleto ausente:\n%s", out)
	}
}

func TestTeclasTrocamGaveta(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m = tecla(m, "3")
	if Gavetas[m.gaveta].ID != "prs" {
		t.Fatalf("esperava prs, veio %s", Gavetas[m.gaveta].ID)
	}
	m = tecla(m, "7")
	m = tecla(m, "right")
	if Gavetas[m.gaveta].ID != "agora" {
		t.Fatalf("esperava agora, veio %s", Gavetas[m.gaveta].ID)
	}
}

func TestColetaNaoSobrepoe(t *testing.T) {
	var n int32
	m := Novo(func(context.Context) dados.Estado { atomic.AddInt32(&n, 1); return fixture() }, Acoes{}, 10*time.Millisecond)
	cmd := m.comecar()
	if cmd == nil {
		t.Fatal("Init sem coleta")
	}
	cmd()
	n1, c1 := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = n1.(Model)
	if c1 != nil {
		c1()
	}
	_, c2 := m.Update(tickColetaMsg{})
	if c2 != nil {
		c2()
	}
	if got := atomic.LoadInt32(&n); got != 1 {
		t.Fatalf("coletas = %d, esperava 1", got)
	}
}

func TestEnterEmPRAbreURL(t *testing.T) {
	r := &registro{}
	m := modelo(fixture(), r)
	m = tecla(m, "3")
	m = tecla(m, "enter")
	if len(r.urls) != 1 || r.urls[0] != "https://github.com/o/PrologWebService/pull/1" {
		t.Fatalf("urls = %v", r.urls)
	}
}

func TestFeedMostraMaisNaFila(t *testing.T) {
	e := fixture()
	for i := 0; i < 12; i++ {
		e.Esperando = append(e.Esperando, dados.Linha{Chave: fmt.Sprintf("q%d/questions.json#p", i), Rotulo: "c-pl-q · alto", Detalhe: "pergunta?"})
	}
	m := modelo(e, &registro{})
	if out := m.Render(96, 20); !strings.Contains(out, "na fila") {
		t.Fatalf("sem 'na fila':\n%s", out)
	}
}

func TestStatusDaResposta(t *testing.T) {
	m := modelo(fixture(), &registro{})
	n, _ := m.Update(respondeuMsg{Entrega: dados.Entrega{Crew: "c-pl-1", Resposta: "Aprovar", Entregue: true}})
	if s := n.(Model).status; s != "entregue a c-pl-1: Aprovar" {
		t.Fatalf("status = %q", s)
	}
	n, _ = m.Update(respondeuMsg{Entrega: dados.Entrega{Crew: "c-pl-1", Motivo: "herdr falhou"}})
	if s := n.(Model).status; s != "gravado; não entregue a c-pl-1: herdr falhou" {
		t.Fatalf("status = %q", s)
	}
}
