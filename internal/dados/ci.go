package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

var reposCI = []string{"PrologWebReact", "PrologWebService", "PrologE2E", "GlorpDesignSystem"}

const maxCIPorRepo = 5

type run struct {
	DatabaseID   int64  `json:"databaseId"`
	DisplayTitle string `json:"displayTitle"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
	CreatedAt    string `json:"createdAt"`
	HeadBranch   string `json:"headBranch"`
}

const camposRun = "databaseId,displayTitle,status,conclusion,workflowName,createdAt,headBranch"

// coletarCI: execuções das suas branches (dos PRs abertos) e as disparadas pelo seu login.
func (c *Coletor) coletarCI(ctx context.Context, prs []Travado) []Linha {
	branches := map[string]bool{}
	repos := append([]string{}, reposCI...)
	visto := map[string]bool{}
	for _, r := range repos {
		visto[r] = true
	}
	for _, p := range prs {
		if p.Branch != "" {
			branches[p.Branch] = true
		}
		if p.Repo != "" && p.Repo != "?" && !visto[p.Repo] {
			visto[p.Repo] = true
			repos = append(repos, p.Repo)
		}
	}
	por := make([][]Linha, len(repos))
	var wg sync.WaitGroup
	for i, r := range repos {
		wg.Add(1)
		go func(i int, r string) {
			defer wg.Done()
			por[i] = c.ciDoRepo(ctx, r, branches)
		}(i, r)
	}
	wg.Wait()
	var out []Linha
	for _, l := range por {
		out = append(out, l...)
	}
	return out
}

func (c *Coletor) ciDoRepo(ctx context.Context, repo string, branches map[string]bool) []Linha {
	base := []string{"gh", "run", "list", "--repo", "prologapp/" + repo, "--limit", "20", "--json", camposRun}
	var geral, meus string
	var okG, okM bool
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); geral, okG = c.rodar(ctx, timeoutPRs, base...) }()
	if c.autor != "" {
		wg.Add(1)
		go func() { defer wg.Done(); meus, okM = c.rodar(ctx, timeoutPRs, append(base, "--user", c.autor)...) }()
	}
	wg.Wait()
	sel := map[int64]run{}
	if okG {
		for _, r := range runsDe(geral) {
			if branches[r.HeadBranch] {
				sel[r.DatabaseID] = r
			}
		}
	}
	if okM {
		for _, r := range runsDe(meus) {
			sel[r.DatabaseID] = r
		}
	}
	rs := make([]run, 0, len(sel))
	for _, r := range sel {
		rs = append(rs, r)
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i].CreatedAt > rs[j].CreatedAt })
	if len(rs) > maxCIPorRepo {
		rs = rs[:maxCIPorRepo]
	}
	out := make([]Linha, 0, len(rs))
	for _, r := range rs {
		out = append(out, LinhaCI(repo, r))
	}
	return out
}

func runsDe(s string) []run {
	var rs []run
	if json.Unmarshal([]byte(s), &rs) != nil {
		return nil
	}
	return rs
}

// LinhaCI mantém o formato do coleta.sh ci.
func LinhaCI(repo string, r run) Linha {
	conc := r.Conclusion
	if conc == "" {
		conc = "-"
	}
	t := []rune(r.DisplayTitle)
	if len(t) > 38 {
		t = t[:38]
	}
	return Linha{
		Chave:   fmt.Sprintf("https://github.com/prologapp/%s/actions/runs/%d", repo, r.DatabaseID),
		Rotulo:  repo + " · " + r.WorkflowName,
		Detalhe: r.Status + "/" + conc + " · " + string(t),
	}
}
