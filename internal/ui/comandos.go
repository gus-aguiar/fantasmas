package ui

import (
	"fmt"
	"time"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
)

// alertasMaquina são as leituras fora do verde: máquina com aviso/erro e fila de pesados ocupada.
func alertasMaquina(e dados.Estado) []string {
	var out []string
	for _, l := range e.Maquina {
		switch {
		case l.Chave == "trava":
			if l.Extra != "ok" && l.Extra != "" {
				out = append(out, "pesados: "+l.Detalhe)
			}
		case l.Extra == "erro" || l.Extra == "aviso":
			out = append(out, l.Rotulo)
		}
	}
	return out
}

// gavetaMaquinaToda junta máquina, ambiente e worktrees; tudo verde vira uma linha que abre no enter.
func gavetaMaquinaToda(c *corpo, e dados.Estado, aberta bool) {
	alertas := alertasMaquina(e)
	if len(alertas) == 0 && !aberta && len(e.Maquina) > 0 {
		no := 0
		for _, s := range e.Servicos {
			if s.Rotulo != "parado" {
				no++
			}
		}
		resumo := fmt.Sprintf("tudo verde · %d serviço(s) no ar · %d worktree(s)", no, len(e.Worktrees))
		c.item(alvo{alterna: "maquina"}, func(bg lipgloss.Color) string {
			return st(bg, T.Ok).Render("● ") + st(bg, T.Texto).Render("máquina") + st(bg, T.Dim).Render("  "+resumo+" · enter abre")
		})
		return
	}
	if aberta && len(alertas) == 0 {
		c.item(alvo{alterna: "maquina"}, func(bg lipgloss.Color) string {
			return st(bg, T.Acao).Render("▾ ") + st(bg, T.Dim).Render("tudo verde · enter recolhe")
		})
		c.branco()
	}
	gavetaMaquina(c, e)
	c.branco()
	gavetaAmbiente(c, e)
	c.branco()
	gavetaWorktrees(c, e)
}

// estadoComandos é o que a gaveta comandos precisa do Model.
type estadoComandos struct {
	disparos                          map[string]dados.Disparo
	armado                            string
	disparando                        string
	agora                             time.Time
	pedindoDD, atualizando, armadoAtu bool
}

const (
	cmdDaydreaming = "daydreaming"
	cmdMCP         = "mcp"
	cmdAtualizar   = "atualizar"
)

// rodandoHa é o "rodando há N min" de um disparo.
func rodandoHa(d dados.Disparo, agora time.Time) string {
	m := int(agora.Sub(d.Inicio).Minutes())
	if m < 1 {
		return "rodando há menos de 1 min"
	}
	return fmt.Sprintf("rodando há %d min", m)
}

// gavetaComandos lista os comandos (enter 2× dispara) e, no fim, os ajustes.
func gavetaComandos(c *corpo, e dados.Estado, ops OpsAjustes, ec estadoComandos) {
	c.secao("comandos", "enter 2× dispara · g vai à aba")
	confirma := func(id string) bool { return ec.armado == id }
	for _, k := range dados.ComandosHerdr {
		k := k
		g, cor, est, det := "›", T.Texto, "", k.Descricao+" · aba em "+k.Workspace
		a := alvo{comando: k.ID}
		if d, ok := ec.disparos[k.ID]; ok {
			g, cor, est, det = "●", T.Andamento, "no ar", rodandoHa(d, ec.agora)+" · "+d.Nome
			a.aba = d.Aba
		}
		switch {
		case ec.disparando == k.ID:
			g, cor, est = "◐", T.Andamento, "abrindo…"
		case confirma(k.ID):
			cor, est = T.Espera, "de novo confirma"
		}
		c.linhaLista(a, g, cor, k.Rotulo, est, cor, det)
	}
	g, cor, est, det := botaoDaydreaming(e.Daydreamer, ec.pedindoDD)
	if confirma(cmdDaydreaming) {
		cor, est = T.Espera, "de novo confirma"
	}
	c.linhaLista(alvo{comando: cmdDaydreaming}, g, cor, "DayDreaming", est, cor, det)
	g, cor, est, det = botaoMCP(e.MCP)
	if confirma(cmdMCP) {
		cor, est = T.Espera, "de novo confirma"
	}
	c.linhaLista(alvo{comando: cmdMCP}, g, cor, "reconectar MCPs", est, cor, det)
	g0, cor0, est0 := "⟳", T.Texto, ""
	switch {
	case ec.atualizando:
		g0, cor0, est0 = "◐", T.Andamento, "atualizando…"
	case ec.armadoAtu || confirma(cmdAtualizar):
		cor0, est0 = T.Espera, "de novo confirma"
	}
	c.linhaLista(alvo{comando: cmdAtualizar}, g0, cor0, "atualizar app", est0, cor0, "git pull, build e reabre · U 2×")
	c.branco()
	gavetaAjustes(c, ops)
}
