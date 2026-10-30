package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Run é um run de pipeline como vem em `ddb state --json` (runs[]).
type Run struct {
	ID          int               `json:"id"`
	Pipeline    string            `json:"pipeline"`
	Title       string            `json:"title"`
	SlugBase    string            `json:"slug_base"`
	Status      string            `json:"status"`
	Coordinator string            `json:"coordinator"`
	Inputs      map[string]string `json:"inputs"`
	IdadeMin    int               `json:"idade_min"`
	EsperaVoce  int               `json:"espera_voce"`
	Etapas      []Etapa           `json:"stages"`
}

type Gate struct {
	Item string `json:"item"`
	By   string `json:"by"`
}

type Etapa struct {
	ID      string          `json:"id"`
	Role    string          `json:"role"`
	Skill   string          `json:"skill"`
	Status  string          `json:"status"`
	Reason  string          `json:"reason"`
	Slug    string          `json:"slug"`
	Attempt int             `json:"attempt"`
	After   []string        `json:"after"`
	Gates   []Gate          `json:"gates"`
	Agent   string          `json:"agent"`
	Tab     string          `json:"tab"`
	Gone    bool            `json:"gone"`
	Report  json.RawMessage `json:"report"`
}

// Pipeline é uma linha de `ddb pipeline list --json`.
type Pipeline struct {
	Name    string   `json:"name"`
	Arquivo string   `json:"arquivo"`
	Version any      `json:"version"`
	Inputs  []string `json:"inputs"`
	Stages  []string `json:"stages"`
	Erros   []string `json:"erros"`
}

// LinksDoReport tira as URLs do report.json da etapa (campos links, url, pr e prs).
func LinksDoReport(raw json.RawMessage) []string {
	var r map[string]any
	if len(raw) == 0 || json.Unmarshal(raw, &r) != nil {
		return nil
	}
	var out []string
	var junta func(v any)
	junta = func(v any) {
		switch x := v.(type) {
		case string:
			if strings.Contains(x, "://") {
				out = append(out, x)
			}
		case []any:
			for _, y := range x {
				junta(y)
			}
		case map[string]any:
			if u, ok := x["url"]; ok {
				junta(u)
			}
		}
	}
	for _, k := range []string{"links", "url", "pr", "prs"} {
		junta(r[k])
	}
	return out
}

// ProximoDoTick lê a saída do `ddb tick --dry-run`: slug → o que o tick faria com ele.
func ProximoDoTick(saida string) map[string]string {
	out := map[string]string{}
	for _, l := range strings.Split(saida, "\n") {
		l = strings.TrimSpace(l)
		switch {
		case l == "" || strings.HasPrefix(l, "["):
		case strings.HasPrefix(l, "lançaria "):
			slug, resto, _ := strings.Cut(strings.TrimPrefix(l, "lançaria "), " ")
			out[slug] = strings.TrimSpace("lançaria " + resto)
		case strings.HasPrefix(l, "não lança "):
			slug, resto, _ := strings.Cut(strings.TrimPrefix(l, "não lança "), ": ")
			out[slug] = "não lança: " + resto
		default:
			if slug, resto, ok := strings.Cut(l, ": "); ok && !strings.Contains(slug, " ") {
				if _, ja := out[slug]; !ja {
					out[slug] = resto
				}
			}
		}
	}
	return out
}

// Pipes roda os comandos de pipeline do ddb.
type Pipes struct {
	R      Runner
	Raiz   string
	Script string
}

const timeoutPipes = 30 * time.Second

func (p *Pipes) ddb(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPipes)
	defer cancel()
	argv := append([]string{"python3", p.Script, "--raiz", p.Raiz}, args...)
	return p.R.Run(ctx, argv...)
}

func (p *Pipes) Listar() ([]Pipeline, error) {
	out, err := p.ddb("pipeline", "list", "--json")
	if err != nil {
		return nil, err
	}
	var ps []Pipeline
	if err := json.Unmarshal([]byte(out), &ps); err != nil {
		return nil, fmt.Errorf("pipeline list ilegível: %w", err)
	}
	return ps, nil
}

// Novo chama `ddb run new` e devolve a primeira linha da saída (id e slugs).
func (p *Pipes) Novo(pipeline string, entradas [][2]string, titulo string) (string, error) {
	args := []string{"run", "new", pipeline}
	for _, kv := range entradas {
		args = append(args, "--input", kv[0]+"="+kv[1])
	}
	if titulo = strings.TrimSpace(titulo); titulo != "" {
		args = append(args, "--title", titulo)
	}
	out, err := p.ddb(args...)
	l, _, _ := strings.Cut(strings.TrimSpace(out), "\n")
	return l, err
}

// Etapa é o escape manual: advance, skip ou retry.
func (p *Pipes) Etapa(run int, etapa, acao string) error {
	switch acao {
	case "advance", "skip", "retry":
	default:
		return fmt.Errorf("ação %q não é advance, skip nem retry", acao)
	}
	_, err := p.ddb("stage", fmt.Sprintf("%d:%s", run, etapa), acao)
	return err
}

func (p *Pipes) Validar(alvo string) (string, error) {
	out, err := p.ddb("pipeline", "validate", alvo)
	return strings.TrimSpace(out), err
}

// ScriptDdb é o caminho do ddb.py (FANTASMAS_DDB_SCRIPT ou o do workspace).
func ScriptDdb(raiz string) string { return scriptDdbDe(raiz) }

// FonteDdbLigada diz se FANTASMAS_FONTE=ddb.
func FonteDdbLigada() bool { return fonteDdbLigada() }
