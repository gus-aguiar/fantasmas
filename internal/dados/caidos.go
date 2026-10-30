package dados

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const maxCaidos = 8

// Caido é um fantasminha que tem meta mas não está vivo no Herdr.
type Caido struct{ Nome, Pane, SessionID, Pasta string }

var reSeguro = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

// CaidosDe lista os metas fora do Herdr que têm session_id e pane; sem leitura do Herdr, ninguém é dado como caído.
func CaidosDe(raiz string, vivos map[string]bool, limite int) []Caido {
	if vivos == nil {
		return nil
	}
	metas, _ := filepath.Glob(filepath.Join(raiz, ".delegate", "*", "meta"))
	type item struct {
		c  Caido
		em int64
	}
	var achados []item
	for _, p := range metas {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		kv := map[string]string{}
		for _, l := range strings.Split(string(b), "\n") {
			if k, v, ok := strings.Cut(strings.TrimSpace(l), "="); ok {
				kv[k] = v
			}
		}
		nome := kv["name"]
		if nome == "" || vivos[nome] || kv["session_id"] == "" || kv["pane"] == "" {
			continue
		}
		var em int64
		if st, err := os.Stat(p); err == nil {
			em = st.ModTime().UnixNano()
		}
		achados = append(achados, item{Caido{Nome: nome, Pane: kv["pane"], SessionID: kv["session_id"], Pasta: filepath.Dir(p)}, em})
	}
	sort.Slice(achados, func(i, j int) bool { return achados[i].em > achados[j].em })
	var out []Caido
	for _, a := range achados {
		if limite > 0 && len(out) >= limite {
			break
		}
		out = append(out, a.c)
	}
	return out
}

// Retomar religa a sessão do fantasminha no pane dele e devolve o nome ao Herdr.
func Retomar(ctx context.Context, r Runner, c Caido) error {
	if c.SessionID == "" || c.Pane == "" {
		return fmt.Errorf("sem session_id; retome à mão")
	}
	for _, v := range []string{c.Nome, c.SessionID, c.Pane} {
		if !reSeguro.MatchString(v) {
			return fmt.Errorf("valor inseguro no meta: %q", v)
		}
	}
	cmd := fmt.Sprintf("claude --dangerously-skip-permissions --name %s --resume %s", c.Nome, c.SessionID)
	if _, err := r.Run(ctx, "herdr", "pane", "run", c.Pane, cmd); err != nil {
		return err
	}
	_, err := r.Run(ctx, "herdr", "agent", "rename", c.Pane, c.Nome)
	return err
}
