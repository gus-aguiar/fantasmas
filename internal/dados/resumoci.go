package dados

// ResumoCI agrupa os PRs pelo estado do CI; PR sem rollup não entra em Verdes nem em Falhos.
type ResumoCI struct {
	Falhos, Rodando []Travado
	Verdes, SemCI   int
}

// ResumoDoCI deixa os rascunhos de fora, salvo com rascunhos=true (mesmo filtro da gaveta PRs).
func ResumoDoCI(prs []Travado, rascunhos bool) ResumoCI {
	var r ResumoCI
	for _, p := range prs {
		if p.Draft && !rascunhos {
			continue
		}
		switch p.CI {
		case "FAILURE", "ERROR":
			r.Falhos = append(r.Falhos, p)
		case "PENDING", "EXPECTED":
			r.Rodando = append(r.Rodando, p)
		case "SUCCESS":
			r.Verdes++
		default:
			r.SemCI++
		}
	}
	return r
}
