package dados

import (
	"encoding/json"
	"reflect"
	"testing"
)

const qJSON = `{
 "crew": "c-pl-17724", "key": "PL-17724", "area": "Edição ampliada",
 "items": [
  {"id": "plano", "kind": "gate", "title": "Aprova o plano?", "peso": "alto",
   "context": ["Contexto.", "- item"], "explain": ["Por quê."],
   "options": [
     {"id": "A", "label": "Aprovar", "effect": "Segue a execução", "default": true},
     {"id": "B", "label": "Segurar", "effect": "Espera"}],
   "recommendation": {"option": "A", "text": "ignorado"},
   "coordinator": "A, plano cobre os ACs", "unlocks": "Execução",
   "merge": {"repo": "PrologWebReact", "pr": 3901, "url": "https://gh/3901"},
   "links": [{"label": "Épico", "url": "https://jira/1"}, {"label": "sem url"}]},
  {"id": "env", "kind": "action", "title": "Copia o .env?", "commands": ["cp a b"]}
 ]}`

func TestPerguntaTrazConteudo(t *testing.T) {
	d, err := PerguntaDe("/d/x/questions.json#plano", []byte(qJSON), "plano")
	if err != nil || d == nil {
		t.Fatalf("d=%v err=%v", d, err)
	}
	if d.Crew != "c-pl-17724" || d.Recomendacao != "A" {
		t.Errorf("crew=%q rec=%q", d.Crew, d.Recomendacao)
	}
	var ops [][2]any
	for _, o := range d.Opcoes {
		ops = append(ops, [2]any{o.ID, o.Padrao})
	}
	if !reflect.DeepEqual(ops, [][2]any{{"A", true}, {"B", false}}) {
		t.Errorf("opcoes=%v", ops)
	}
	if !reflect.DeepEqual(d.Contexto, []string{"Contexto.", "- item"}) {
		t.Errorf("contexto=%v", d.Contexto)
	}
	var urls []string
	for _, l := range d.Links {
		urls = append(urls, l.URL)
	}
	if !reflect.DeepEqual(urls, []string{"https://jira/1", "https://gh/3901"}) {
		t.Errorf("links=%v", urls)
	}
	if d.Links[1].Label != "PR #3901 PrologWebReact" {
		t.Errorf("label=%q", d.Links[1].Label)
	}
}

func TestAcaoSemOpcoes(t *testing.T) {
	d, _ := PerguntaDe("k#env", []byte(qJSON), "env")
	if len(d.Opcoes) != 2 || d.Opcoes[0].Label != "Feito" || d.Opcoes[1].Label != "Não vou fazer" || !d.Opcoes[1].Livre {
		t.Errorf("opcoes=%+v", d.Opcoes)
	}
	if !reflect.DeepEqual(d.Comandos, []string{"cp a b"}) {
		t.Errorf("comandos=%v", d.Comandos)
	}
}

func TestIDInexistenteNilNil(t *testing.T) {
	d, err := PerguntaDe("k#nada", []byte(qJSON), "nada")
	if d != nil || err != nil {
		t.Errorf("d=%v err=%v", d, err)
	}
	if _, err := PerguntaDe("k", []byte("{ruim"), "x"); err == nil {
		t.Error("esperava erro de JSON")
	}
}

func TestRelatorio(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"crew": "c", "status": "done", "summary": []string{"feito"},
		"links": []map[string]string{{"label": "PR", "url": "u"}}, "report_path": ".delegate/x/report.md"})
	d, err := RelatorioDe("/d/x", b)
	if err != nil {
		t.Fatal(err)
	}
	if d.Tipo != "relatorio" || !reflect.DeepEqual(d.Resumo, []string{"feito"}) || d.Relatorio != ".delegate/x/report.md" || d.Titulo != "Relatório done" {
		t.Errorf("d=%+v", d)
	}
}

func TestEhLivre(t *testing.T) {
	for _, l := range []string{"Ajustar", "Ajustar o plano antes", "Segurar", "Outro nome", "Não aprovo", "Quero mudar alguma já", "Aprovo com ajustes", "Ajustar (escreva o que muda)", "Refazer"} {
		if !EhLivre(l, false) {
			t.Errorf("%q deveria ser livre", l)
		}
	}
	for _, l := range []string{"Aprovar", "Recusar destino inativo", "Mudar o Glorp", "Refazer neste PR", "React · endereço (em revisão)", "Sempre que o código muda"} {
		if EhLivre(l, false) {
			t.Errorf("%q não deveria ser livre", l)
		}
	}
	if !EhLivre("Aprovar", true) {
		t.Error("free:true deveria valer")
	}
}

func TestDocumentosPlanoPrevia(t *testing.T) {
	item := map[string]any{
		"context": []any{"Spec: docs/superpowers/tasks/x/_overview.md.", "Plano: docs/superpowers/tasks/x/_plan.md"},
		"plan":    ".delegate/x/plano.md",
	}
	docs := DocumentosDe(item)
	want := []string{".delegate/x/plano.md", "docs/superpowers/tasks/x/_overview.md", "docs/superpowers/tasks/x/_plan.md"}
	if !reflect.DeepEqual(docs, want) {
		t.Errorf("docs=%v", docs)
	}
	if PlanoDe(docs) != "docs/superpowers/tasks/x/_plan.md" {
		t.Errorf("plano=%q", PlanoDe(docs))
	}
	if PlanoDe(nil) != "" {
		t.Error("sem docs deveria ser vazio")
	}
	if p := PreviaDe("# T\n\n```\nx\n```\n---\n- a\n", 5); !reflect.DeepEqual(p, []string{"# T", "- a"}) {
		t.Errorf("previa=%v", p)
	}
}

func TestSeparar(t *testing.T) {
	a, id := Separar("/d/x/questions.json#plano#2")
	if a != "/d/x/questions.json#plano" || id != "2" {
		t.Errorf("a=%q id=%q", a, id)
	}
	if a, id := Separar("/d/x"); a != "/d/x" || id != "" {
		t.Errorf("a=%q id=%q", a, id)
	}
}

const qRico = `{"crew": "c-x", "key": "PL-9", "area": "Área",
 "items": [{"id": "juntar", "kind": "gate", "title": "Junta?",
   "merge": {"repo": "PrologWebReact", "pr": 3589, "url": "https://gh/3589", "base": "tronco", "spark": "aprovado", "checks": "verdes"},
   "diagram": "flowchart LR\n  A --> B", "diagram_caption": "o fluxo",
   "compare_title": "Lado a lado", "compare_head": ["Ponto", "Pedido", "Hoje", ""],
   "compare": [{"item": "Título", "a": "x", "b": "y", "status": "diff", "tag": "difere"}]}]}`

func TestPerguntaRica(t *testing.T) {
	d, err := PerguntaDe("/d/x/questions.json#juntar", []byte(qRico), "juntar")
	if err != nil || d == nil {
		t.Fatalf("d=%v err=%v", d, err)
	}
	if d.Key != "PL-9" || d.Area != "Área" || d.Uid != "/d/x/questions.json#juntar" {
		t.Errorf("key=%q area=%q uid=%q", d.Key, d.Area, d.Uid)
	}
	if d.Merge == nil || d.Merge.PR != 3589 || d.Merge.Base != "tronco" || d.Merge.Checks != "verdes" {
		t.Errorf("merge=%+v", d.Merge)
	}
	if d.Diagrama != "flowchart LR\n  A --> B" || d.LegendaDiagrama != "o fluxo" {
		t.Errorf("diagrama=%q legenda=%q", d.Diagrama, d.LegendaDiagrama)
	}
	want := []LinhaCompara{{Item: "Título", A: "x", B: "y", Status: "diff", Tag: "difere"}}
	if d.ComparaTitulo != "Lado a lado" || !reflect.DeepEqual(d.ComparaCab, []string{"Ponto", "Pedido", "Hoje", ""}) || !reflect.DeepEqual(d.Compara, want) {
		t.Errorf("compara=%q %v %v", d.ComparaTitulo, d.ComparaCab, d.Compara)
	}
}

func TestRelatorioTemOpcoesEUid(t *testing.T) {
	d, _ := RelatorioDe("/d/x", []byte(`{"crew":"c-pl-5","status":"done"}`))
	if d.Uid != "c-pl-5#report" || d.Kind != "report" {
		t.Errorf("uid=%q kind=%q", d.Uid, d.Kind)
	}
	if len(d.Opcoes) != 2 || d.Opcoes[0].Label != "Lido" || d.Opcoes[1].Label != "Lido, e tenho pedido" || !d.Opcoes[1].Livre || d.Opcoes[0].Livre {
		t.Errorf("opcoes=%+v", d.Opcoes)
	}
}
