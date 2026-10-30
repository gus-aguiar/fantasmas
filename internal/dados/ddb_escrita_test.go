package dados

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResponderPeloDdb(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	t.Setenv("FANTASMAS_DDB_SCRIPT", "/s/ddb.py")
	f := &respRunner{resp: []respResp{{"python3 /s/ddb.py", "a\tc-a\tPergunta?\tSim\tsim: c-a\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, "/raiz/.delegate/a/questions.json#p1", "-ajustar: x")
	if err != nil {
		t.Fatal(err)
	}
	if got := f.juntas(); len(got) != 1 || got[0] != "python3 /s/ddb.py --raiz /raiz answer --via control-panel -- a#p1 -ajustar: x" {
		t.Fatalf("argv: %v", got)
	}
	want := Entrega{Slug: "a", Crew: "c-a", Titulo: "Pergunta?", Resposta: "Sim", Entregue: true}
	if e != want {
		t.Fatalf("entrega: %+v", e)
	}
	nenhumPromptLocal(t, f, "c-a")
}

func TestResponderPeloDdbNaoEntregue(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := &respRunner{resp: []respResp{{"python3", "a\tc-a\tP\tSim\tnão: herdr falhou (x)\t-\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{}, "/raiz/.delegate/a/questions.json#p1", "Sim")
	if err != nil || e.Entregue || e.Motivo != "herdr falhou (x)" {
		t.Fatalf("%+v %v", e, err)
	}
	f = &respRunner{resp: []respResp{{"python3", "", errors.New("exit 3: já respondido")}}}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{}, "/raiz/.delegate/a/questions.json#p1", "Sim"); err == nil {
		t.Fatal("erro do ddb deveria subir")
	}
}

func TestSemChaveSegueOCaminhoAntigo(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "")
	f := &respRunner{resp: []respResp{{"python3 /p/painel.py", "pl-1\tc-pl-1\tP\tA\tsim: c-pl-1\n", nil}}}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py"}, "k", "A"); err != nil {
		t.Fatal(err)
	}
	if got := f.juntas(); len(got) != 1 || !strings.Contains(got[0], "painel.py responder") {
		t.Fatalf("argv: %v", got)
	}
}

func TestArquivarPeloDdb(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	t.Setenv("FANTASMAS_DDB_SCRIPT", "/s/ddb.py")
	f := &respRunner{}
	if err := Arquivador(f, "/raiz")("/raiz/.delegate/a/questions.json#p1"); err != nil {
		t.Fatal(err)
	}
	if got := f.juntas(); len(got) != 1 || got[0] != "python3 /s/ddb.py --raiz /raiz arquivar a#p1" {
		t.Fatalf("argv: %v", got)
	}
}

func TestArquivarSemChaveEscreveArquivo(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "")
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	f := &respRunner{}
	if err := Arquivador(f, d)(filepath.Join(d, "a", "questions.json") + "#p1"); err != nil {
		t.Fatal(err)
	}
	if _, e := os.Stat(filepath.Join(d, "a", "answers.json")); e != nil {
		t.Fatal(e)
	}
	if len(f.calls) != 0 {
		t.Fatalf("não deveria chamar o ddb: %v", f.juntas())
	}
}

func TestRetomarMontaComandos(t *testing.T) {
	f := &respRunner{}
	c := Caido{Nome: "c-pl-9", Pane: "w1:p9", SessionID: "abc-123"}
	if err := Retomar(context.Background(), f, c); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"herdr pane run w1:p9 claude --dangerously-skip-permissions --name c-pl-9 --resume abc-123",
		"herdr agent rename w1:p9 c-pl-9",
	}
	got := f.juntas()
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("argv: %v", got)
	}
	if len(f.calls[0]) != 5 || f.calls[0][4] != "claude --dangerously-skip-permissions --name c-pl-9 --resume abc-123" {
		t.Fatalf("o comando deve ir como um argumento só: %q", f.calls[0])
	}
}

func TestRetomarRecusaSemSessionOuValorInseguro(t *testing.T) {
	f := &respRunner{}
	if err := Retomar(context.Background(), f, Caido{Nome: "c", Pane: "p"}); err == nil {
		t.Fatal("sem session_id deveria falhar")
	}
	if err := Retomar(context.Background(), f, Caido{Nome: "c; rm -rf /", Pane: "p", SessionID: "s"}); err == nil {
		t.Fatal("nome inseguro deveria falhar")
	}
	if err := Retomar(context.Background(), f, Caido{Nome: "c", Pane: "p", SessionID: "-x"}); err == nil {
		t.Fatal("session_id começando com - deveria falhar")
	}
	if len(f.calls) != 0 {
		t.Fatalf("nada deveria rodar: %v", f.juntas())
	}
}

func TestCaidosDe(t *testing.T) {
	raiz := t.TempDir()
	meta := func(slug, corpo string) {
		d := filepath.Join(raiz, ".delegate", slug)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "meta"), []byte(corpo), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	meta("a", "name=c-a\npane=w1:p1\nsession_id=s-a\n")
	meta("b", "name=c-b\npane=w1:p2\nsession_id=s-b\n")
	meta("c", "name=c-c\npane=w1:p3\n")
	got := CaidosDe(raiz, map[string]bool{"c-b": true}, 8)
	if len(got) != 1 || got[0].Nome != "c-a" || got[0].SessionID != "s-a" || got[0].Pane != "w1:p1" {
		t.Fatalf("%+v", got)
	}
	if CaidosDe(raiz, nil, 8) != nil {
		t.Fatal("sem leitura do herdr ninguém é caído")
	}
}

type envRespRunner struct {
	respRunner
	envs [][]string
}

func (f *envRespRunner) RunEnv(ctx context.Context, env []string, argv ...string) (string, error) {
	f.mu.Lock()
	f.envs = append(f.envs, env)
	f.mu.Unlock()
	return f.Run(ctx, argv...)
}

func TestResponderComPainelPassaDdbEscritaNoEnv(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := &envRespRunner{respRunner: respRunner{resp: []respResp{{"python3 /p/painel.py", "a\tc-a\tP\tSim\tsim: c-a\t-\n", nil}}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py", Herdr: true}, "/raiz/.delegate/a/questions.json#p1", "Sim")
	if err != nil || !e.Entregue {
		t.Fatalf("%+v %v", e, err)
	}
	if got := f.juntas(); len(got) != 1 || got[0] != "python3 /p/painel.py responder /raiz /raiz/.delegate/a/questions.json#p1 Sim" {
		t.Fatalf("argv: %v", got)
	}
	if len(f.envs) != 1 || len(f.envs[0]) != 1 || f.envs[0][0] != "DDB_ESCRITA=on" {
		t.Fatalf("env: %v", f.envs)
	}
}

func TestResponderComPainelSemChaveNaoMexeNoEnv(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "")
	f := &envRespRunner{respRunner: respRunner{resp: []respResp{{"python3", "a\tc-a\tP\tSim\tsim: c-a\n", nil}}}}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py"}, "k", "Sim"); err != nil {
		t.Fatal(err)
	}
	if len(f.envs) != 0 {
		t.Fatalf("env inesperado: %v", f.envs)
	}
}

func TestEnvDeFallbackPrefixaEnv(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := &respRunner{resp: []respResp{{"env DDB_ESCRITA=on python3 /p/painel.py", "a\tc-a\tP\tSim\tsim: c-a\n", nil}}}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py"}, "k", "Sim"); err != nil {
		t.Fatal(err)
	}
	if got := f.juntas(); len(got) != 1 || !strings.HasPrefix(got[0], "env DDB_ESCRITA=on python3 /p/painel.py responder") {
		t.Fatalf("argv: %v", got)
	}
}
