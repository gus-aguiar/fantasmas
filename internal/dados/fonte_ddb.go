package dados

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxTasksDdb = 14

// ParteDdb é o pedaço do Estado que o `ddb state` substitui.
type ParteDdb struct {
	Tasks, Esperando, Opcoes []Linha
	Mural                    []Aviso
}

type estadoDdbJSON struct {
	Tasks []struct {
		Chave     string `json:"chave"`
		Nome      string `json:"nome"`
		Tab       string `json:"tab"`
		Marca     string `json:"marca"`
		Pendentes int    `json:"pendentes"`
		Total     int    `json:"total"`
		Quando    string `json:"quando"`
		Herdr     string `json:"herdr"`
	} `json:"tasks"`
	Pendentes []struct {
		Chave   string          `json:"chave"`
		Rotulo  string          `json:"rotulo"`
		Detalhe string          `json:"detalhe"`
		Extra   string          `json:"extra"`
		Merge   json.RawMessage `json:"merge"`
	} `json:"pendentes"`
	OpcoesRaw map[string][]opcaoDdb `json:"opcoes"`
	Mural     []struct {
		At    string `json:"at"`
		Autor string `json:"autor"`
		Texto string `json:"texto"`
		Ate   string `json:"ate"`
	} `json:"mural"`
}

type opcaoDdb struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Marcas string `json:"marcas"`
}

// DeDdb converte a saída de `ddb state --json` nos mesmos []Linha que vêm do TSV do delegate.py.
func DeDdb(saida []byte, agora time.Time) (ParteDdb, error) {
	var s estadoDdbJSON
	if err := json.Unmarshal(saida, &s); err != nil {
		return ParteDdb{}, fmt.Errorf("state --json ilegível: %w", err)
	}
	var p ParteDdb
	for i, t := range s.Tasks {
		if i >= maxTasksDdb {
			break
		}
		det := fmt.Sprintf("%s · %d/%d perguntas · %s", t.Marca, t.Pendentes, t.Total, t.Quando)
		if t.Herdr != "" {
			det += " · " + t.Herdr
		}
		p.Tasks = append(p.Tasks, Linha{Chave: t.Chave, Rotulo: t.Nome, Detalhe: det, Extra: t.Tab})
	}
	for _, q := range s.Pendentes {
		l := Linha{Chave: q.Chave, Rotulo: q.Rotulo, Detalhe: q.Detalhe, Extra: q.Extra}
		if m := bytes.TrimSpace(q.Merge); len(m) > 0 && string(m) != "null" {
			var b bytes.Buffer
			if json.Compact(&b, m) == nil {
				l.Merge = b.String()
			}
		}
		p.Esperando = append(p.Esperando, l)
		for _, o := range s.OpcoesRaw[q.Chave] {
			p.Opcoes = append(p.Opcoes, Linha{Chave: q.Chave, Rotulo: o.ID, Detalhe: o.Label, Extra: o.Marcas})
		}
	}
	for _, m := range s.Mural {
		b, _ := json.Marshal(m)
		if a, ok := avisoDe(b, agora); ok {
			p.Mural = append(p.Mural, a)
		}
	}
	sort.SliceStable(p.Mural, func(i, j int) bool { return p.Mural[i].At.After(p.Mural[j].At) })
	if len(p.Mural) > 3 {
		p.Mural = p.Mural[:3]
	}
	return p, nil
}

func fonteDdbLigada() bool { return os.Getenv("FANTASMAS_FONTE") == "ddb" }

func scriptDdbDe(raiz string) string {
	if s := os.Getenv("FANTASMAS_DDB_SCRIPT"); s != "" {
		return s
	}
	return filepath.Join(raiz, "scripts", "delegate", "ddb.py")
}

func (c *Coletor) scriptDdb() string { return scriptDdbDe(c.Raiz) }

// lerDdb roda import + state. O erro vira a linha de "falta nesta máquina".
func (c *Coletor) lerDdb(ctx context.Context) (ParteDdb, error) {
	py := c.scriptDdb()
	if st, err := os.Stat(py); err != nil || st.IsDir() {
		return ParteDdb{}, fmt.Errorf("%s não existe", py)
	}
	if _, ok := c.rodar(ctx, timeoutColetor, "python3", py, "--raiz", c.Raiz, "import"); !ok {
		return ParteDdb{}, fmt.Errorf("import falhou")
	}
	out, ok := c.rodar(ctx, timeoutColetor, "python3", py, "--raiz", c.Raiz, "state", "--json")
	if !ok {
		return ParteDdb{}, fmt.Errorf("state falhou")
	}
	p, err := DeDdb([]byte(out), time.Now())
	if err != nil {
		return ParteDdb{}, err
	}
	return p, nil
}

func faltaDdb(r Recursos) []string {
	if r.FalhaDdb == "" {
		return nil
	}
	return []string{"ddb: " + strings.TrimSpace(r.FalhaDdb) + " (usando o caminho antigo)"}
}
