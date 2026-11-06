package ui

import (
	"errors"
	"strings"
	"testing"

	"fantasmas/internal/dados"
)

func comCaido(sid string) (Model, *[]dados.Caido) {
	e := fixture()
	e.Caidos = []dados.Caido{{Nome: "c-pl-9", Pane: "w1:p9", SessionID: sid, Pasta: "/x/.delegate/s9"}}
	var chamadas []dados.Caido
	m := modelo(e, &registro{})
	m.acoes.Retomar = func(c dados.Caido) error { chamadas = append(chamadas, c); return nil }
	m.gaveta = indiceGaveta("agora")
	m.largura, m.altura = 96, 40
	for m.sel = 0; m.sel < len(m.alvos()); m.sel++ {
		if m.alvos()[m.sel].retomar != "" {
			break
		}
	}
	return m, &chamadas
}

func TestGavetaCrewMostraCaidos(t *testing.T) {
	m, _ := comCaido("sid-1")
	out := render(m, 96, 40)
	if !strings.Contains(out, "caído") || !strings.Contains(out, "R 2× retoma") || !strings.Contains(out, "c-pl-9") {
		t.Fatalf("raia dos caídos ausente:\n%s", out)
	}
}

func TestRetomarPedeConfirmacao(t *testing.T) {
	m, chamadas := comCaido("sid-1")
	m = aperta(t, m, "R")
	if len(*chamadas) != 0 || !strings.Contains(m.status, "R de novo") {
		t.Fatalf("primeiro R deveria só armar: %v %q", *chamadas, m.status)
	}
	m = aperta(t, m, "o")
	m = aperta(t, m, "R")
	if len(*chamadas) != 0 {
		t.Fatal("outra tecla deveria desarmar")
	}
	m = aperta(t, m, "R")
	if len(*chamadas) != 1 || (*chamadas)[0].SessionID != "sid-1" {
		t.Fatalf("segundo R deveria retomar: %v", *chamadas)
	}
}

func TestRetomarSemSessionId(t *testing.T) {
	m, chamadas := comCaido("")
	m = aperta(t, m, "R")
	m = aperta(t, m, "R")
	if len(*chamadas) != 0 || !strings.Contains(m.toast, "sem session_id; retome à mão") {
		t.Fatalf("chamadas=%v toast=%q", *chamadas, m.toast)
	}
}

func TestRetomarErroViraToast(t *testing.T) {
	m, _ := comCaido("sid-1")
	m.acoes.Retomar = func(dados.Caido) error { return errors.New("herdr caiu") }
	m = aperta(t, m, "R")
	m = aperta(t, m, "R")
	if !strings.Contains(m.toast, "não retomei: herdr caiu") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestDicaSoNaLinhaSelecionada(t *testing.T) {
	m, _ := comCaido("sid-1")
	if out := render(m, 96, 40); strings.Contains(out, "retoma caído") {
		t.Fatalf("dica não devia ir no título da seção:\n%s", out)
	}
	m.sel = 0
	if m.alvos()[0].retomar != "" {
		t.Skip("caído é o primeiro item")
	}
	if out := render(m, 96, 40); strings.Contains(out, "R 2× retoma") {
		t.Fatalf("dica do caído fora da linha selecionada:\n%s", out)
	}
}

func TestDicaPorTipo(t *testing.T) {
	casos := map[string]alvo{
		"enter responder · g ir à aba · o abrir": {pergunta: "q", aba: "a", url: "u"},
		"g ir à aba":             {aba: "a"},
		"R 2× retoma":            {retomar: "c"},
		"enter detalhes · g aba": {missao: "m"},
		"o abrir":                {url: "u", dica: "o abrir"},
		"enter expande":          {antigas: true},
	}
	for want, a := range casos {
		if got := dicaItem(a); got != want {
			t.Errorf("dicaItem(%+v) = %q, quero %q", a, got, want)
		}
	}
}
