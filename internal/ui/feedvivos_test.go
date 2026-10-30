package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fantasmas/internal/dados"
)

func comMortos() dados.Estado {
	e := fixture()
	e.Vivos = map[string]bool{"c-pl-1": true, "c-pl-2": true, "c-pl-d": true}
	e.Esperando = append(e.Esperando, dados.Linha{Chave: "/x/morto/questions.json#q9", Rotulo: "c-pl-morto · alto", Detalhe: "pergunta velha?"})
	return e
}

func TestFeedSemPRsNemCI(t *testing.T) {
	m := modelo(fixture(), &registro{})
	out := render(m, 96, 40)
	for _, s := range []string{"WebService #", "Flutter", "CI vermelho"} {
		if strings.Contains(out, s) {
			t.Errorf("feed com %q:\n%s", s, out)
		}
	}
}

func TestAntigasRecolhidasEArquivarEmDoisPassos(t *testing.T) {
	var arquivadas []string
	m := modelo(comMortos(), &registro{})
	m.acoes.Arquivar = func(c string) error { arquivadas = append(arquivadas, c); return nil }
	m.largura, m.altura = 96, 40
	out := render(m, 96, 40)
	if !strings.Contains(out, "antigas (1)") || strings.Contains(out, "pergunta velha?") {
		t.Fatalf("antigas deveria vir recolhida:\n%s", out)
	}
	m = aperta(t, m, "a")
	if out = render(m, 96, 40); !strings.Contains(out, "pergunta velha?") {
		t.Fatalf("a não abriu:\n%s", out)
	}
	for m.sel = 0; m.sel < len(m.alvos()); m.sel++ {
		if m.alvos()[m.sel].arquivar != "" {
			break
		}
	}
	if !strings.Contains(render(m, 96, 40), BotaoArquivar) {
		t.Fatalf("sem botão arquivar:\n%s", render(m, 96, 40))
	}
	m = aperta(t, m, "x")
	if len(arquivadas) != 0 || !strings.Contains(render(m, 96, 40), BotaoConfirmaArq) {
		t.Fatal("primeiro x deveria só armar")
	}
	m = aperta(t, m, "o")
	m = aperta(t, m, "x")
	if len(arquivadas) != 0 {
		t.Fatal("outra tecla deveria desarmar")
	}
	m = aperta(t, m, "x")
	if len(arquivadas) != 1 || arquivadas[0] != "/x/morto/questions.json#q9" {
		t.Fatalf("arquivadas=%v", arquivadas)
	}
	if out = render(m, 96, 40); strings.Contains(out, "pergunta velha?") || strings.Contains(out, "antigas (") {
		t.Fatalf("item arquivado não sumiu:\n%s", out)
	}
}

func TestCliqueNoTituloEArquivar(t *testing.T) {
	var arquivadas []string
	m := modelo(comMortos(), &registro{})
	m.acoes.Arquivar = func(c string) error { arquivadas = append(arquivadas, c); return nil }
	m.largura, m.altura = 96, 40
	x, y := onde(t, m, "antigas (1)", 0)
	m, _ = clique(t, m, x, y)
	if !m.antigas {
		t.Fatal("clique no título deveria abrir")
	}
	x, y = onde(t, m, "pergunta velha?", 0)
	m.sel = -1
	for i, a := range m.alvos() {
		if a.arquivar != "" {
			m.sel = i
		}
	}
	x, y = onde(t, m, BotaoArquivar, 0)
	m, _ = clique(t, m, x, y)
	x, y = onde(t, m, BotaoConfirmaArq, 0)
	m, _ = clique(t, m, x, y)
	if len(arquivadas) != 1 {
		t.Fatalf("clique duplo deveria arquivar: %v", arquivadas)
	}
}

func TestLerSomeDepoisDeAberto(t *testing.T) {
	t.Setenv("FANTASMAS_CACHE_DIR", t.TempDir())
	pasta := t.TempDir()
	if err := os.WriteFile(filepath.Join(pasta, "report.json"), []byte(`{"status":"done","summary":["ok"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	e := fixture()
	e.Tasks = []dados.Linha{{Chave: pasta, Rotulo: "c-pl-rel", Detalhe: "done · 0/0 · 1h · done"}}
	m := modelo(e, &registro{})
	m.acoes.MarcarVisto = dados.MarcarVisto
	m.largura, m.altura = 96, 40
	if !strings.Contains(render(m, 96, 40), "c-pl-rel") {
		t.Fatal("sem ler")
	}
	m = m.abrirDetalhe(pasta, "relatorio")
	if m.aberto == nil {
		t.Fatalf("não abriu: %s", m.status)
	}
	m = m.fecharDetalhe()
	if strings.Contains(render(m, 96, 40), "c-pl-rel") {
		t.Fatal("ler deveria sumir depois de aberto")
	}
	m, _ = atualiza(t, m, coletouMsg{e})
	if strings.Contains(render(m, 96, 40), "c-pl-rel") {
		t.Fatal("coleta velha não deveria trazer de volta")
	}
	if v := dados.CarregarVistos(); len(v) != 1 {
		t.Fatalf("cache: %v", v)
	}
}

func TestRascunhosOcultos(t *testing.T) {
	e := fixture()
	e.Travados[1].Draft = true
	e.Travados[1].Title = "rascunho-x"
	m := modelo(e, &registro{})
	m.largura, m.altura = 96, 40
	m = aperta(t, m, "4")
	out := render(m, 96, 40)
	if strings.Contains(out, "rascunho-x") || !strings.Contains(out, "1 rascunhos ocultos · d mostra") {
		t.Fatalf("rascunho deveria estar oculto:\n%s", out)
	}
	m = aperta(t, m, "d")
	if out = render(m, 96, 40); !strings.Contains(out, "rascunho-x") || !strings.Contains(out, "d esconde") {
		t.Fatalf("d deveria mostrar:\n%s", out)
	}
}

func TestLargurasSemCorte(t *testing.T) {
	m := modelo(comMortos(), &registro{})
	m.antigas = true
	for i, a := range m.alvos() {
		if a.arquivar != "" {
			m.sel = i
		}
	}
	for _, w := range []int{46, 56, 96} {
		m.largura, m.altura = w, 40
		cabeTudo(t, m, w, 40)
		out := render(m, w, 40)
		for _, s := range []string{"antigas (1)", BotaoArquivar} {
			if !strings.Contains(out, s) {
				t.Errorf("%d: sem %q:\n%s", w, s, out)
			}
		}
	}
	e := fixture()
	e.Travados[0].Draft = true
	p := modelo(e, &registro{})
	p = aperta(t, p, "4")
	for _, w := range []int{46, 56, 96} {
		cabeTudo(t, p, w, 30)
		if !strings.Contains(render(p, w, 30), "rascunhos ocultos · d mostra") {
			t.Errorf("%d: aviso de rascunho cortado", w)
		}
	}
}

func TestAjudaTemAXD(t *testing.T) {
	m := modelo(fixture(), &registro{})
	m = aperta(t, m, "?")
	out := render(m, 96, 60)
	for _, s := range []string{"antigas", "arquiva", "rascunhos"} {
		if !strings.Contains(out, s) {
			t.Errorf("ajuda sem %q", s)
		}
	}
}
