package dados

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Mergeado é um PR seu mergeado no dia, com quantos commits levou.
type Mergeado struct {
	Repo, Title, URL, Branch string
	Number, Commits          int
	Em                       time.Time
}

// QueryMergeadosDe busca os PRs do autor mergeados a partir do dia de desde (fuso local), na mesma org dos PRs abertos.
func QueryMergeadosDe(autor string, desde time.Time) string {
	filtro := "author:" + autor + " merged:>=" + desde.Format("2006-01-02")
	if org := strings.TrimSpace(os.Getenv(ChaveGHOrg)); org != "" {
		filtro += " org:" + org
	}
	return fmt.Sprintf(`
query {
  search(query: "is:pr is:merged %s", type: ISSUE, first: 50) {
    nodes { ... on PullRequest { number title url headRefName mergedAt repository { name } commits { totalCount } } }
  }
}`, filtro)
}

// MergeadosDe lê a resposta do QueryMergeadosDe, do mais recente para o mais antigo.
func MergeadosDe(jsonGraphQL []byte) []Mergeado {
	var r struct {
		Data struct {
			Search struct {
				Nodes []struct {
					Number      int       `json:"number"`
					Title       string    `json:"title"`
					URL         string    `json:"url"`
					HeadRefName string    `json:"headRefName"`
					MergedAt    time.Time `json:"mergedAt"`
					Repository  struct {
						Name string `json:"name"`
					} `json:"repository"`
					Commits struct {
						TotalCount int `json:"totalCount"`
					} `json:"commits"`
				} `json:"nodes"`
			} `json:"search"`
		} `json:"data"`
	}
	if json.Unmarshal(jsonGraphQL, &r) != nil {
		return nil
	}
	var out []Mergeado
	for _, n := range r.Data.Search.Nodes {
		if n.Number == 0 {
			continue
		}
		out = append(out, Mergeado{Repo: n.Repository.Name, Title: n.Title, URL: n.URL, Branch: n.HeadRefName,
			Number: n.Number, Commits: n.Commits.TotalCount, Em: n.MergedAt})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Em.After(out[j].Em) })
	return out
}

// FeitoMissao é o resumo do dia de uma missão; Missao vazia junta o que não casou com nenhuma.
type FeitoMissao struct {
	Missao  string
	PRs     []Mergeado
	Commits int
}

// FeitoHoje agrupa os PRs mergeados no dia de agora por missão. Um PR é da missão quando o branch
// ou o título cita o slug dela, um card dela ou o card de uma sidequest dela.
func FeitoHoje(e Estado, agora time.Time) []FeitoMissao {
	ano, mes, dia := agora.Date()
	inicio := time.Date(ano, mes, dia, 0, 0, 0, 0, agora.Location())
	cardsSQ := map[int]string{}
	for _, s := range e.Sidequests {
		cardsSQ[s.ID] = s.Inputs["card"]
	}
	marcas := func(m Missao) []string {
		ms := []string{m.Slug}
		ms = append(ms, m.Cards...)
		for _, id := range m.Sidequests {
			ms = append(ms, cardsSQ[id])
		}
		return ms
	}
	porMissao := map[string]*FeitoMissao{}
	var ordem []string
	for _, p := range e.Mergeados {
		if p.Em.Before(inicio) {
			continue
		}
		dono := ""
		alvo := strings.ToLower(p.Branch + " " + p.Title)
		for _, m := range e.Missoes {
			for _, mc := range marcas(m) {
				if mc != "" && strings.Contains(alvo, strings.ToLower(mc)) {
					dono = m.Slug
					break
				}
			}
			if dono != "" {
				break
			}
		}
		f, ok := porMissao[dono]
		if !ok {
			f = &FeitoMissao{Missao: dono}
			porMissao[dono] = f
			ordem = append(ordem, dono)
		}
		f.PRs = append(f.PRs, p)
		f.Commits += p.Commits
	}
	out := make([]FeitoMissao, 0, len(ordem))
	for _, k := range ordem {
		if k != "" {
			out = append(out, *porMissao[k])
		}
	}
	if f, ok := porMissao[""]; ok {
		out = append(out, *f)
	}
	return out
}
