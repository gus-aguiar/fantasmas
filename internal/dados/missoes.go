package dados

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Principal é o workspace do GENGAR coordenador geral (por padrão "Genguinha").
type Principal struct{ Nome, Workspace, Gengar string }

type Contagem struct {
	Trabalhando, Parados, Pendentes, PipesAtivos int
	// PrsVermelhos fica nil no ddb; a TUI preenche com os PRs que já tem.
	PrsVermelhos *int
}

// Missao é um workspace do Herdr (fora o principal) com seu gengar e seus cards.
type Missao struct {
	Slug, Nome, Workspace, Origem, OrigemRef, Estado, Gengar, Objetivo, Prototipo string
	Cards                                                                         []string
	Contagem                                                                      Contagem
	UltimaAtividade                                                               time.Time
}

type principalDdb struct {
	Nome      string `json:"nome"`
	Workspace string `json:"workspace"`
	Gengar    string `json:"gengar"`
}

type missaoDdb struct {
	Slug      string   `json:"slug"`
	Nome      string   `json:"nome"`
	Workspace string   `json:"workspace"`
	Origem    string   `json:"origem"`
	OrigemRef string   `json:"origem_ref"`
	Cards     []string `json:"cards"`
	Estado    string   `json:"estado"`
	Gengar    string   `json:"gengar"`
	Objetivo  string   `json:"objetivo"`
	Prototipo string   `json:"prototipo"`
	Contagem  struct {
		Trabalhando  int  `json:"trabalhando"`
		Parados      int  `json:"parados"`
		Pendentes    int  `json:"pendentes"`
		PipesAtivos  int  `json:"pipes_ativos"`
		PrsVermelhos *int `json:"prs_vermelhos"`
	} `json:"contagem"`
	Ultima string `json:"ultima_atividade"`
}

func (x missaoDdb) missao() Missao {
	c := x.Contagem
	return Missao{Slug: x.Slug, Nome: x.Nome, Workspace: x.Workspace, Origem: x.Origem, OrigemRef: x.OrigemRef,
		Estado: x.Estado, Gengar: x.Gengar, Objetivo: x.Objetivo, Prototipo: x.Prototipo, Cards: x.Cards,
		Contagem:        Contagem{c.Trabalhando, c.Parados, c.Pendentes, c.PipesAtivos, c.PrsVermelhos},
		UltimaAtividade: quandoISO(x.Ultima)}
}

func quandoISO(s string) time.Time {
	for _, f := range []string{time.RFC3339, "2006-01-02T15:04:05-0700", "2006-01-02T15:04Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// PreencherPRsVermelhos conta, por missão, os PRs de CI vermelho cujo título cita um card dela.
func PreencherPRsVermelhos(ms []Missao, prs []Travado) {
	for i := range ms {
		n := 0
		for _, p := range prs {
			if p.CI == "FAILURE" && CitaCard(p.Title, ms[i].Cards) {
				n++
			}
		}
		ms[i].Contagem.PrsVermelhos = &n
	}
}

// CitaCard diz se o título menciona algum dos cards (palavra inteira).
func CitaCard(titulo string, cards []string) bool {
	for _, c := range cards {
		if c != "" && regexp.MustCompile(`\b`+regexp.QuoteMeta(c)+`\b`).MatchString(titulo) {
			return true
		}
	}
	return false
}

// WorkspaceDe lê workspace= do meta da pasta; sem ele, o prefixo da aba do Herdr (w15:t2 → w15).
func WorkspaceDe(pasta, aba string) string {
	if b, err := os.ReadFile(filepath.Join(pasta, "meta")); err == nil {
		for _, l := range strings.Split(string(b), "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(l), "workspace="); ok && v != "" {
				return v
			}
		}
	}
	if ws, _, ok := strings.Cut(aba, ":"); ok {
		return ws
	}
	return ""
}

// AtribuirMissoes liga cada linha viva à missão: pela task da mesma pasta, senão pelo workspace.
func AtribuirMissoes(e *Estado) {
	porPasta := map[string]string{}
	for _, t := range e.Tasks {
		if t.Missao != "" {
			porPasta[t.Chave] = t.Missao
		}
	}
	porWs := map[string]string{}
	for _, m := range e.Missoes {
		if m.Workspace != "" {
			porWs[m.Workspace] = m.Slug
		}
	}
	for i, l := range e.Linhas {
		if s := porPasta[l.Pasta]; s != "" {
			e.Linhas[i].Missao = s
		} else if s := porWs[l.Workspace]; s != "" {
			e.Linhas[i].Missao = s
		}
	}
}

// LinhaCoordenador é a linha viva de um gengar (principal ou de missão), que não tem meta.
func LinhaCoordenador(nome string, ag Agente, raiz, claudeDir string, agora time.Time) LinhaViva {
	ses, at := atividadeDoAgente(nome, ag, SessoesClaude(filepath.Join(claudeDir, "sessions")), raiz, claudeDir)
	l := MontarLinhaViva(nome, ag, ses, at, nil, false, false, agora)
	l.Etapa = "gengar"
	l.Workspace = WorkspaceDe("", ag.Aba)
	return l
}

func atividadeDoAgente(nome string, ag Agente, sessoes map[string]SessaoClaude, raiz, claudeDir string) (*SessaoClaude, Atividade) {
	ses, id := sessaoDe(sessoes, nome, ag.Sessao)
	if ag.Sessao != "" {
		id = ag.Sessao
	}
	var at Atividade
	if id != "" {
		cwd := raiz
		if ses != nil && ses.Cwd != "" {
			cwd = ses.Cwd
		}
		at = AtividadeDe(LerCauda(CaminhoTranscript(claudeDir, cwd, id), maxCauda))
	}
	return ses, at
}

// Coordenadores lê a linha viva do GENGAR principal e do gengar de cada missão vivos no Herdr.
func Coordenadores(e Estado, raiz, claudeDir string, agora time.Time) map[string]LinhaViva {
	out := map[string]LinhaViva{}
	nomes := []string{e.Principal.Gengar}
	for _, m := range e.Missoes {
		nomes = append(nomes, m.Gengar)
	}
	for _, n := range nomes {
		if ag, ok := e.Agentes[n]; ok && n != "" {
			out[n] = LinhaCoordenador(n, ag, raiz, claudeDir, agora)
		}
	}
	return out
}
