package dados

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Rollup struct {
	State string `json:"state"`
}

type NoCommit struct {
	Commit struct {
		StatusCheckRollup *Rollup `json:"statusCheckRollup"`
	} `json:"commit"`
}

// No é o nó GraphQL de um PullRequest; campos ausentes ficam no zero.
type No struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	IsDraft        bool   `json:"isDraft"`
	HeadRefName    string `json:"headRefName"`
	Mergeable      string `json:"mergeable"`
	ReviewDecision string `json:"reviewDecision"`
	Repository     struct {
		Name string `json:"name"`
	} `json:"repository"`
	Commits struct {
		Nodes []NoCommit `json:"nodes"`
	} `json:"commits"`
}

func QueryDe(autor string) string {
	return fmt.Sprintf(`
query {
  search(query: "is:pr is:open author:%s org:prologapp", type: ISSUE, first: 60) {
    nodes { ... on PullRequest {
      number title url isDraft mergeable reviewDecision headRefName
      repository { name }
      commits(last: 1) { nodes { commit { statusCheckRollup { state } } } }
    } }
  }
}`, autor)
}

func MotivoDe(n No) string {
	ci := ""
	if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
		ci = n.Commits.Nodes[0].Commit.StatusCheckRollup.State
	}
	rascunho := ""
	if n.IsDraft {
		rascunho = " · rascunho"
	}
	switch {
	case n.Mergeable == "CONFLICTING":
		return "conflito" + rascunho
	case n.ReviewDecision == "CHANGES_REQUESTED":
		return "mudanças pedidas" + rascunho
	case ci == "FAILURE" || ci == "ERROR":
		return "CI vermelho" + rascunho
	case n.IsDraft:
		return "rascunho"
	case n.ReviewDecision == "APPROVED" && n.Mergeable == "MERGEABLE":
		return "aprovado · mergear"
	}
	return "aguardando revisão"
}

var reAviso = regexp.MustCompile(`^(conflito|mudanças pedidas|CI vermelho)`)

// TomDoPR: ok = pronto para mergear, aviso = pede ação sua, neutro = esperando outra coisa.
func TomDoPR(motivo string) string {
	switch {
	case motivo == "aprovado · mergear":
		return "ok"
	case reAviso.MatchString(motivo):
		return "aviso"
	}
	return "neutro"
}

var ordemTom = map[string]int{"ok": 0, "aviso": 1, "neutro": 2}

func PRsDe(jsonGraphQL []byte) []Travado {
	var r struct {
		Data struct {
			Search struct {
				Nodes []No `json:"nodes"`
			} `json:"search"`
		} `json:"data"`
	}
	if json.Unmarshal(jsonGraphQL, &r) != nil {
		return nil
	}
	var out []Travado
	for _, n := range r.Data.Search.Nodes {
		if n.Number == 0 {
			continue
		}
		repo := n.Repository.Name
		if repo == "" {
			repo = "?"
		}
		out = append(out, Travado{Repo: repo, Number: n.Number, Motivo: MotivoDe(n), Title: n.Title, URL: n.URL, Draft: n.IsDraft, Branch: n.HeadRefName})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return ordemTom[TomDoPR(out[i].Motivo)] < ordemTom[TomDoPR(out[j].Motivo)]
	})
	return out
}

func MergeDe(l Linha) *Merge {
	if l.Merge == "" {
		return nil
	}
	var m Merge
	if json.Unmarshal([]byte(l.Merge), &m) != nil {
		return nil
	}
	return &m
}

var reTituloPrefixo = regexp.MustCompile(`^\[[^\]]+\]\s*`)

func limparTitulo(t string) string { return reTituloPrefixo.ReplaceAllString(t, "") }

func semPrefixo(s, p string) string { return strings.Replace(s, p, "", 1) }
