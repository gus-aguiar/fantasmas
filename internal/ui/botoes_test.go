package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/x/ansi"
)

func TestGavetaMaquinaTemBotoes(t *testing.T) {
	e := dados.Estado{At: time.Now(),
		Daydreamer: dados.Daydreamer{Nome: "c-sonho", Status: "working", Ultima: time.Date(2026, 10, 23, 9, 5, 0, 0, time.Local)},
		MCP:        dados.ProgressoMCP{Inicio: time.Now(), Total: 4, Rodando: true, Sessoes: []dados.SessaoMCP{{Pane: "p", Nome: "g", Detalhe: "ok", Tom: "ok"}}},
	}
	m := Novo(nil, Acoes{}, time.Minute).ComEstado(e).irGaveta(indiceGaveta("maquina"))
	tela := ansi.Strip(m.Render(110, 30))
	for _, s := range []string{"atualizar DayDreaming", "trabalhando", "última 23/10 09:05", "reconectando…", "1/4 sessões"} {
		if !strings.Contains(tela, s) {
			t.Errorf("falta %q", s)
		}
	}
}
