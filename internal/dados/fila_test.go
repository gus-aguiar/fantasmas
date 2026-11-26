package dados

import (
	"errors"
	"testing"
)

const questionsFila = `{"crew":"c-1","items":[
 {"id":"q1","title":"Aprovar o plano?","recommendation":{"option":"sim"},"vision":"O merge é sempre humano",
  "options":[{"id":"sim","label":"Aprovar"},{"id":"nao","label":"Pedir ajuste","free":true}]},
 {"id":"q2","title":"Qual lib?","recommendation":"a mais leve",
  "options":[{"id":"a","label":"A"},{"id":"b","label":"B"}]},
 {"id":"q3","title":"Texto?","recommendation":{"option":"nao"},
  "options":[{"id":"sim","label":"Aprovar"},{"id":"nao","label":"Pedir ajuste","free":true}]}
]}`

func TestLerRecomendacoes(t *testing.T) {
	lidos := 0
	ler := func(p string) ([]byte, error) {
		lidos++
		if p == "/ws/.delegate/t1/questions.json" {
			return []byte(questionsFila), nil
		}
		return nil, errors.New("sem arquivo")
	}
	esp := []Linha{
		{Chave: ".delegate/t1/questions.json#q1"}, {Chave: ".delegate/t1/questions.json#q2"},
		{Chave: ".delegate/t1/questions.json#q3"}, {Chave: ".delegate/t2/questions.json#q1"}, {Chave: "sem-id"},
	}
	r := LerRecomendacoes("/ws", esp, ler)
	if lidos != 2 {
		t.Fatalf("o mesmo arquivo devia ser lido uma vez só: %d leituras", lidos)
	}
	q1 := r[".delegate/t1/questions.json#q1"]
	if q1.Opcao != "sim" || q1.Texto != "Aprovar" || q1.Regra != "O merge é sempre humano" || !q1.Aprovavel() {
		t.Fatalf("q1=%+v", q1)
	}
	if q2 := r[".delegate/t1/questions.json#q2"]; q2.Aprovavel() || q2.Texto != "a mais leve" || q2.Regra != "" {
		t.Fatalf("recomendação só em texto não é aprovável: %+v", q2)
	}
	if q3 := r[".delegate/t1/questions.json#q3"]; q3.Aprovavel() || q3.Texto != "Pedir ajuste" {
		t.Fatalf("opção de texto livre não é aprovável com uma tecla: %+v", q3)
	}
	if _, ok := r[".delegate/t2/questions.json#q1"]; ok {
		t.Fatal("arquivo ilegível não devia gerar recomendação")
	}
}
