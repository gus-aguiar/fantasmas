package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestBarraCtx(t *testing.T) {
	casos := map[int]string{-1: "—", 0: "0%▫▫▫", 55: "55%▪▪▫", 100: "100%▪▪▪"}
	for pct, quer := range casos {
		if got := ansi.Strip(barraCtx(pct)); got != quer {
			t.Errorf("barraCtx(%d) = %q, quer %q", pct, got, quer)
		}
	}
}

func linhasEx() []dados.LinhaViva {
	h := time.Date(2026, 10, 30, 10, 12, 0, 0, time.Local)
	longo := "rodando a suíte de integração do módulo de pneus com Testcontainers 🚀 e conferindo a migração de índices — ação válida até o fim"
	return []dados.LinhaViva{
		{Nome: "c-pl-2", Aba: "a2", Etapa: "implementar", Agora: "1 sua(s): plano", Estado: dados.EsperaVoce, CtxPct: 72, Hora: h},
		{Nome: "c-pl-1-nome-bem-comprido", Aba: "a1", Etapa: "implementar", Agora: longo, Estado: dados.Trabalhando, CtxPct: 55, Hora: h},
		{Nome: "h-pl-3", Aba: "a3", Etapa: "homologar", Agora: "parado com tarefa há 14 min", Estado: dados.Parado, CtxPct: -1, Hora: h},
		{Nome: "c-pl-4", Aba: "a4", Etapa: "implementar", Agora: "entregou relatório", Estado: dados.Terminou, CtxPct: 100, Hora: h},
	}
}

func TestLinhaVivaCabe(t *testing.T) {
	for _, w := range larguras {
		for _, l := range linhasEx() {
			s := renderLinhaViva(l, w, T.Base)
			if lipgloss.Width(s) > w {
				t.Fatalf("w=%d largura %d: %q", w, lipgloss.Width(s), ansi.Strip(s))
			}
		}
	}
}

func textoAgora(e dados.Estado) string {
	c := &corpo{largura: 140}
	gavetaAgora(c, e, 42)
	var b strings.Builder
	for _, l := range c.linhas {
		b.WriteString(ansi.Strip(l.desenha(T.Base)) + "\n")
	}
	return b.String()
}

func TestRaiaVaziaSome(t *testing.T) {
	e := dados.Estado{At: time.Now()}
	for _, l := range linhasEx() {
		if l.Estado != dados.Parado {
			e.Linhas = append(e.Linhas, l)
		}
	}
	if s := textoAgora(e); strings.Contains(s, "parados") {
		t.Fatalf("raia vazia apareceu:\n%s", s)
	}
}

func TestRaiasOrdem(t *testing.T) {
	e := EstadoDemo(time.Date(2026, 10, 30, 10, 0, 0, 0, time.Local))
	e.Linhas = linhasEx()
	s := textoAgora(e)
	a, b, c := strings.Index(s, "precisa de você"), strings.Index(s, "› agora"), strings.Index(s, "parados")
	if a < 0 || b < 0 || c < 0 || !(a < b && b < c) {
		t.Fatalf("ordem errada (%d, %d, %d):\n%s", a, b, c, s)
	}
}

func TestHoraVivaEmHoraLocal(t *testing.T) {
	sp, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Skip(err)
	}
	antes := time.Local
	time.Local = sp
	defer func() { time.Local = antes }()
	l := dados.LinhaViva{Nome: "c-a", Agora: "x", Estado: dados.Trabalhando, CtxPct: -1, Hora: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)}
	got := ansi.Strip(renderLinhaViva(l, 100, T.Base))
	if !strings.HasSuffix(strings.TrimRight(got, " "), "06:00") {
		t.Errorf("hora não local: %q", got)
	}
}
