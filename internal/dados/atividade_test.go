package dados

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fixture(t *testing.T, nome string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", nome))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAtividadeBash(t *testing.T) {
	a := AtividadeDe(fixture(t, "transcript-bash.jsonl"))
	if a.Acao != "Espera o veredito do Spark no PR do React" {
		t.Errorf("Acao = %q", a.Acao)
	}
	quando, _ := time.Parse(time.RFC3339Nano, "2026-10-26T19:15:16.768Z")
	if !a.Quando.Equal(quando) {
		t.Errorf("Quando = %v", a.Quando)
	}
	if a.Tokens != 2+545174+806 {
		t.Errorf("Tokens = %d", a.Tokens)
	}
	if a.Modelo != "claude-opus-5-5" {
		t.Errorf("Modelo = %q", a.Modelo)
	}
}

func linhaTool(nome, input string) []byte {
	return []byte(`{"type":"assistant","timestamp":"2026-10-26T19:00:00Z","message":{"model":"m","usage":{"input_tokens":1},"content":[{"type":"tool_use","name":"` + nome + `","input":` + input + `}]}}` + "\n")
}

func TestAtividadeFallbacks(t *testing.T) {
	casos := []struct{ nome, input, esperado string }{
		{"Edit", `{"file_path":"/a/b/gavetas.go"}`, "edita gavetas.go"},
		{"Read", `{"file_path":"/x/y/leia.md"}`, "lê leia.md"},
		{"Skill", `{"skill":"writing-plans"}`, "skill writing-plans"},
		{"Agent", `{"description":"Map X"}`, "agente: Map X"},
		{"SendMessage", `{}`, "SendMessage"},
		{"Agent", `{}`, "Agent"},
		{"Edit", `{}`, "Edit"},
	}
	for _, c := range casos {
		if a := AtividadeDe(linhaTool(c.nome, c.input)); a.Acao != c.esperado {
			t.Errorf("%s: Acao = %q, esperado %q", c.nome, a.Acao, c.esperado)
		}
	}
}

func TestAtividadeSoTexto(t *testing.T) {
	a := AtividadeDe(fixture(t, "transcript-so-texto.jsonl"))
	if a.Acao != "" {
		t.Errorf("Acao = %q, esperado vazio", a.Acao)
	}
	if a.Tokens != 2+738750+435 {
		t.Errorf("Tokens = %d", a.Tokens)
	}
}

func TestAtividadeLinhaGigante(t *testing.T) {
	gigante := append([]byte(`{"type":"user","x":"`), bytes.Repeat([]byte("a"), 2<<20)...)
	gigante = append(gigante, []byte("\"}\n")...)
	cauda := append(gigante, linhaTool("Bash", `{"description":"depois da linha gigante"}`)...)
	if a := AtividadeDe(cauda); a.Acao != "depois da linha gigante" {
		t.Errorf("Acao = %q", a.Acao)
	}
}

func TestAtividadeLixo(t *testing.T) {
	if a := AtividadeDe([]byte("{lixo\n\n")); a != (Atividade{}) {
		t.Errorf("esperado Atividade{}, veio %+v", a)
	}
}

func TestLerCaudaLinhaParcial(t *testing.T) {
	caminho := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(caminho, []byte("aaaaaaaaaa\nbbbb\ncccc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := string(LerCauda(caminho, 10)); got != "bbbb\ncccc\n" {
		t.Errorf("corte logo após quebra de linha descartou linha íntegra: %q", got)
	}
	got := string(LerCauda(caminho, 14))
	if got != "bbbb\ncccc\n" {
		t.Errorf("cauda = %q", got)
	}
	if string(LerCauda(caminho, 1000)) != "aaaaaaaaaa\nbbbb\ncccc\n" {
		t.Errorf("arquivo menor que max deve vir inteiro")
	}
	if LerCauda(filepath.Join(t.TempDir(), "nao-existe"), 10) != nil {
		t.Error("arquivo ausente deve dar nil")
	}
}

func TestCtxPct(t *testing.T) {
	t.Setenv("FANTASMAS_CTX_JANELA", "")
	for _, c := range []struct{ tokens, esperado int }{{547242, 54}, {0, -1}, {2000000, 100}} {
		if got := CtxPct(Atividade{Tokens: c.tokens}); got != c.esperado {
			t.Errorf("CtxPct(%d) = %d, esperado %d", c.tokens, got, c.esperado)
		}
	}
	t.Setenv("FANTASMAS_CTX_JANELA", "200000")
	if got := CtxPct(Atividade{Tokens: 100000}); got != 50 {
		t.Errorf("com janela custom = %d", got)
	}
}

func TestFormatoConhecido(t *testing.T) {
	for _, nome := range []string{"transcript-bash.jsonl", "transcript-misto.jsonl", "transcript-so-texto.jsonl"} {
		a := AtividadeDe(fixture(t, nome))
		if a.Acao == "" && a.Tokens == 0 {
			t.Errorf("%s: parser zerou tudo, o formato do Claude Code mudou?", nome)
		}
	}
}
