package ui

import (
	"fmt"
	"strings"
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

// gavetaAcoes é a seção "ações" da gaveta configurações: reconectar MCPs e atualizar app, com enter 2×.
func gavetaAcoes(c *corpo, e dados.Estado, ec estadoComandos) {
	c.branco()
	c.secao("ações", "")
	semTecla := func(s string) string {
		if i := strings.LastIndex(s, " · "); i >= 0 {
			return s[:i]
		}
		return s
	}
	g, cor, est, det := botaoMCP(e.MCP)
	det = semTecla(det)
	if ec.armado == cmdMCP {
		est, cor = "enter de novo confirma", T.EstadoAviso()
	}
	c.linhaLista(alvo{comando: cmdMCP}, g, cor, "Reconectar MCPs", est, cor, det)
	g, cor, est = "○", T.Dim, ""
	if ec.atualizando {
		g, cor, est = "◐", T.Andamento, "atualizando…"
	} else if ec.armado == cmdAtualizar || ec.armadoAtu {
		est, cor = "enter de novo confirma", T.EstadoAviso()
	}
	c.linhaLista(alvo{comando: cmdAtualizar}, g, cor, "Atualizar app", est, cor, "git pull, build e reabre")
}

// dicaItem é o atalho que aparece à direita só na linha selecionada, conforme o tipo do item.
func dicaItem(a alvo) string {
	var ps []string
	switch {
	case a.dica != "":
		return a.dica
	case a.comando != "":
		return "enter 2× executa"
	case a.retomar != "":
		return "R 2× retoma"
	case a.antigas:
		return "enter expande"
	case a.missao != "":
		return "enter detalhes · g aba"
	case a.arquivar != "":
		return ""
	case a.pergunta != "":
		ps = append(ps, "enter responder")
		if a.aba != "" {
			ps = append(ps, "g ir à aba")
		}
		if a.url != "" {
			ps = append(ps, "o abrir")
		}
	case a.aba != "":
		ps = append(ps, "g ir à aba")
	}
	return strings.Join(ps, " · ")
}
