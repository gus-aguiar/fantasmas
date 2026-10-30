package dados

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const estadoDdb = `{"tasks":[{"chave":"/r/.delegate/a","slug":"a","nome":"c-a","tab":"t1","marca":"esperando","status":"done","pendentes":1,"total":3,"idade_min":5,"quando":"5min","herdr":"idle"},
{"chave":"/r/.delegate/b","slug":"b","nome":"c-b","tab":"","marca":"done","status":"done","pendentes":0,"total":1,"idade_min":120,"quando":"2h","herdr":""}],
"pendentes":[{"chave":"/r/.delegate/a/questions.json#p1","slug":"a","item_id":"p1","kind":"gate","rotulo":"c-a · alto","detalhe":"Pergunta?","extra":"sugere: A","merge":{"repo":"X","pr":7}}],
"opcoes":{"/r/.delegate/a/questions.json#p1":[{"id":"A","label":"Sim","marcas":"padrão"},{"id":"x","label":"Ajustar","marcas":"livre"}]},
"mural":[{"at":"2026-10-30T09:00:00-03:00","autor":"g1","texto":"oi"},{"at":"2026-10-30T09:30:00-03:00","autor":"g2","texto":"tchau","ate":"2026-10-30T09:40:00-03:00"}],
"coordinator":"gengar-1","gerado_em":"2026-10-30T10:00:00-0300"}`

func TestFonteDdbEquivaleTSV(t *testing.T) {
	agora := time.Date(2026, 10, 30, 13, 0, 0, 0, time.UTC)
	p, err := DeDdb([]byte(estadoDdb), agora)
	if err != nil {
		t.Fatal(err)
	}
	tasks := TSV("/r/.delegate/a\tc-a\tesperando · 1/3 perguntas · 5min · idle\tt1\n/r/.delegate/b\tc-b\tdone · 0/1 perguntas · 2h\t\n")
	esperando := TSV("/r/.delegate/a/questions.json#p1\tc-a · alto\tPergunta?\tsugere: A\t{\"repo\":\"X\",\"pr\":7}\n")
	opcoes := TSV("/r/.delegate/a/questions.json#p1\tA\tSim\tpadrão\n/r/.delegate/a/questions.json#p1\tx\tAjustar\tlivre\n")
	if !reflect.DeepEqual(p.Tasks, tasks) {
		t.Errorf("tasks:\n%#v\n%#v", p.Tasks, tasks)
	}
	if !reflect.DeepEqual(p.Esperando, esperando) {
		t.Errorf("esperando:\n%#v\n%#v", p.Esperando, esperando)
	}
	if !reflect.DeepEqual(p.Opcoes, opcoes) {
		t.Errorf("opcoes:\n%#v\n%#v", p.Opcoes, opcoes)
	}
	if len(p.Mural) != 1 || p.Mural[0].Autor != "g1" {
		t.Errorf("mural=%+v (o aviso com ate vencido deve sair)", p.Mural)
	}
}

func TestFonteDdbCortaEm14(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`{"tasks":[`)
	for i := 0; i < 20; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(`{"chave":"k","nome":"n","marca":"done","pendentes":0,"total":0,"quando":"0min"}`)
	}
	sb.WriteString(`],"pendentes":[],"opcoes":{},"mural":[]}`)
	p, err := DeDdb([]byte(sb.String()), time.Now())
	if err != nil || len(p.Tasks) != 14 {
		t.Errorf("err=%v len=%d", err, len(p.Tasks))
	}
}

func raizDdb(t *testing.T) string {
	raiz := raizFake(t)
	d := filepath.Join(raiz, "scripts", "delegate")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "ddb.py"), []byte("#"), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func TestFonteDdbUsaState(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := novoFake()
	f.saida[" state --json"] = estadoDdb
	e := (&Coletor{Raiz: raizDdb(t), R: f}).Coletar(context.Background())
	if f.vezes("python3 "+e.Raiz+"/scripts/delegate/ddb.py") != 2 {
		t.Errorf("calls=%v", f.calls)
	}
	if len(e.Tasks) != 2 || e.Tasks[0].Rotulo != "c-a" || len(e.Esperando) != 1 || len(e.Opcoes) != 2 {
		t.Errorf("estado=%+v", e)
	}
	for _, c := range f.calls {
		if strings.HasSuffix(c, " tasks") || strings.HasSuffix(c, " esperando") || strings.HasSuffix(c, " opcoes-todas") {
			t.Errorf("caminho antigo rodou: %s", c)
		}
	}
	if e.Recursos.FalhaDdb != "" {
		t.Errorf("falha=%q", e.Recursos.FalhaDdb)
	}
}

func TestFonteDdbFalhaCaiNoAntigo(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := novoFake()
	f.erro[" state --json"] = true
	e := (&Coletor{Raiz: raizDdb(t), R: f}).Coletar(context.Background())
	if len(e.Tasks) != 1 || e.Tasks[0].Chave != "t1" || len(e.Esperando) != 1 {
		t.Errorf("deveria cair no antigo: %+v", e.Tasks)
	}
	if e.Recursos.FalhaDdb == "" {
		t.Error("falha do ddb deveria ser registrada")
	}
	achou := false
	for _, l := range FaltaDe(e.Recursos) {
		if strings.HasPrefix(l, "ddb:") {
			achou = true
		}
	}
	if !achou {
		t.Errorf("falta=%v", FaltaDe(e.Recursos))
	}
}

func TestFonteDdbJSONQuebradoCaiNoAntigo(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := novoFake()
	f.saida[" state --json"] = "não é json"
	e := (&Coletor{Raiz: raizDdb(t), R: f}).Coletar(context.Background())
	if len(e.Tasks) != 1 || e.Recursos.FalhaDdb == "" {
		t.Errorf("tasks=%v falha=%q", e.Tasks, e.Recursos.FalhaDdb)
	}
}

func TestSemVariavelUsaAntigo(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "")
	f := novoFake()
	e := (&Coletor{Raiz: raizDdb(t), R: f}).Coletar(context.Background())
	if f.vezes("python3 "+e.Raiz+"/scripts/delegate/ddb.py") != 0 {
		t.Errorf("ddb chamado sem a variável: %v", f.calls)
	}
	if len(e.Tasks) != 1 || e.Tasks[0].Chave != "t1" || e.Recursos.FalhaDdb != "" {
		t.Errorf("estado=%+v", e.Tasks)
	}
}
