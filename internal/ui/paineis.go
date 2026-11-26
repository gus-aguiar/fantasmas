package ui

import (
	"fmt"
	"strings"

	"fantasmas/internal/dados"

	"github.com/charmbracelet/lipgloss"
)

// maxLicoes é quantas lições a gaveta missões mostra; as menos repetidas ficam de fora.
const maxLicoes = 5

// paineisDoDia fecha a gaveta missões com o que foi feito hoje (PRs mergeados por missão) e as lições
// (erros que os fantasminhas corrigiram, por repetição). Seção sem dado não aparece.
func paineisDoDia(c *corpo, e dados.Estado) {
	feito := dados.FeitoHoje(e, e.At)
	if len(feito) > 0 {
		prs := 0
		for _, f := range feito {
			prs += len(f.PRs)
		}
		c.branco()
		c.secaoN("feito hoje", prs, "PRs mergeados, por missão")
		w := c.largura - 2
		for _, f := range feito {
			f := f
			nome, cor, glifo := f.Missao, corMissao(f.Missao), "●"
			if nome == "" {
				nome, cor, glifo = "sem missão", T.Dim, "○"
			}
			var quais []string
			for _, p := range f.PRs {
				quais = append(quais, fmt.Sprintf("%s #%d %s", semPrefixo(p.Repo), p.Number, limpaTitulo(p.Title)))
			}
			if !c.passa(append([]string{nome}, quais...)...) {
				continue
			}
			titulo := fmt.Sprintf("%-16s %s · %s", nome, plural(len(f.PRs), "PR", "PRs"), plural(f.Commits, "commit", "commits"))
			c.item(alvo{url: f.PRs[0].URL, dica: "o abrir o mais recente"}, func(bg lipgloss.Color) string {
				return colunas(bg, w, glifo, cor, titulo, strings.Join(quais, " · "))
			})
		}
	}
	if len(e.Licoes) > 0 {
		c.branco()
		c.secaoN("lições", len(e.Licoes), "Erros corrigidos nos fantasminhas, do mais repetido")
		w := c.largura - 2
		for i, l := range e.Licoes {
			if i >= maxLicoes {
				break
			}
			l := l
			if !c.passa(l.Erro, l.Correcao) {
				continue
			}
			cor := T.Dim
			if l.Vezes > 1 {
				cor = T.Espera
			}
			correcao := l.Correcao
			if correcao == "" {
				correcao = "sem correção registrada"
			}
			c.texto(func(bg lipgloss.Color) string {
				return colunas(bg, w, fmt.Sprintf("%2d×", l.Vezes), cor, l.Erro, "→ "+correcao)
			})
		}
	}
}

func plural(n int, um, varios string) string {
	if n == 1 {
		return "1 " + um
	}
	return fmt.Sprintf("%d %s", n, varios)
}

// limpaTitulo tira o "[CARD] " do começo do título do PR, que a missão já diz.
func limpaTitulo(t string) string {
	if strings.HasPrefix(t, "[") {
		if i := strings.Index(t, "] "); i > 0 {
			return t[i+2:]
		}
	}
	return t
}
