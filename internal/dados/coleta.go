package dados

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Coletor struct {
	Raiz string
	R    Runner

	autor      string
	recursos   Recursos
	descoberto bool

	tickMu     sync.Mutex
	tickEm     time.Time
	tickUltimo map[string]string
}

const (
	timeoutColetor = 60 * time.Second
	timeoutPRs     = 25 * time.Second
	timeoutProbe   = 10 * time.Second
	timeoutTick    = 20 * time.Second
	validadeTick   = 2 * time.Minute
)

func (c *Coletor) rodar(ctx context.Context, limite time.Duration, argv ...string) (string, bool) {
	cctx, cancel := context.WithTimeout(ctx, limite)
	defer cancel()
	out, err := c.R.Run(cctx, argv...)
	return out, err == nil
}

func (c *Coletor) descobrir(ctx context.Context) {
	if c.descoberto {
		return
	}
	var gh, herdr bool
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, gh = c.rodar(ctx, timeoutProbe, "gh", "--version") }()
	go func() { defer wg.Done(); _, herdr = c.rodar(ctx, timeoutProbe, "herdr", "--version") }()
	wg.Wait()
	autor := ""
	if gh {
		if out, ok := c.rodar(ctx, timeoutProbe, "gh", "api", "user", "--jq", ".login"); ok {
			autor = strings.TrimSpace(out)
		}
	}
	painel := filepath.Join(c.Raiz, "shared", "skills", "delegate-painel", "painel.py")
	if st, err := os.Stat(painel); err != nil || st.IsDir() {
		painel = ""
	}
	st, err := os.Stat(filepath.Join(c.Raiz, ".delegate"))
	c.autor = autor
	c.recursos = Recursos{GH: gh, Herdr: herdr, Delegate: err == nil && st.IsDir(), Painel: painel, Autor: autor}
	c.descoberto = true
}

func (c *Coletor) Coletar(ctx context.Context) Estado {
	c.descobrir(ctx)
	scripts := filepath.Join(c.Raiz, "shared", "mods", "control-panel", "scripts")
	py := filepath.Join(scripts, "delegate.py")
	cl := filepath.Join(scripts, "coleta.sh")

	var tasks, esperando, opcoes, local, worktrees, maquina, prs, agentes string
	var ddb *ParteDdb
	var tickPendente bool
	var proximo map[string]string
	c.recursos.FalhaDdb = ""
	if fonteDdbLigada() {
		if p, err := c.lerDdb(ctx); err == nil {
			ddb = &p
			tickPendente = true
		} else {
			c.recursos.FalhaDdb = err.Error()
		}
	}
	var wg sync.WaitGroup
	roda := func(dst *string, limite time.Duration, argv ...string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if out, ok := c.rodar(ctx, limite, argv...); ok {
				*dst = out
			}
		}()
	}
	if tickPendente {
		wg.Add(1)
		go func() { defer wg.Done(); proximo = c.lerTick(ctx) }()
	}
	if ddb == nil {
		roda(&tasks, timeoutColetor, "python3", py, "tasks")
		roda(&esperando, timeoutColetor, "python3", py, "esperando")
		roda(&opcoes, timeoutColetor, "python3", py, "opcoes-todas")
	}
	roda(&local, timeoutColetor, "bash", filepath.Join(scripts, "local.sh"))
	roda(&worktrees, timeoutColetor, "bash", cl, "worktrees")
	roda(&maquina, timeoutColetor, "bash", filepath.Join(scripts, "maquina.sh"))
	if c.recursos.Herdr {
		roda(&agentes, timeoutProbe, "herdr", "agent", "list")
	}
	if c.recursos.GH && c.autor != "" {
		roda(&prs, timeoutPRs, "gh", "api", "graphql", "-f", "query="+QueryDe(c.autor))
	}
	var dd Daydreamer
	if c.recursos.Herdr {
		wg.Add(1)
		go func() { defer wg.Done(); dd = LerDaydreamer(ctx, c.R, c.Raiz) }()
	} else {
		dd = Daydreamer{Nome: NomeDaydreamer(c.Raiz), Ultima: UltimaDaydreamer(c.Raiz)}
	}
	wg.Wait()

	var travados []Travado
	if prs != "" {
		travados = PRsDe([]byte(prs))
	}
	var vivos map[string]bool
	var ags map[string]Agente
	if agentes != "" {
		vivos = VivosDe([]byte(agentes))
		ags = AgentesDe([]byte(agentes))
	}
	e := Estado{
		Tasks:      TSV(tasks),
		Esperando:  TSV(esperando),
		Opcoes:     TSV(opcoes),
		Worktrees:  TSV(worktrees),
		Maquina:    TSV(maquina),
		Servicos:   TSV(local),
		Travados:   travados,
		Recursos:   c.recursos,
		At:         time.Now(),
		Raiz:       c.Raiz,
		Vivos:      vivos,
		Vistos:     CarregarVistos(),
		Daydreamer: dd,
		MCP:        LerProgressoMCP(c.Raiz, time.Now()),
		Agentes:    ags,
		Mural:      MuralDe(filepath.Join(c.Raiz, ".delegate", "_mural.jsonl"), time.Now(), 3),
	}
	if ddb != nil {
		e.Tasks, e.Esperando, e.Opcoes, e.Mural = ddb.Tasks, ddb.Esperando, ddb.Opcoes, ddb.Mural
		e.Runs, e.PipesLigado, e.Proximo = ddb.Runs, true, proximo
		if os.Getenv("FANTASMAS_MURAL") == "off" {
			e.Mural = nil
		}
		e.MissoesLigado, e.Principal, e.Missoes = true, ddb.Principal, ddb.Missoes
		PreencherPRsVermelhos(e.Missoes, e.Travados)
	}
	e.Caidos = CaidosDe(c.Raiz, vivos, maxCaidos)
	e.Linhas = LinhasVivas(e, c.Raiz, DirClaude(), time.Now())
	AtribuirMissoes(&e)
	if e.MissoesLigado {
		e.Coordenadores = Coordenadores(e, c.Raiz, DirClaude(), time.Now())
	}
	return e
}

func FaltaDe(r Recursos) []string {
	var f []string
	if !r.GH {
		f = append(f, "gh: gavetas PRs e CI vazias (instale e rode gh auth login)")
	}
	if r.GH && r.Autor == "" {
		f = append(f, `login do GitHub: gaveta PRs vazia (gh auth login ou a opção "autor")`)
	}
	if !r.Herdr {
		f = append(f, `herdr: sem "ir →" para a aba do fantasminha`)
	}
	if !r.Delegate {
		f = append(f, ".delegate/: nenhuma task do /delegate ainda")
	}
	if r.Delegate && r.Painel == "" {
		f = append(f, "painel.py: resposta só grava; a entrega fica com o coordenador")
	}
	return append(f, faltaDdb(r)...)
}
