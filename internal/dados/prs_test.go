package dados

import (
	"reflect"
	"strings"
	"testing"
)

func TestMotivoDe(t *testing.T) {
	rollup := func(s string) No {
		var n No
		var c NoCommit
		c.Commit.StatusCheckRollup = &Rollup{State: s}
		n.Commits.Nodes = []NoCommit{c}
		return n
	}
	ci := rollup("FAILURE")
	casos := []struct {
		n    No
		quer string
	}{
		{No{Mergeable: "CONFLICTING"}, "conflito"},
		{No{ReviewDecision: "CHANGES_REQUESTED"}, "mudanças pedidas"},
		{ci, "CI vermelho"},
		{No{IsDraft: true}, "rascunho"},
		{No{ReviewDecision: "APPROVED", Mergeable: "MERGEABLE"}, "aprovado · mergear"},
		{No{}, "aguardando revisão"},
		{No{Mergeable: "CONFLICTING", IsDraft: true}, "conflito · rascunho"},
	}
	for _, c := range casos {
		if g := MotivoDe(c.n); g != c.quer {
			t.Errorf("got %q want %q", g, c.quer)
		}
	}
}

func TestTomDoPR(t *testing.T) {
	for m, q := range map[string]string{"aprovado · mergear": "ok", "conflito": "aviso", "CI vermelho · rascunho": "aviso", "mudanças pedidas": "aviso", "rascunho": "neutro", "aguardando revisão": "neutro"} {
		if g := TomDoPR(m); g != q {
			t.Errorf("%q: got %q want %q", m, g, q)
		}
	}
}

func TestPRsDe(t *testing.T) {
	js := `{"data":{"search":{"nodes":[
	 {"number":1,"title":"a","url":"u1","repository":{"name":"R"}},
	 {"number":2,"title":"b","url":"u2","mergeable":"CONFLICTING","repository":{"name":"R"}},
	 {"number":3,"title":"c","url":"u3","reviewDecision":"APPROVED","mergeable":"MERGEABLE","repository":{"name":"R"}},
	 {}]}}}`
	var ns []int
	for _, p := range PRsDe([]byte(js)) {
		ns = append(ns, p.Number)
	}
	if !reflect.DeepEqual(ns, []int{3, 2, 1}) {
		t.Errorf("ordem %v", ns)
	}
	if PRsDe([]byte("lixo")) != nil {
		t.Error("json inválido deve dar nil")
	}
}

func TestMergeDe(t *testing.T) {
	if MergeDe(Linha{}) != nil || MergeDe(Linha{Merge: "{x"}) != nil {
		t.Error("esperava nil")
	}
	m := MergeDe(Linha{Merge: `{"repo":"R","pr":7,"url":"u"}`})
	if m == nil || m.PR != 7 || m.Repo != "R" {
		t.Errorf("got %+v", m)
	}
}

func TestQueryDe(t *testing.T) {
	if q := QueryDe("fulano"); !strings.Contains(q, "author:fulano org:prologapp") {
		t.Error(q)
	}
}
