package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"
)

func modeloRelatorio(raiz string) Model {
	e := dados.Estado{At: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC), Raiz: raiz}
	m := Novo(func(context.Context) dados.Estado { return e }, Acoes{LerArquivo: os.ReadFile}, time.Second).ComEstado(e)
	m.largura, m.altura = 96, 40
	return m
}

func pastaTask(t *testing.T) (string, string) {
	raiz := t.TempDir()
	pasta := filepath.Join(raiz, ".delegate", "x")
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		t.Fatal(err)
	}
	return raiz, pasta
}

func TestAbrirRelatorioSoJSON(t *testing.T) {
	raiz, pasta := pastaTask(t)
	_ = os.WriteFile(filepath.Join(pasta, "report.json"), []byte(`{"crew":"c","status":"done","summary":["ok"]}`), 0o644)
	m := modeloRelatorio(raiz).abrirDetalhe(pasta, "relatorio")
	if m.aberto == nil || m.aberto.Titulo != "Relatório done" {
		t.Fatalf("aberto = %+v, status %q", m.aberto, m.status)
	}
}

func TestAbrirRelatorioSoMarkdown(t *testing.T) {
	raiz, pasta := pastaTask(t)
	md := "# Título\n\nprimeira\n\n```\ncodigo oculto\n```\nsegunda\n"
	_ = os.WriteFile(filepath.Join(pasta, "report.md"), []byte(md), 0o644)
	m := modeloRelatorio(raiz).abrirDetalhe(pasta, "relatorio")
	a := m.aberto
	if a == nil {
		t.Fatalf("não abriu: %q", m.status)
	}
	if a.Titulo != "Relatório" || a.Status != "done" || a.Tipo != "relatorio" || a.Relatorio != filepath.Join(pasta, "report.md") {
		t.Fatalf("aberto = %+v", a)
	}
	if got := strings.Join(a.Resumo, "|"); got != "# Título|primeira|segunda" {
		t.Fatalf("resumo = %q", got)
	}
}

func TestAbrirRelatorioNenhum(t *testing.T) {
	raiz, pasta := pastaTask(t)
	m := modeloRelatorio(raiz).abrirDetalhe(pasta, "relatorio")
	if m.aberto != nil || m.status != "não abri: a task não tem report.json nem report.md" {
		t.Fatalf("aberto %v, status %q", m.aberto, m.status)
	}
}

func TestErroDeArquivoMostraCaminhoRelativo(t *testing.T) {
	raiz := t.TempDir()
	m := modeloRelatorio(raiz).abrirDetalhe(filepath.Join(raiz, "q", "questions.json")+"#a", "pergunta")
	if strings.Contains(m.status, raiz) || !strings.Contains(m.status, filepath.Join("q", "questions.json")) {
		t.Fatalf("status = %q", m.status)
	}
}
