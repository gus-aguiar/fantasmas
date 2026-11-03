package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/x/ansi"
)

func TestSettingsTemBotoes(t *testing.T) {
	e := dados.Estado{At: time.Now(),
		Daydreamer: dados.Daydreamer{Nome: "c-sonho", Status: "working", Ultima: time.Date(2026, 10, 23, 9, 5, 0, 0, time.Local)},
		MCP:        dados.ProgressoMCP{Inicio: time.Now(), Total: 4, Rodando: true, Sessoes: []dados.SessaoMCP{{Pane: "p", Nome: "g", Detalhe: "ok", Tom: "ok"}}},
	}
	m := Novo(nil, Acoes{}, time.Minute).ComEstado(e).irGaveta(indiceGaveta("settings"))
	tela := ansi.Strip(m.Render(110, 30))
	for _, s := range []string{"ações", "atualizar DayDreaming", "trabalhando", "última 23/10 09:05", "reconectando…", "1/4 sessões"} {
		if !strings.Contains(tela, s) {
			t.Errorf("falta %q", s)
		}
	}
	if wt := ansi.Strip(m.irGaveta(indiceGaveta("worktrees")).Render(110, 30)); strings.Contains(wt, "atualizar") {
		t.Errorf("worktrees deveria ter só os worktrees:\n%s", wt)
	}
	maq := ansi.Strip(m.irGaveta(indiceGaveta("maquina")).Render(110, 30))
	if strings.Contains(maq, "DayDreaming") || strings.Contains(maq, "reconectar MCPs") {
		t.Errorf("máquina deveria ser só monitoramento:\n%s", maq)
	}
}

func TestAtualizarPedeSegundoU(t *testing.T) {
	chamou := 0
	ac := Acoes{Atualizar: func() (dados.Resultado, error) {
		chamou++
		return dados.Resultado{Novos: 2, Trocou: true, Binario: "/x/bin/fantasmas"}, nil
	}}
	m := Novo(nil, ac, time.Minute).ComEstado(dados.Estado{At: time.Now()}).irGaveta(indiceGaveta("settings"))
	if !strings.Contains(ansi.Strip(m.Render(110, 30)), "atualizar app fantasmas") {
		t.Fatal("falta o item nas ações de settings")
	}
	n, cmd := m.tecla("U")
	m = n.(Model)
	if cmd != nil || !m.armadoAtu || chamou != 0 {
		t.Fatalf("primeiro U devia só armar")
	}
	n, _ = m.tecla("j")
	if n.(Model).armadoAtu {
		t.Fatal("outra tecla devia desarmar")
	}
	n, _ = m.tecla("U")
	m = n.(Model)
	if !m.atualizando {
		t.Fatal("segundo U devia rodar")
	}
	n, _ = m.Update(atualizouMsg{Resultado: dados.Resultado{Novos: 2, Trocou: true, Binario: "/x/bin/fantasmas"}})
	if n.(Model).Reiniciar() != "/x/bin/fantasmas" {
		t.Fatal("devia pedir reinício com o binário novo")
	}
	n, _ = m.Update(atualizouMsg{err: errors.New("mudanças locais, não atualizei")})
	if got := n.(Model).toast; !strings.Contains(got, "mudanças locais") {
		t.Fatalf("toast = %q", got)
	}
}
