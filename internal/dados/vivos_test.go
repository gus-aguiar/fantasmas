package dados

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestVivosDe(t *testing.T) {
	v := VivosDe([]byte(`{"result":{"agents":[{"name":"c-a"},{"name":null},{"agent":"x"}]}}`))
	if !reflect.DeepEqual(v, map[string]bool{"c-a": true}) {
		t.Error(v)
	}
	if VivosDe([]byte("lixo")) != nil {
		t.Error("saída ilegível deveria dar nil")
	}
}

func TestFeedSoDeVivosEAntigasDosMortos(t *testing.T) {
	e := Estado{
		Vivos: map[string]bool{"c-vivo": true},
		Esperando: []Linha{
			L("/d/v/questions.json#1", "c-vivo · alto", "Aprova?"),
			L("/d/m/questions.json#2", "c-morto · alto", "Antiga?"),
		},
		Tasks: []Linha{
			L("/d/m", "c-morto", "sem report · 0/0 · 1h · blocked"),
			L("/d/v2", "c-vivo", "done · 0/0 · 1h · done"),
		},
		Travados: []Travado{PR(2, "aprovado · mergear"), PR(3, "CI vermelho")},
	}
	f := MontarFeed(e)
	if g := tipos(f); !reflect.DeepEqual(g, []TipoFeed{TipoDecidir, TipoLer}) {
		t.Errorf("feed=%v", g)
	}
	a := MontarAntigas(e)
	if len(a) != 1 || a[0].Sujeito != "c-morto" || a[0].Pergunta.Chave != "/d/m/questions.json#2" {
		t.Errorf("antigas=%+v", a)
	}
	if Resumo(e) != "1 precisa de você" {
		t.Error(Resumo(e))
	}
}

func TestSemLeituraDoHerdrNinguemMorre(t *testing.T) {
	e := Estado{Esperando: []Linha{L("/d/m/questions.json#2", "c-x · alto", "?")}}
	if len(MontarFeed(e)) != 1 || len(MontarAntigas(e)) != 0 {
		t.Error("Vivos nil deveria manter tudo no feed")
	}
}

func TestArquivarPreservaRespostas(t *testing.T) {
	d := t.TempDir()
	ans := filepath.Join(d, "answers.json")
	if err := os.WriteFile(ans, []byte(`{"1":{"answer":"sim","at":"x","via":"control-panel"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Arquivar(filepath.Join(d, "questions.json") + "#2"); err != nil {
		t.Fatal(err)
	}
	var a map[string]map[string]string
	b, _ := os.ReadFile(ans)
	if err := json.Unmarshal(b, &a); err != nil {
		t.Fatal(err)
	}
	if a["1"]["answer"] != "sim" || a["2"]["answer"] != "(arquivado)" || a["2"]["via"] != "arquivado" || a["2"]["at"] == "" {
		t.Errorf("%v", a)
	}
	ents, _ := os.ReadDir(d)
	if len(ents) != 1 {
		t.Errorf("sobrou temporário: %v", ents)
	}
}

func TestArquivarSemAnswersCria(t *testing.T) {
	d := t.TempDir()
	if err := Arquivar(filepath.Join(d, "questions.json") + "#x"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(d, "answers.json"))
	if !strings.Contains(string(b), `"(arquivado)"`) {
		t.Error(string(b))
	}
	if Arquivar("sem-id") == nil {
		t.Error("chave sem id deveria falhar")
	}
}

func TestVistoSomeEReapareceQuandoMuda(t *testing.T) {
	t.Setenv("FANTASMAS_CACHE_DIR", t.TempDir())
	pasta := t.TempDir()
	rep := filepath.Join(pasta, "report.json")
	if err := os.WriteFile(rep, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := Estado{Tasks: []Linha{L(pasta, "c", "done · 0/0 · 1h · done")}}
	if count(MontarFeed(e), TipoLer) != 1 {
		t.Fatal("deveria ter ler")
	}
	if _, err := MarcarVisto(pasta); err != nil {
		t.Fatal(err)
	}
	e.Vistos = CarregarVistos()
	if count(MontarFeed(e), TipoLer) != 0 {
		t.Error("visto deveria sumir")
	}
	novo := time.Now().Add(time.Minute)
	if err := os.Chtimes(rep, novo, novo); err != nil {
		t.Fatal(err)
	}
	if count(MontarFeed(e), TipoLer) != 1 {
		t.Error("relatório mudou, deveria voltar")
	}
}

func TestCacheQuebradoNaoQuebra(t *testing.T) {
	d := t.TempDir()
	t.Setenv("FANTASMAS_CACHE_DIR", d)
	if err := os.WriteFile(filepath.Join(d, "vistos.json"), []byte("{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v := CarregarVistos(); len(v) != 0 {
		t.Error(v)
	}
	t.Setenv("FANTASMAS_CACHE_DIR", filepath.Join(d, "vistos.json", "x"))
	pasta := t.TempDir()
	os.WriteFile(filepath.Join(pasta, "report.md"), []byte("x"), 0o644)
	if k, err := MarcarVisto(pasta); k == "" || err == nil {
		t.Errorf("k=%q err=%v", k, err)
	}
}

func TestPRsTrazemBranchERascunho(t *testing.T) {
	p := PRsDe([]byte(`{"data":{"search":{"nodes":[{"number":7,"url":"u","repository":{"name":"R"},"isDraft":true,"headRefName":"feat/x"}]}}}`))
	if len(p) != 1 || !p[0].Draft || p[0].Branch != "feat/x" {
		t.Errorf("%+v", p)
	}
	if !strings.Contains(QueryDe("gus"), "headRefName") {
		t.Error("query sem headRefName")
	}
}
