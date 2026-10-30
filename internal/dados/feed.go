package dados

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type TipoFeed int

const (
	TipoDecidir TipoFeed = iota
	TipoJuntar
	TipoTravado
	TipoEsperar
	TipoLer
)

func (t TipoFeed) String() string {
	return [...]string{"decidir", "juntar", "travado", "esperar", "ler"}[t]
}

type ItemFeed struct {
	ID       string
	Tipo     TipoFeed
	Sujeito  string
	Texto    string
	Pergunta *Linha
	Merge    *Merge
	URL      string
	Pasta    string
	AbaHerdr string
}

const maxLer = 3

var reDone = regexp.MustCompile(`\bdone\b`)

// MontarFeed: só perguntas, gates, ações e relatórios de fantasminhas vivos; PRs e CI ficam nas gavetas.
func MontarFeed(e Estado) []ItemFeed {
	var itens []ItemFeed
	for i := range e.Esperando {
		p := e.Esperando[i]
		crew := crewDe(p)
		if !Vivo(e, crew) {
			continue
		}
		if m := MergeDe(p); m != nil {
			itens = append(itens, ItemFeed{ID: "q:" + p.Chave, Tipo: TipoJuntar, Sujeito: crew, Texto: p.Detalhe, Pergunta: &p, Merge: m, URL: m.URL})
		} else {
			itens = append(itens, ItemFeed{ID: "q:" + p.Chave, Tipo: TipoDecidir, Sujeito: crew, Texto: p.Detalhe, Pergunta: &p})
		}
	}

	ler := 0
	for _, t := range e.Tasks {
		if !Vivo(e, t.Rotulo) {
			continue
		}
		temPergunta := false
		for _, p := range e.Esperando {
			if strings.HasPrefix(p.Chave, t.Chave+"/") {
				temPergunta = true
				break
			}
		}
		if strings.Contains(t.Detalhe, "blocked") {
			if temPergunta {
				continue
			}
			itens = append(itens, ItemFeed{ID: "t:" + t.Chave, Tipo: TipoTravado, Sujeito: t.Rotulo, Texto: "fantasminha travado", Pasta: t.Chave, AbaHerdr: t.Extra})
		} else if reDone.MatchString(t.Detalhe) && !strings.Contains(t.Detalhe, "sem report") && ler < maxLer {
			if k := ChaveVisto(t.Chave); k != "" && e.Vistos[k] {
				continue
			}
			ler++
			itens = append(itens, ItemFeed{ID: "t:" + t.Chave, Tipo: TipoLer, Sujeito: t.Rotulo, Texto: "entregou relatório", Pasta: t.Chave, AbaHerdr: t.Extra})
		}
	}

	sort.SliceStable(itens, func(i, j int) bool { return itens[i].Tipo < itens[j].Tipo })
	return itens
}

// MontarAntigas: perguntas pendentes de fantasminha sem sessão no Herdr.
func MontarAntigas(e Estado) []ItemFeed {
	var itens []ItemFeed
	for i := range e.Esperando {
		p := e.Esperando[i]
		crew := crewDe(p)
		if Vivo(e, crew) {
			continue
		}
		itens = append(itens, ItemFeed{ID: "q:" + p.Chave, Tipo: TipoDecidir, Sujeito: crew, Texto: p.Detalhe, Pergunta: &p})
	}
	return itens
}

func Resumo(e Estado) string {
	var praVoce, travado, ativos int
	for _, i := range MontarFeed(e) {
		switch i.Tipo {
		case TipoDecidir, TipoJuntar:
			praVoce++
		case TipoTravado:
			travado++
		}
	}
	for _, t := range e.Tasks {
		if strings.Contains(t.Detalhe, "working") {
			ativos++
		}
	}
	var partes []string
	if praVoce > 0 {
		partes = append(partes, fmt.Sprintf("%d precisa de você", praVoce))
	}
	if travado > 0 {
		s := ""
		if travado > 1 {
			s = "s"
		}
		partes = append(partes, fmt.Sprintf("%d travado%s", travado, s))
	}
	if ativos > 0 {
		partes = append(partes, fmt.Sprintf("%d trabalhando", ativos))
	}
	if len(partes) == 0 {
		return "tudo calmo por aqui"
	}
	return strings.Join(partes, " · ")
}
