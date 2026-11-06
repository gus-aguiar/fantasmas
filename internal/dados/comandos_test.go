package dados

import (
	"context"
	"strings"
	"testing"
	"time"
)

// runnerHerdr finge o herdr: devolve os workspaces e a aba criada, e grava cada chamada.
type runnerHerdr struct{ calls [][]string }

func (r *runnerHerdr) Run(_ context.Context, argv ...string) (string, error) {
	r.calls = append(r.calls, argv)
	switch strings.Join(argv[:min(3, len(argv))], " ") {
	case "herdr workspace list":
		return `{"result":{"workspaces":[{"workspace_id":"w2","label":"GENGAR"},{"workspace_id":"w9","label":"saúde multas"}]}}`, nil
	case "herdr tab create":
		return `{"result":{"root_pane":{"pane_id":"w9:p4"},"tab":{"tab_id":"w9:t3"}}}`, nil
	}
	return "", nil
}

func TestNomeAba(t *testing.T) {
	if n := NomeAba("dia", time.Date(2026, 11, 6, 9, 5, 0, 0, time.UTC)); n != "dia-1106-0905" {
		t.Fatalf("nome %s", n)
	}
}

func TestWorkspacePorRotulo(t *testing.T) {
	r := &runnerHerdr{}
	for rot, quer := range map[string]string{"GENGAR": "w2", "gengar": "w2", "saude multas": "w9", "saúde multas": "w9"} {
		if id, err := WorkspacePorRotulo(context.Background(), r, rot); err != nil || id != quer {
			t.Fatalf("%s: %q %v", rot, id, err)
		}
	}
	if _, err := WorkspacePorRotulo(context.Background(), r, "nada"); err == nil {
		t.Fatal("rótulo inexistente devia falhar")
	}
}

func TestDispararMontaHerdr(t *testing.T) {
	agora := time.Date(2026, 11, 6, 14, 30, 0, 0, time.UTC)
	casos := []struct {
		id, ws, nome, cmd string
	}{
		{"dia-abrir", "w2", "dia-1106-1430", `claude "/dia abrir"`},
		{"dia-fechar", "w2", "dia-1106-1430", `claude "/dia fechar"`},
		{"saude-multas", "w9", "saude-1106-1430", `claude "/saude-multas"`},
	}
	for _, c := range casos {
		r := &runnerHerdr{}
		k, ok := ComandoPorID(c.id)
		if !ok {
			t.Fatalf("sem %s", c.id)
		}
		d, err := Disparar(context.Background(), r, "/spark", k, agora)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.calls) != 3 {
			t.Fatalf("%s: chamadas %v", c.id, r.calls)
		}
		criar := strings.Join(r.calls[1], " ")
		if criar != "herdr tab create --workspace "+c.ws+" --cwd /spark --label "+c.nome+" --no-focus" {
			t.Fatalf("%s: %s", c.id, criar)
		}
		run := r.calls[2]
		if strings.Join(run[:4], " ") != "herdr pane run w9:p4" || run[4] != c.cmd {
			t.Fatalf("%s: %v", c.id, run)
		}
		if d.Aba != "w9:t3" || d.Nome != c.nome || !d.Inicio.Equal(agora) || d.ID != c.id {
			t.Fatalf("%s: disparo %+v", c.id, d)
		}
	}
}
