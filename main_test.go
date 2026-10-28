package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestUmaVezDemoCor(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t0 := time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)
	out := ui.Novo(nil, ui.Acoes{}, time.Minute).ComEstado(ui.EstadoDemo(t0)).Render(140, 42)
	if !strings.Contains(out, "\x1b[38;2;") || !strings.Contains(out, "c-pl-1") {
		t.Fatalf("quadro sem cor de 24 bits ou sem fantasminha da demo:\n%s", out)
	}
}

func TestComandoAbrir(t *testing.T) {
	casos := []struct {
		goos, editor, alvo string
		quer               []string
	}{
		{"linux", "", "https://x", []string{"xdg-open", "https://x"}},
		{"linux", "", "/home/a/.delegate/r.md", []string{"xdg-open", "/home/a/.delegate/r.md"}},
		{"darwin", "", "/a", []string{"open", "/a"}},
		{"windows", "", "/a", []string{"cmd", "/c", "start", "", "/a"}},
		{"linux", "code -n", "/a", []string{"code", "-n", "/a"}},
		{"linux", "code", "https://x", []string{"xdg-open", "https://x"}},
	}
	for _, c := range casos {
		if got := comandoAbrir(c.goos, c.editor, c.alvo); !reflect.DeepEqual(got, c.quer) {
			t.Fatalf("%+v: %v", c, got)
		}
	}
}

func TestEscolherModo(t *testing.T) {
	casos := []struct {
		env       map[string]string
		modo      string
		consultar bool
	}{
		{map[string]string{"TERM_PROGRAM": "WezTerm"}, "kitty", false},
		{map[string]string{"TERM_PROGRAM": "kitty"}, "kitty", false},
		{map[string]string{"HERDR_PANE_ID": "p4"}, "kitty", false},
		{map[string]string{"WEZTERM_EXECUTABLE": "/usr/bin/wezterm"}, "kitty", false},
		{map[string]string{"KITTY_WINDOW_ID": "1"}, "kitty", false},
		{map[string]string{"FANTASMAS_IMAGEM": "celulas", "HERDR_PANE_ID": "p4"}, "celulas", false},
		{map[string]string{"FANTASMAS_IMAGEM": "0", "TERM_PROGRAM": "WezTerm"}, "celulas", false},
		{map[string]string{"FANTASMAS_IMAGEM": "kitty"}, "kitty", false},
		{map[string]string{"FANTASMAS_IMAGEM": "1"}, "kitty", false},
		{map[string]string{"TERM_PROGRAM": "vscode"}, "", true},
		{map[string]string{}, "", true},
	}
	for _, c := range casos {
		modo, motivo, consultar := escolherModo(func(k string) string { return c.env[k] })
		if modo != c.modo || consultar != c.consultar || motivo == "" {
			t.Fatalf("%v: %q %q %v", c.env, modo, motivo, consultar)
		}
	}
}

func TestEscolherModoImagem(t *testing.T) {
	casos := []struct {
		env  map[string]string
		quer string
	}{
		{map[string]string{}, "animacao"},
		{map[string]string{"HERDR_PANE_ID": "wG:p4"}, "T"},
		{map[string]string{"HERDR_PANE_ID": "wG:p4", "FANTASMAS_IMAGEM_MODO": "animacao"}, "animacao"},
		{map[string]string{"HERDR_PANE_ID": "wG:p4", "FANTASMAS_IMAGEM_MODO": "quadros"}, "quadros"},
		{map[string]string{"FANTASMAS_IMAGEM_MODO": "T"}, "T"},
	}
	for _, c := range casos {
		if got := escolherModoImagem(func(k string) string { return c.env[k] }); got != c.quer {
			t.Fatalf("%v: %q, quer %q", c.env, got, c.quer)
		}
	}
}
