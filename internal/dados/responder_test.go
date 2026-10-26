package dados

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type respResp struct {
	prefixo string
	saida   string
	err     error
}

type respRunner struct {
	mu    sync.Mutex
	resp  []respResp
	calls [][]string
}

func (f *respRunner) Run(ctx context.Context, argv ...string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, argv)
	k := strings.Join(argv, " ")
	for _, r := range f.resp {
		if strings.HasPrefix(k, r.prefixo) {
			return r.saida, r.err
		}
	}
	return "", nil
}

func (f *respRunner) juntas() []string {
	var o []string
	for _, c := range f.calls {
		o = append(o, strings.Join(c, " "))
	}
	return o
}

// nenhumPromptLocal garante que todo argv com "prompt" é exatamente `herdr agent prompt <destino> ...`.
func nenhumPromptLocal(t *testing.T, f *respRunner, destino string) {
	t.Helper()
	for _, c := range f.calls {
		for _, a := range c {
			if a == "prompt" {
				if len(c) < 4 || c[0] != "herdr" || c[1] != "agent" || c[2] != "prompt" || c[3] != destino {
					t.Fatalf("prompt fora do destino esperado %q: %v", destino, c)
				}
			}
		}
	}
}

func TestRespondePeloPainel(t *testing.T) {
	f := &respRunner{resp: []respResp{{"python3 /p/painel.py", "pl-1\tc-pl-1\tPlano\tAprovar\tsim: c-pl-1\n", nil}}}
	rec := Recursos{Painel: "/p/painel.py", Herdr: true}
	e, err := Responder(context.Background(), f, "/raiz", rec, "pl-1:7", "Aprovar")
	if err != nil {
		t.Fatal(err)
	}
	if got := f.juntas(); len(got) != 1 || got[0] != "python3 /p/painel.py responder /raiz pl-1:7 Aprovar" {
		t.Fatalf("argv: %v", got)
	}
	want := Entrega{Slug: "pl-1", Crew: "c-pl-1", Titulo: "Plano", Resposta: "Aprovar", Entregue: true}
	if e != want {
		t.Fatalf("entrega: %+v", e)
	}
	nenhumPromptLocal(t, f, "c-pl-1")
}

func TestPainelNaoEntregou(t *testing.T) {
	f := &respRunner{resp: []respResp{{"python3", "pl-1\tc-pl-1\tPlano\tAprovar\tnão: herdr falhou (x)\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py", Herdr: true}, "k", "Aprovar")
	if err != nil {
		t.Fatal(err)
	}
	if e.Entregue || e.Motivo != "herdr falhou (x)" {
		t.Fatalf("entrega: %+v", e)
	}
	if len(f.calls) != 1 {
		t.Fatalf("esperava 1 chamada, vieram %v", f.juntas())
	}
	nenhumPromptLocal(t, f, "c-pl-1")
}

func TestPainelJaRespondido(t *testing.T) {
	f := &respRunner{resp: []respResp{{"python3", "", errors.New("python3: exit status 1: k não está pendente (já respondido ou não existe)")}}}
	_, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py", Herdr: true}, "k", "Aprovar")
	if err == nil || !strings.Contains(err.Error(), "não está pendente") {
		t.Fatalf("erro: %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("esperava 1 chamada, vieram %v", f.juntas())
	}
	nenhumPromptLocal(t, f, "c-pl-1")
}

func pastaPerguntas(t *testing.T, crew, kind, meta string) string {
	t.Helper()
	d := filepath.Join(t.TempDir(), "slug-x")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	q := `{"crew":"` + crew + `","items":[{"id":"plano","kind":"` + kind + `","title":"Plano"}]}`
	if err := os.WriteFile(filepath.Join(d, "questions.json"), []byte(q), 0o644); err != nil {
		t.Fatal(err)
	}
	if meta != "" {
		if err := os.WriteFile(filepath.Join(d, "meta"), []byte(meta), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(d, "questions.json")
}

const scriptDelegate = "python3 /raiz/shared/mods/control-panel/scripts/delegate.py responder "

func TestSemPainelGravaEEntregaAoCrew(t *testing.T) {
	arq := pastaPerguntas(t, "c-pl-1", "question", "")
	f := &respRunner{resp: []respResp{{"python3", "slug-x\tc-pl-1\tPlano\tAprovar\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, arq+"#plano", "Aprovar")
	if err != nil {
		t.Fatal(err)
	}
	got := f.juntas()
	if len(got) != 2 || got[0] != scriptDelegate+arq+" plano Aprovar" || got[1] != "herdr agent prompt c-pl-1 decision: plano: Aprovar" {
		t.Fatalf("argv: %v", got)
	}
	if !e.Entregue || e.Crew != "c-pl-1" || e.Slug != "slug-x" || e.Titulo != "Plano" {
		t.Fatalf("entrega: %+v", e)
	}
	nenhumPromptLocal(t, f, "c-pl-1")
}

func TestSemPainelAcaoFeitoVaiComoInfo(t *testing.T) {
	arq := pastaPerguntas(t, "c-pl-1", "action", "")
	f := &respRunner{resp: []respResp{{"python3", "slug-x\tc-pl-1\tPlano\tFeito\n", nil}}}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, arq+"#plano", "Feito"); err != nil {
		t.Fatal(err)
	}
	got := f.juntas()
	if len(got) != 2 || got[1] != "herdr agent prompt c-pl-1 info: plano: Feito" {
		t.Fatalf("argv: %v", got)
	}
}

func TestSemPainelCoordenadorVaiParaMeta(t *testing.T) {
	arq := pastaPerguntas(t, "coordenador", "question", "coordinator=prolog-claude-bd\n")
	f := &respRunner{resp: []respResp{{"python3", "slug-x\tcoordenador\tPlano\tAprovar\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, arq+"#plano", "Aprovar")
	if err != nil {
		t.Fatal(err)
	}
	got := f.juntas()
	if len(got) != 2 || got[1] != "herdr agent prompt prolog-claude-bd decision: plano: Aprovar" || !e.Entregue {
		t.Fatalf("argv: %v entrega: %+v", got, e)
	}
	nenhumPromptLocal(t, f, "prolog-claude-bd")
}

func TestSemPainelSemHerdrOuSemDestino(t *testing.T) {
	arq := pastaPerguntas(t, "c-pl-1", "question", "")
	f := &respRunner{resp: []respResp{{"python3", "slug-x\tc-pl-1\tPlano\tAprovar\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: false}, arq+"#plano", "Aprovar")
	if err != nil || e.Entregue || e.Motivo == "" || len(f.calls) != 1 {
		t.Fatalf("sem herdr: %+v err=%v calls=%v", e, err, f.juntas())
	}

	arq = pastaPerguntas(t, "coordenador", "question", "")
	f = &respRunner{resp: []respResp{{"python3", "slug-x\tcoordenador\tPlano\tAprovar\n", nil}}}
	e, err = Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, arq+"#plano", "Aprovar")
	if err != nil || e.Entregue || e.Motivo == "" || len(f.calls) != 1 {
		t.Fatalf("sem destino: %+v err=%v calls=%v", e, err, f.juntas())
	}
}

func TestRelatorioPeloPainelUsaUid(t *testing.T) {
	f := &respRunner{resp: []respResp{{"python3 /p/painel.py", "pl-5\tc-pl-5\tRelatório\tLido\tnão: relatório só lido, nada a entregar\n", nil}}}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Painel: "/p/painel.py"}, "c-pl-5#report", "Lido")
	if err != nil || e.Crew != "c-pl-5" || e.Entregue {
		t.Fatalf("e=%+v err=%v", e, err)
	}
	if got := f.juntas(); len(got) != 1 || got[0] != "python3 /p/painel.py responder /raiz c-pl-5#report Lido" {
		t.Errorf("calls=%v", got)
	}
}

func TestRelatorioSemPainel(t *testing.T) {
	f := &respRunner{}
	e, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, "c-pl-5#report", "Lido")
	if err != nil || e.Entregue || e.Crew != "c-pl-5" || e.Resposta != "Lido" || len(f.calls) != 0 {
		t.Errorf("lido: e=%+v err=%v calls=%v", e, err, f.calls)
	}
	if _, err := Responder(context.Background(), f, "/raiz", Recursos{Herdr: true}, "c-pl-5#report", "Lido, e tenho pedido: x"); err == nil || len(f.calls) != 0 {
		t.Errorf("pedido sem painel deveria falhar sem rodar nada: err=%v calls=%v", err, f.calls)
	}
}
