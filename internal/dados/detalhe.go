package dados

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Opcao struct {
	ID, Label, Effect string
	Padrao, Livre     bool
}

type Link struct{ Label, URL string }

type LinhaCompara struct{ Item, A, B, Status, Tag string }

// Aberto é o detalhe de uma pergunta ou relatório; Uid é a chave que o Responder recebe.
type Aberto struct {
	Chave, Uid, Tipo, Crew, Key, Area   string
	Kind, Peso, Titulo                  string
	Contexto, Explicacao, Comandos      []string
	Opcoes                              []Opcao
	Documentos                          []string
	Previa                              []string
	PreviaDe, Recomendacao, Coordenador string
	Destrava                            string
	Links                               []Link
	Status                              string
	Resumo                              []string
	Relatorio, Quando                   string
	Merge                               *Merge
	Diagrama, LegendaDiagrama           string
	ComparaTitulo                       string
	ComparaCab                          []string
	Compara                             []LinhaCompara
}

var opcoesRelatorio = []Opcao{{ID: "lido", Label: "Lido"}, {ID: "acao", Label: "Lido, e tenho pedido", Livre: true}}

type bruto = map[string]any

func s(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(x)
	}
}

func linhas(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, len(x))
		for i, e := range x {
			out[i] = s(e)
		}
		return out
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	default:
		return []string{s(x)}
	}
}

func textoRec(v any) string {
	if m, ok := v.(bruto); ok {
		if o, ok := m["option"]; ok && o != nil {
			return s(o)
		}
		return s(m["text"])
	}
	return s(v)
}

func linksDe(v any) []Link {
	var out []Link
	arr, _ := v.([]any)
	for _, e := range arr {
		m, ok := e.(bruto)
		if !ok || s(m["url"]) == "" {
			continue
		}
		label := s(m["url"])
		if m["label"] != nil {
			label = s(m["label"])
		}
		out = append(out, Link{Label: label, URL: s(m["url"])})
	}
	return out
}

var (
	livreInicio = regexp.MustCompile(`(?i)^(ajust|segurar|refazer$|não aprovo|outr[ao]\b|quero mudar)`)
	livreEm     = regexp.MustCompile(`(?i)com ajustes?\b|\(escrev`)
	reMD        = regexp.MustCompile("(?:^|[\\s`'\"(])((?:/|docs/|\\.delegate/|shared/)[^\\s`'\"(),;]+\\.md)")
	reFimMD     = regexp.MustCompile(`[.:]+$`)
)

func EhLivre(label string, free bool) bool {
	r := strings.TrimSpace(label)
	return free || livreEm.MatchString(r) || (len(strings.Fields(r)) <= 4 && livreInicio.MatchString(r))
}

func DocumentosDe(item map[string]any) []string {
	textos := append(append(linhas(item["context"]), linhas(item["explain"])...), s(item["unlocks"]), s(item["recommendation"]))
	vistos := map[string]bool{}
	var out []string
	add := func(d string) {
		if strings.HasSuffix(d, ".md") && !vistos[d] {
			vistos[d] = true
			out = append(out, d)
		}
	}
	for _, d := range linhas(item["plan"]) {
		add(d)
	}
	for _, t := range textos {
		for _, m := range reMD.FindAllStringSubmatch(t, -1) {
			add(reFimMD.ReplaceAllString(m[1], ""))
		}
	}
	return out
}

func PlanoDe(docs []string) string {
	for _, d := range docs {
		if strings.HasSuffix(d, "_plan.md") {
			return d
		}
	}
	if len(docs) > 0 {
		return docs[0]
	}
	return ""
}

func PreviaDe(texto string, max int) []string {
	var out []string
	codigo := false
	for _, l := range strings.Split(texto, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			codigo = !codigo
			continue
		}
		if codigo || strings.TrimSpace(l) == "" || strings.HasPrefix(l, "---") {
			continue
		}
		out = append(out, strings.TrimRight(l, " \t\r"))
		if len(out) >= max {
			break
		}
	}
	return out
}

func Separar(chave string) (arquivo, id string) {
	i := strings.LastIndex(chave, "#")
	if i < 0 {
		return chave, ""
	}
	return chave[:i], chave[i+1:]
}

func PerguntaDe(chave string, questions []byte, id string) (*Aberto, error) {
	var q bruto
	if err := json.Unmarshal(questions, &q); err != nil {
		return nil, err
	}
	itens, _ := q["items"].([]any)
	var i bruto
	for _, e := range itens {
		if m, ok := e.(bruto); ok && s(m["id"]) == id {
			i = m
			break
		}
	}
	if i == nil {
		return nil, nil
	}
	kind := s(i["kind"])
	if kind == "" {
		kind = "question"
	}
	ops, _ := i["options"].([]any)
	if len(ops) == 0 && kind == "action" {
		ops = []any{
			bruto{"id": "feito", "label": "Feito"},
			bruto{"id": "nao", "label": "Não vou fazer", "free": true},
		}
	}
	a := &Aberto{
		Chave: chave, Uid: chave, Tipo: "pergunta", Crew: s(q["crew"]), Key: s(q["key"]), Area: s(q["area"]), Kind: kind,
		Peso: s(i["peso"]), Titulo: s(i["title"]),
		Contexto: linhas(i["context"]), Explicacao: linhas(i["explain"]), Comandos: linhas(i["commands"]),
		Documentos:   DocumentosDe(i),
		Recomendacao: textoRec(i["recommendation"]),
		Coordenador:  s(i["coordinator"]), Destrava: s(i["unlocks"]),
		Links: linksDe(i["links"]), Quando: s(i["asked_at"]),
		Diagrama: s(i["diagram"]), LegendaDiagrama: s(i["diagram_caption"]),
		ComparaTitulo: s(i["compare_title"]), ComparaCab: linhas(i["compare_head"]),
	}
	if cs, ok := i["compare"].([]any); ok {
		for _, e := range cs {
			if c, ok := e.(bruto); ok {
				a.Compara = append(a.Compara, LinhaCompara{Item: s(c["item"]), A: s(c["a"]), B: s(c["b"]), Status: s(c["status"]), Tag: s(c["tag"])})
			}
		}
	}
	for _, e := range ops {
		o, _ := e.(bruto)
		free, _ := o["free"].(bool)
		padrao, _ := o["default"].(bool)
		label := s(o["label"])
		a.Opcoes = append(a.Opcoes, Opcao{ID: s(o["id"]), Label: label, Effect: s(o["effect"]), Padrao: padrao, Livre: EhLivre(label, free)})
	}
	if m, ok := i["merge"].(bruto); ok {
		if b, err := json.Marshal(m); err == nil {
			var mg Merge
			if json.Unmarshal(b, &mg) == nil {
				a.Merge = &mg
			}
		}
	}
	if m, ok := i["merge"].(bruto); ok && s(m["url"]) != "" {
		a.Links = append(a.Links, Link{Label: fmt.Sprintf("PR #%s %s", s(m["pr"]), s(m["repo"])), URL: s(m["url"])})
	}
	return a, nil
}

func RelatorioDe(chave string, report []byte) (*Aberto, error) {
	var r bruto
	if err := json.Unmarshal(report, &r); err != nil {
		return nil, err
	}
	return &Aberto{
		Chave: chave, Uid: s(r["crew"]) + "#report", Tipo: "relatorio", Kind: "report", Crew: s(r["crew"]),
		Key: s(r["key"]), Area: s(r["area"]), Opcoes: append([]Opcao(nil), opcoesRelatorio...),
		Titulo: strings.TrimSpace("Relatório " + s(r["status"])),
		Status: s(r["status"]), Resumo: linhas(r["summary"]), Links: linksDe(r["links"]),
		Relatorio: s(r["report_path"]), Quando: s(r["at"]),
	}, nil
}

// RelatorioMdDe monta o relatório de uma task que só entregou report.md: resumo = primeiras linhas não vazias fora de blocos de código.
func RelatorioMdDe(chave, caminho, md string) *Aberto {
	var res []string
	cerca := false
	for _, l := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			cerca = !cerca
			continue
		}
		if cerca || strings.TrimSpace(l) == "" {
			continue
		}
		if res = append(res, strings.TrimSpace(l)); len(res) == 40 {
			break
		}
	}
	return &Aberto{
		Chave: chave, Uid: "#report", Tipo: "relatorio", Kind: "report", Opcoes: append([]Opcao(nil), opcoesRelatorio...),
		Titulo: "Relatório", Status: "done", Resumo: res, Relatorio: caminho,
	}
}
