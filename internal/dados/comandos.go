package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Comando é um disparo da gaveta comandos: abre uma aba nova num workspace do Herdr achado pelo rótulo
// e roda `claude "<Prompt>"` nela.
type Comando struct {
	ID, Rotulo, Workspace, Prompt, Prefixo, Descricao string
}

// ComandosHerdr é o catálogo, na ordem da gaveta.
var ComandosHerdr = []Comando{
	{ID: "dia-abrir", Rotulo: "/dia abrir", Workspace: "GENGAR", Prompt: "/dia abrir", Prefixo: "dia", Descricao: "começa o dia: Jira, Intralog, PRs e DayDreaming"},
	{ID: "dia-fechar", Rotulo: "/dia fechar", Workspace: "GENGAR", Prompt: "/dia fechar", Prefixo: "dia", Descricao: "fecha o dia e faz a passagem"},
	{ID: "saude-multas", Rotulo: "/saude-multas", Workspace: "saúde multas", Prompt: "/saude-multas", Prefixo: "saude", Descricao: "saúde do módulo de multas (só leitura)"},
}

// ComandoPorID devolve o comando do catálogo.
func ComandoPorID(id string) (Comando, bool) {
	for _, c := range ComandosHerdr {
		if c.ID == id {
			return c, true
		}
	}
	return Comando{}, false
}

// Disparo é um comando já lançado: a aba criada e quando.
type Disparo struct {
	ID, Aba, Pane, Nome string
	Inicio              time.Time
}

// NomeAba é o rótulo da aba nova, ex.: dia-1106-0915.
func NomeAba(prefixo string, t time.Time) string { return prefixo + "-" + t.Format("0102-1504") }

// WorkspacePorRotulo acha o id do workspace do Herdr pelo rótulo (sem diferença de caixa nem acento).
func WorkspacePorRotulo(ctx context.Context, r Runner, rotulo string) (string, error) {
	out, err := r.Run(ctx, "herdr", "workspace", "list")
	if err != nil {
		return "", err
	}
	var v struct {
		Result struct {
			Workspaces []struct {
				ID    string `json:"workspace_id"`
				Label string `json:"label"`
			} `json:"workspaces"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		return "", fmt.Errorf("herdr workspace list ilegível: %w", err)
	}
	alvo := chaveRotulo(rotulo)
	for _, w := range v.Result.Workspaces {
		if chaveRotulo(w.Label) == alvo && w.ID != "" {
			return w.ID, nil
		}
	}
	return "", fmt.Errorf("sem workspace %q no Herdr", rotulo)
}

var tiraAcento = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c")

func chaveRotulo(s string) string { return tiraAcento.Replace(strings.ToLower(strings.TrimSpace(s))) }

// Disparar abre a aba do comando no workspace dele, com cwd na raiz do Spark, e roda o claude nela.
func Disparar(ctx context.Context, r Runner, raiz string, c Comando, agora time.Time) (Disparo, error) {
	ws, err := WorkspacePorRotulo(ctx, r, c.Workspace)
	if err != nil {
		return Disparo{}, err
	}
	nome := NomeAba(c.Prefixo, agora)
	out, err := r.Run(ctx, "herdr", "tab", "create", "--workspace", ws, "--cwd", raiz, "--label", nome, "--no-focus")
	if err != nil {
		return Disparo{}, err
	}
	var v struct {
		Result struct {
			RootPane struct {
				ID string `json:"pane_id"`
			} `json:"root_pane"`
			Tab struct {
				ID string `json:"tab_id"`
			} `json:"tab"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(out), &v) != nil || v.Result.RootPane.ID == "" {
		return Disparo{}, fmt.Errorf("herdr tab create não devolveu o pane")
	}
	pane := v.Result.RootPane.ID
	if _, err := r.Run(ctx, "herdr", "pane", "run", pane, fmt.Sprintf("claude %q", c.Prompt)); err != nil {
		return Disparo{}, err
	}
	return Disparo{ID: c.ID, Aba: v.Result.Tab.ID, Pane: pane, Nome: nome, Inicio: agora}, nil
}
