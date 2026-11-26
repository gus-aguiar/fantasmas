package dados

import (
	"strings"
	"testing"
	"time"
)

func TestQueryMergeadosDe(t *testing.T) {
	t.Setenv(ChaveGHOrg, "acme-org")
	q := QueryMergeadosDe("eu", time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC))
	for _, s := range []string{"is:pr is:merged author:eu merged:>=2026-11-26 org:acme-org", "commits { totalCount }", "mergedAt"} {
		if !strings.Contains(q, s) {
			t.Errorf("query sem %q:\n%s", s, q)
		}
	}
}

func TestMergeadosDe(t *testing.T) {
	js := `{"data":{"search":{"nodes":[
	 {"number":3,"title":"Placar","url":"u3","headRefName":"feat/d21-placar","mergedAt":"2026-11-26T10:00:00Z","repository":{"name":"d21"},"commits":{"totalCount":4}},
	 {},
	 {"number":7,"title":"VT-19826 preço","url":"u7","headRefName":"x","mergedAt":"2026-11-26T12:00:00Z","repository":{"name":"mtg"},"commits":{"totalCount":2}}]}}}`
	ms := MergeadosDe([]byte(js))
	if len(ms) != 2 || ms[0].Number != 7 || ms[1].Commits != 4 || ms[1].Branch != "feat/d21-placar" {
		t.Fatalf("ms=%+v", ms)
	}
	if MergeadosDe([]byte("lixo")) != nil {
		t.Fatal("json ilegível devia dar nil")
	}
}

func TestFeitoHojeAgrupaPorMissao(t *testing.T) {
	agora := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC)
	hoje := agora.Add(-2 * time.Hour)
	e := Estado{
		Missoes: []Missao{
			{Slug: "d21-placar", Sidequests: []int{13}},
			{Slug: "mtg-preco-deck", Cards: []string{"VT-19826"}},
		},
		Sidequests: []Sidequest{{ID: 13, Inputs: map[string]string{"card": "VT-19597"}}},
		Mergeados: []Mergeado{
			{Repo: "d21", Number: 1, Branch: "feat/d21-placar-api", Commits: 3, Em: hoje},
			{Repo: "mtg", Number: 2, Title: "VT-19826: preço", Commits: 2, Em: hoje},
			{Repo: "d21", Number: 3, Branch: "fix/vt-19597", Commits: 1, Em: hoje},
			{Repo: "vitrine", Number: 4, Branch: "docs/readme", Commits: 5, Em: hoje},
			{Repo: "velho", Number: 5, Branch: "d21-placar", Commits: 9, Em: agora.Add(-30 * time.Hour)},
		},
	}
	f := FeitoHoje(e, agora)
	if len(f) != 3 {
		t.Fatalf("esperava d21, mtg e sem missão: %+v", f)
	}
	if f[0].Missao != "d21-placar" || len(f[0].PRs) != 2 || f[0].Commits != 4 {
		t.Fatalf("d21 (slug no branch + card da sidequest): %+v", f[0])
	}
	if f[1].Missao != "mtg-preco-deck" || f[1].Commits != 2 {
		t.Fatalf("mtg (card no título): %+v", f[1])
	}
	if f[2].Missao != "" || f[2].PRs[0].Number != 4 {
		t.Fatalf("sem missão por último: %+v", f[2])
	}
}
