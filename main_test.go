package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

func TestValidarGaveta(t *testing.T) {
	if validarGaveta("") != nil || validarGaveta("crew") != nil || validarGaveta("ci") != nil {
		t.Fatal("gaveta válida recusada")
	}
	err := validarGaveta("fantasminhas")
	if err == nil || !strings.Contains(err.Error(), "crew") || !strings.Contains(err.Error(), "agora") {
		t.Fatalf("erro sem lista de válidas: %v", err)
	}
}

// instantaneo lista arquivos e diretórios com tamanho, modo e mtime.
func instantaneo(t *testing.T, raiz string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(raiz, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(raiz, p)
		out = append(out, fmt.Sprintf("%s|%d|%o|%d", rel, info.Size(), info.Mode(), info.ModTime().UnixNano()))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func TestUmaVezNaoEscreve(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "fantasmas-teste")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	raiz := t.TempDir()
	cache := t.TempDir()
	d := filepath.Join(raiz, ".delegate", "slug-x")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "meta"), []byte("slug=slug-x\nname=c-x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	q := `[{"id":"q1","title":"Pergunta","question":"Segue?","options":[{"label":"Sim"}]}]`
	if err := os.WriteFile(filepath.Join(d, "questions.json"), []byte(q), 0o644); err != nil {
		t.Fatal(err)
	}

	antesRaiz, antesCache := instantaneo(t, raiz), instantaneo(t, cache)
	cmd := exec.Command(bin, "--uma-vez", "--raiz", raiz)
	cmd.Env = append(os.Environ(), "FANTASMAS_CACHE_DIR="+cache)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("--uma-vez: %v\n%s", err, out)
	}
	if got := instantaneo(t, raiz); !reflect.DeepEqual(antesRaiz, got) {
		t.Errorf("raiz mudou:\nantes=%v\ndepois=%v", antesRaiz, got)
	}
	if got := instantaneo(t, cache); !reflect.DeepEqual(antesCache, got) {
		t.Errorf("cache mudou:\nantes=%v\ndepois=%v", antesCache, got)
	}
}

func TestGavetaInicial(t *testing.T) {
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "FANTASMAS_FONTE" {
				return v
			}
			return ""
		}
	}
	if g := gavetaInicial("", env("")); g != "" {
		t.Fatalf("sem a variável deve manter o padrão (agora), veio %q", g)
	}
	if g := gavetaInicial("", env("ddb")); g != "missoes" || !ui.GavetaValida(g) {
		t.Fatalf("com FANTASMAS_FONTE=ddb deve abrir em missoes, veio %q", g)
	}
	if g := gavetaInicial("pipes", env("ddb")); g != "pipes" {
		t.Fatalf("a flag --gaveta manda, veio %q", g)
	}
	if g := gavetaInicial("", env("arquivos")); g != "" {
		t.Fatalf("outra fonte não muda a gaveta, veio %q", g)
	}
}
