package dados

import (
	"reflect"
	"testing"
)

const graphqlCI = `{"data":{"search":{"nodes":[
 {"number":1,"title":"a","url":"u1","repository":{"name":"R"},"commits":{"nodes":[{"commit":{"statusCheckRollup":{"state":"FAILURE","contexts":{"nodes":[
   {"__typename":"CheckRun","name":"lint","conclusion":"FAILURE"},
   {"__typename":"CheckRun","name":"build","conclusion":"SUCCESS"},
   {"__typename":"CheckRun","name":"vitest","conclusion":"TIMED_OUT"}]}}}}]}},
 {"number":2,"title":"b","url":"u2","repository":{"name":"R"},"commits":{"nodes":[{"commit":{"statusCheckRollup":{"state":"PENDING","contexts":{"nodes":[]}}}}]}},
 {"number":3,"title":"c","url":"u3","repository":{"name":"R"},"commits":{"nodes":[{"commit":{"statusCheckRollup":{"state":"SUCCESS","contexts":{"nodes":[{"__typename":"StatusContext","context":"ci/x","state":"SUCCESS"}]}}}}]}},
 {"number":4,"title":"d","url":"u4","repository":{"name":"R"},"commits":{"nodes":[{"commit":{"statusCheckRollup":null}}]}}
]}}}`

func porNumero(ps []Travado) map[int]Travado {
	m := map[int]Travado{}
	for _, p := range ps {
		m[p.Number] = p
	}
	return m
}

func TestPRsDeCIEFalhos(t *testing.T) {
	m := porNumero(PRsDe([]byte(graphqlCI)))
	if m[1].CI != "FAILURE" || !reflect.DeepEqual(m[1].Falhos, []string{"lint", "vitest"}) {
		t.Errorf("falho: %+v", m[1])
	}
	if m[2].CI != "PENDING" || m[3].CI != "SUCCESS" || m[4].CI != "" {
		t.Errorf("CI: %q %q %q", m[2].CI, m[3].CI, m[4].CI)
	}
	if len(m[3].Falhos) != 0 || len(m[4].Falhos) != 0 {
		t.Errorf("falhos indevidos: %v %v", m[3].Falhos, m[4].Falhos)
	}
}

func TestPRsDeStatusContextFalho(t *testing.T) {
	js := `{"data":{"search":{"nodes":[{"number":1,"repository":{"name":"R"},"commits":{"nodes":[{"commit":{"statusCheckRollup":{"state":"ERROR","contexts":{"nodes":[{"__typename":"StatusContext","context":"sonar","state":"ERROR"},{"__typename":"StatusContext","context":"ok","state":"SUCCESS"}]}}}}]}}]}}}`
	p := PRsDe([]byte(js))
	if len(p) != 1 || !reflect.DeepEqual(p[0].Falhos, []string{"sonar"}) {
		t.Errorf("%+v", p)
	}
}

func TestResumoDoCI(t *testing.T) {
	r := ResumoDoCI(PRsDe([]byte(graphqlCI)), false)
	if len(r.Falhos) != 1 || len(r.Rodando) != 1 || r.Verdes != 1 || r.SemCI != 1 {
		t.Errorf("%+v", r)
	}
}

func TestResumoDoCISemRollup(t *testing.T) {
	r := ResumoDoCI([]Travado{{Number: 1}}, false)
	if r.Verdes != 0 || len(r.Falhos) != 0 || r.SemCI != 1 {
		t.Errorf("%+v", r)
	}
}

func TestResumoDoCIRascunhos(t *testing.T) {
	ps := []Travado{{Number: 1, Draft: true, CI: "FAILURE"}, {Number: 2, CI: "SUCCESS"}}
	if r := ResumoDoCI(ps, false); len(r.Falhos) != 0 || r.Verdes != 1 {
		t.Errorf("sem rascunhos: %+v", r)
	}
	if r := ResumoDoCI(ps, true); len(r.Falhos) != 1 {
		t.Errorf("com rascunhos: %+v", r)
	}
}
