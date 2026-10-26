package dados

import (
	"reflect"
	"testing"
)

func L(chave, rotulo, detalhe string, extra ...string) Linha {
	l := Linha{Chave: chave, Rotulo: rotulo, Detalhe: detalhe}
	if len(extra) > 0 {
		l.Extra = extra[0]
	}
	if len(extra) > 1 {
		l.Merge = extra[1]
	}
	return l
}

func PR(n int, motivo string) Travado {
	return Travado{Repo: "PrologWebService", Number: n, Motivo: motivo, Title: "x", URL: "https://gh/" + string(rune('0'+n))}
}

func tipos(f []ItemFeed) []TipoFeed {
	var t []TipoFeed
	for _, i := range f {
		t = append(t, i.Tipo)
	}
	return t
}

func count(f []ItemFeed, t TipoFeed) int {
	n := 0
	for _, i := range f {
		if i.Tipo == t {
			n++
		}
	}
	return n
}

func TestFeedOrdemUrgencia(t *testing.T) {
	e := Estado{
		Tasks: []Linha{
			L("/d/a", "c-pl-1", "done · 0/2 perguntas · 3h · done"),
			L("/d/b", "c-pl-2", "sem report · 1/1 perguntas · 5min · blocked", "w1:t9"),
		},
		Esperando: []Linha{L("/d/c/questions.json#plano", "c-pl-3 · alto", "Aprova o plano?", "coord: B")},
		Travados:  []Travado{PR(1, "aguardando revisão"), PR(2, "aprovado · mergear"), PR(3, "CI vermelho")},
	}
	want := []TipoFeed{TipoDecidir, TipoTravado, TipoLer}
	if g := tipos(MontarFeed(e)); !reflect.DeepEqual(g, want) {
		t.Errorf("got %v want %v", g, want)
	}
}

func TestGateJuntarNaoDuplica(t *testing.T) {
	m := `{"repo":"PrologWebService","pr":2,"url":"https://gh/2"}`
	e := Estado{Esperando: []Linha{L("/d/c/questions.json#juntar", "c-pl-3 · alto", "Libera?", "", m)}, Travados: []Travado{PR(2, "aprovado · mergear")}}
	f := MontarFeed(e)
	if count(f, TipoJuntar) != 1 || f[0].Pergunta == nil || f[0].Pergunta.Chave != "/d/c/questions.json#juntar" || f[0].Merge == nil {
		t.Errorf("%+v", f)
	}
}

func TestMax3Ler(t *testing.T) {
	var e Estado
	for n := 1; n <= 5; n++ {
		e.Tasks = append(e.Tasks, L("/d/"+string(rune('0'+n)), "c", "done · 0/2 perguntas · 3h · done"))
	}
	if n := count(MontarFeed(e), TipoLer); n != 3 {
		t.Error(n)
	}
}

func TestFeedVazio(t *testing.T) {
	if f := MontarFeed(Estado{}); len(f) != 0 {
		t.Error(f)
	}
}

func TestTaskTravadaComPerguntaSoDecidir(t *testing.T) {
	e := Estado{
		Tasks:     []Linha{L("/d/b", "c-pl-2", "sem report · 1/1 perguntas · 5min · blocked", "w1:t9")},
		Esperando: []Linha{L("/d/b/questions.json#plano", "c-pl-2 · alto", "Aprova?")},
	}
	if g := tipos(MontarFeed(e)); !reflect.DeepEqual(g, []TipoFeed{TipoDecidir}) {
		t.Error(g)
	}
	e.Esperando = []Linha{L("/d/b2/questions.json#x", "c-pl-9 · alto", "Aprova?")}
	if g := tipos(MontarFeed(e)); !reflect.DeepEqual(g, []TipoFeed{TipoDecidir, TipoTravado}) {
		t.Error(g)
	}
}

func TestTipoString(t *testing.T) {
	if TipoDecidir.String() != "decidir" || TipoLer.String() != "ler" {
		t.Error("String")
	}
}

func TestResumo(t *testing.T) {
	if g := Resumo(Estado{}); g != "tudo calmo por aqui" {
		t.Error(g)
	}
	e := Estado{
		Esperando: []Linha{L("/d/c/questions.json#p", "c-pl-3 · alto", "Aprova?")},
		Travados:  []Travado{PR(1, "conflito"), PR(2, "CI vermelho")},
		Tasks:     []Linha{L("/d/w", "c", "sem report · 0/0 · 1min · working")},
	}
	if g := Resumo(e); g != "1 pra você · 1 trabalhando" {
		t.Error(g)
	}
}
