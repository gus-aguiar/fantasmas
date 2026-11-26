package dados

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Licao é um erro que os fantasminhas corrigiram, com quantas vezes ele se repetiu.
type Licao struct {
	// Erro é o texto do erro como apareceu da primeira vez.
	Erro string
	// Vezes soma os registros de .delegate/_licoes.jsonl e as novas tentativas das etapas de sidequest.
	Vezes int
	// Correcao é a última correção registrada; vazia quando o erro só apareceu como retry.
	Correcao string
	Ultima   time.Time
}

const (
	janelaLicoes = 30 * 24 * time.Hour
	caudaLicoes  = 256 * 1024
)

var (
	reNumeros = regexp.MustCompile(`\d+`)
	reEspacos = regexp.MustCompile(`\s+`)
)

// chaveErro junta variações do mesmo erro: caixa, espaços e números (linha, porta, id) não contam.
func chaveErro(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reNumeros.ReplaceAllString(s, "#")
	return reEspacos.ReplaceAllString(s, " ")
}

// LicoesDe junta as lições dos últimos 30 dias: cada linha de .delegate/_licoes.jsonl
// ({"at","crew","missao","erro","correcao"}, gravada pelo workspace) conta uma vez, e cada
// etapa de sidequest com motivo e tentativa > 0 conta as tentativas a mais. Do mais repetido ao menos.
func LicoesDe(caminho string, sidequests []Sidequest, agora time.Time) []Licao {
	por := map[string]*Licao{}
	add := func(erro, correcao string, vezes int, em time.Time) {
		k := chaveErro(erro)
		if k == "" || vezes <= 0 {
			return
		}
		l, ok := por[k]
		if !ok {
			l = &Licao{Erro: strings.TrimSpace(erro)}
			por[k] = l
		}
		l.Vezes += vezes
		if correcao != "" && !em.Before(l.Ultima) {
			l.Correcao = strings.TrimSpace(correcao)
		}
		if em.After(l.Ultima) {
			l.Ultima = em
		}
	}
	if buf := LerCauda(caminho, caudaLicoes); len(buf) > 0 {
		sc := bufio.NewScanner(bytes.NewReader(buf))
		sc.Buffer(make([]byte, 0, 64*1024), caudaLicoes)
		for sc.Scan() {
			var x struct {
				At       time.Time `json:"at"`
				Erro     string    `json:"erro"`
				Correcao string    `json:"correcao"`
			}
			if json.Unmarshal(sc.Bytes(), &x) != nil || (!x.At.IsZero() && agora.Sub(x.At) > janelaLicoes) {
				continue
			}
			add(x.Erro, x.Correcao, 1, x.At)
		}
	}
	for _, s := range sidequests {
		for _, e := range s.Etapas {
			if e.Attempt > 0 && e.Reason != "" {
				add(e.Reason, "", e.Attempt, time.Time{})
			}
		}
	}
	out := make([]Licao, 0, len(por))
	for _, l := range por {
		out = append(out, *l)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Vezes != out[j].Vezes {
			return out[i].Vezes > out[j].Vezes
		}
		if !out[i].Ultima.Equal(out[j].Ultima) {
			return out[i].Ultima.After(out[j].Ultima)
		}
		return out[i].Erro < out[j].Erro
	})
	return out
}
