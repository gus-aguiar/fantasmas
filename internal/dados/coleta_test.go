package dados

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type fakeRunner struct {
	mu    sync.Mutex
	saida map[string]string
	erro  map[string]bool
	calls []string
}

func (f *fakeRunner) Run(ctx context.Context, argv ...string) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		return "", errors.New("sem timeout no ctx")
	}
	k := strings.Join(argv, " ")
	f.mu.Lock()
	f.calls = append(f.calls, k)
	f.mu.Unlock()
	for p, e := range f.erro {
		if strings.HasPrefix(k, p) || strings.HasSuffix(k, p) {
			if e {
				return "partial", errors.New("exit 1")
			}
		}
	}
	for p, o := range f.saida {
		if strings.HasPrefix(k, p) || strings.HasSuffix(k, p) {
			return o, nil
		}
	}
	return "", nil
}

func (f *fakeRunner) vezes(prefixo string) int {
	n := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefixo) {
			n++
		}
	}
	return n
}

const graphql = `{"data":{"search":{"nodes":[{"number":7,"title":"[PL-1] x","url":"u","repository":{"name":"PrologWebReact"},"isDraft":true}]}}}`

func raizFake(t *testing.T) string {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, ".delegate"), 0o755); err != nil {
		t.Fatal(err)
	}
	painel := filepath.Join(raiz, "shared", "skills", "delegate-painel")
	if err := os.MkdirAll(painel, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(painel, "painel.py"), []byte("#"), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func novoFake() *fakeRunner {
	return &fakeRunner{
		erro: map[string]bool{},
		saida: map[string]string{
			"gh api user":         "gus\n",
			"gh api graphql":      graphql,
			" tasks":              "t1\tcrew1\tworking\n",
			" esperando":          "q1\tPergunta\tdet\n",
			" opcoes-todas":       "o1\tOpção\tefeito\n",
			"local.sh":            "tronco React\tno ar\t:3000\n",
			"--user gus":          `[{"databaseId":9,"displayTitle":"t","status":"completed","conclusion":"failure","workflowName":"build","createdAt":"2026-10-19T10:00:00Z","headBranch":"feat/x"}]`,
			"herdr agent list":    `{"result":{"agents":[{"name":"crew1"}]}}`,
			"coleta.sh worktrees": "w1\twt\tlimpo\n",
			"maquina.sh":          "m1\tRAM\t50%\n",
		},
	}
}

func TestColetaPreencheTudo(t *testing.T) {
	raiz := raizFake(t)
	f := novoFake()
	e := (&Coletor{Raiz: raiz, R: f}).Coletar(context.Background())
	if e.At.IsZero() || e.Raiz != raiz {
		t.Errorf("At=%v Raiz=%q", e.At, e.Raiz)
	}
	checks := map[string]int{
		"Tasks": len(e.Tasks), "Esperando": len(e.Esperando), "Opcoes": len(e.Opcoes),
		"Worktrees": len(e.Worktrees), "Maquina": len(e.Maquina), "Servicos": len(e.Servicos), "Travados": len(e.Travados),
	}
	for k, n := range checks {
		if n != 1 {
			t.Errorf("%s: len=%d", k, n)
		}
	}
	if len(e.CI) != len(reposCI) || e.CI[0].Detalhe != "completed/failure · t" {
		t.Errorf("CI=%v", e.CI)
	}
	if !e.Vivos["crew1"] {
		t.Errorf("vivos=%v", e.Vivos)
	}
	if e.Servicos[0].Chave != "tronco React" || e.Travados[0].Number != 7 {
		t.Errorf("servicos=%v travados=%v", e.Servicos, e.Travados)
	}
	r := e.Recursos
	if !r.GH || !r.Herdr || !r.Delegate || r.Autor != "gus" || !strings.HasSuffix(r.Painel, "painel.py") {
		t.Errorf("recursos=%+v", r)
	}
}

func TestColetorQueFalhaDeixaCampoVazio(t *testing.T) {
	f := novoFake()
	f.erro["gh run list"] = true
	e := (&Coletor{Raiz: raizFake(t), R: f}).Coletar(context.Background())
	if len(e.CI) != 0 {
		t.Errorf("CI=%v", e.CI)
	}
	if len(e.Tasks) != 1 || len(e.Maquina) != 1 || len(e.Travados) != 1 || e.At.IsZero() {
		t.Errorf("resto deveria estar cheio: %+v", e)
	}
}

func TestSemGHNaoChamaGraphql(t *testing.T) {
	f := novoFake()
	f.erro["gh --version"] = true
	e := (&Coletor{Raiz: raizFake(t), R: f}).Coletar(context.Background())
	if f.vezes("gh api graphql") != 0 || f.vezes("gh api user") != 0 {
		t.Errorf("calls=%v", f.calls)
	}
	if e.Recursos.GH || len(e.Travados) != 0 || len(e.Tasks) != 1 {
		t.Errorf("estado=%+v", e)
	}
}

func TestDescobertaUmaVez(t *testing.T) {
	f := novoFake()
	c := &Coletor{Raiz: raizFake(t), R: f}
	c.Coletar(context.Background())
	c.Coletar(context.Background())
	if n := f.vezes("gh api user"); n != 1 {
		t.Errorf("gh api user chamado %d vezes", n)
	}
	if n := f.vezes("gh --version"); n != 1 {
		t.Errorf("gh --version chamado %d vezes", n)
	}
}

func TestSemRaizDelegateNemPainel(t *testing.T) {
	e := (&Coletor{Raiz: t.TempDir(), R: novoFake()}).Coletar(context.Background())
	if e.Recursos.Delegate || e.Recursos.Painel != "" {
		t.Errorf("recursos=%+v", e.Recursos)
	}
}

func TestFaltaDe(t *testing.T) {
	if got := FaltaDe(Recursos{GH: true, Herdr: true, Delegate: true, Painel: "p", Autor: "a"}); len(got) != 0 {
		t.Errorf("got=%v", got)
	}
	tudo := FaltaDe(Recursos{})
	want := []string{
		"gh: gavetas PRs e CI vazias (instale e rode gh auth login)",
		`herdr: sem "ir →" para a aba do fantasminha`,
		".delegate/: nenhuma task do /delegate ainda",
	}
	if !reflect.DeepEqual(tudo, want) {
		t.Errorf("got=%v", tudo)
	}
	got := FaltaDe(Recursos{GH: true, Herdr: true, Delegate: true})
	if len(got) != 2 || !strings.HasPrefix(got[0], "login do GitHub") || got[1] != "painel.py: resposta só grava; a entrega fica com o coordenador" {
		t.Errorf("got=%v", got)
	}
}
