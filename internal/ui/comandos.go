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

// estadoComandos é o que o painel de botões precisa do Model.
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
