package dados

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Agente struct{ Status, Sessao, Pane, Aba string }

// AgentesDe lê o JSON do `herdr agent list`; nil quando a saída não é legível.
func AgentesDe(saida []byte) map[string]Agente {
	var r struct {
		Result struct {
			Agents []struct {
				Name    string `json:"name"`
				Status  string `json:"agent_status"`
				Session *struct {
					Value string `json:"value"`
				} `json:"agent_session"`
				Pane string `json:"pane_id"`
				Tab  string `json:"tab_id"`
			} `json:"agents"`
		} `json:"result"`
	}
	if json.Unmarshal(saida, &r) != nil {
		return nil
	}
	m := map[string]Agente{}
	for _, a := range r.Result.Agents {
		if a.Name == "" {
			continue
		}
		ag := Agente{Status: a.Status, Pane: a.Pane, Aba: a.Tab}
		if a.Session != nil {
			ag.Sessao = a.Session.Value
		}
		m[a.Name] = ag
	}
	return m
}

type SessaoClaude struct {
	Nome, Status, Cwd string
	Atualizado        time.Time
}

// DirClaude é FANTASMAS_CLAUDE_DIR ou ~/.claude.
func DirClaude() string {
	if d := os.Getenv("FANTASMAS_CLAUDE_DIR"); d != "" {
		return d
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".claude")
	}
	return ""
}

// SessoesClaude lê dir/*.json indexado por sessionId; arquivo ilegível é ignorado.
func SessoesClaude(dir string) map[string]SessaoClaude {
	m := map[string]SessaoClaude{}
	arqs, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, a := range arqs {
		b, err := os.ReadFile(a)
		if err != nil {
			continue
		}
		var s struct {
			ID     string `json:"sessionId"`
			Name   string `json:"name"`
			Status string `json:"status"`
			Cwd    string `json:"cwd"`
			Upd    int64  `json:"updatedAt"`
		}
		if json.Unmarshal(b, &s) != nil || s.ID == "" {
			continue
		}
		m[s.ID] = SessaoClaude{Nome: s.Name, Status: s.Status, Cwd: s.Cwd, Atualizado: time.UnixMilli(s.Upd)}
	}
	return m
}

// sessaoDe: o id do Herdr vence; sem ele, a sessão mais recente com o mesmo nome.
func sessaoDe(m map[string]SessaoClaude, nome, id string) (*SessaoClaude, string) {
	if s, ok := m[id]; ok && id != "" {
		return &s, id
	}
	var melhor *SessaoClaude
	melhorID := ""
	for k, s := range m {
		if s.Nome == nome && (melhor == nil || s.Atualizado.After(melhor.Atualizado)) {
			s := s
			melhor, melhorID = &s, k
		}
	}
	return melhor, melhorID
}

// MetasDe: nome do agente -> pasta, lido de .delegate/*/meta só para os nomes pedidos;
// para de varrer quando acha todos.
func MetasDe(raiz string, nomes map[string]Agente) map[string]string {
	m := map[string]string{}
	if len(nomes) == 0 {
		return m
	}
	metas, _ := filepath.Glob(filepath.Join(raiz, ".delegate", "*", "meta"))
	for _, p := range metas {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, l := range strings.Split(string(b), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(l), "name="); ok && v != "" {
				if _, quer := nomes[v]; quer {
					m[v] = filepath.Dir(p)
				}
				break
			}
		}
		if len(m) == len(nomes) {
			break
		}
	}
	return m
}

var reNaoAlfanum = regexp.MustCompile(`[^A-Za-z0-9]`)

func CaminhoTranscript(claudeDir, cwd, sessao string) string {
	return filepath.Join(claudeDir, "projects", reNaoAlfanum.ReplaceAllString(cwd, "-"), sessao+".jsonl")
}
