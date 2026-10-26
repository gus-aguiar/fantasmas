package dados

import "strings"

type Linha struct{ Chave, Rotulo, Detalhe, Extra, Merge string }

func TSV(saida string) []Linha {
	var out []Linha
	for _, l := range strings.Split(saida, "\n") {
		l = strings.TrimSuffix(l, "\r")
		if strings.TrimSpace(l) == "" {
			continue
		}
		c := strings.Split(l, "\t")
		x := Linha{Chave: c[0], Rotulo: "?"}
		if len(c) > 1 {
			x.Rotulo = c[1]
		}
		if len(c) > 2 {
			x.Detalhe = c[2]
		}
		if len(c) > 3 {
			x.Extra = c[3]
		}
		if len(c) > 4 {
			x.Merge = c[4]
		}
		out = append(out, x)
	}
	return out
}
