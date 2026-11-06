package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/x/ansi"
)

func TestAcoesDeSistemaNaGaveta5(t *testing.T) {
	e := dados.Estado{At: time.Now(),
		MCP: dados.ProgressoMCP{Inicio: time.Now(), Total: 4, Rodando: true, Sessoes: []dados.SessaoMCP{{Pane: "p", Nome: "g", Detalhe: "ok", Tom: "ok"}}},
	}
	m := Novo(nil, Acoes{}, time.Minute).ComEstado(e).irGaveta(indiceGaveta("configuracoes"))
	tela := ansi.Strip(m.Render(110, 40))
	for _, s := range []string{"Reconectar MCPs", "reconectando…", "1/4 sessões", "Atualizar app"} {
		if !strings.Contains(tela, s) {
			t.Errorf("falta %q:\n%s", s, tela)
		}
	}
	if p := ansi.Strip(m.ComBotoes().Render(110, 40)); strings.Contains(p, "Reconectar MCPs") || strings.Contains(p, "DayDreaming") {
		t.Errorf("paleta não devia ter ações de sistema:\n%s", p)
	}
	for m.sel = 0; m.sel < len(m.alvos()) && m.alvos()[m.sel].comando != cmdMCP; m.sel++ {
	}
	if !strings.Contains(ansi.Strip(m.Render(110, 40)), "enter 2× executa") {
		t.Error("linha selecionada sem a dica enter 2×")
	}
	n, _ := m.tecla("enter")
	m = n.(Model)
	if m.armadoCmd != cmdMCP {
		t.Fatal("primeiro enter devia armar")
	}
}

func TestAtualizarPedeSegundoU(t *testing.T) {
	chamou := 0
	ac := Acoes{Atualizar: func() (dados.Resultado, error) {
		chamou++
		return dados.Resultado{Novos: 2, Trocou: true, Binario: "/x/bin/fantasmas"}, nil
	}}
	m := Novo(nil, ac, time.Minute).ComEstado(dados.Estado{At: time.Now()})
	if !strings.Contains(ansi.Strip(m.irGaveta(indiceGaveta("configuracoes")).Render(110, 50)), "Atualizar app") {
		t.Fatal("falta o item na gaveta 5")
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
