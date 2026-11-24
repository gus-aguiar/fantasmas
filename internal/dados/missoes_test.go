package dados

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const estadoMissoes = `{"tasks":[{"chave":"/r/.delegate/a","nome":"c-a","tab":"w15:t2","marca":"working","pendentes":0,"total":0,"quando":"1min","missao":"placar"},
{"chave":"/r/.delegate/b","nome":"c-b","tab":"w1:t3","marca":"done","pendentes":0,"total":0,"quando":"1min","missao":null}],
"pendentes":[{"chave":"/r/.delegate/a/questions.json#p1","rotulo":"c-a · plano","detalhe":"?","missao":"placar"}],
"opcoes":{},
"mural":[{"at":"2026-10-30T12:50:00Z","autor":"g-placar","texto":"marco","missao":"placar"}],
"runs":[{"id":3,"pipeline":"x","title":"t","status":"running","missao":"placar","stages":[]}],
"principal":{"nome":"Genguinha","workspace":"w1","gengar":"gengar-1030-0900"},
"missoes":[{"slug":"placar","nome":"Placar por unidade","workspace":"w15","origem":"issue","origem_ref":"GH-123",
 "cards":["VT-1203","VT-1204"],"estado":"aberta","gengar":"g-placar","objetivo":"Filtrar o placar.","prototipo":null,
 "contagem":{"trabalhando":2,"parados":1,"pendentes":1,"pipes_ativos":1,"prs_vermelhos":null},
 "ultima_atividade":"2026-10-30T12:57:00Z"},
 {"slug":"listas","nome":"Listas","workspace":"w16","origem":"produto","origem_ref":"pedido","cards":[],"estado":"concluida",
 "gengar":"g-listas","objetivo":"","contagem":{"trabalhando":0,"parados":0,"pendentes":0,"pipes_ativos":0,"prs_vermelhos":null},
 "ultima_atividade":null}]}`

func TestDeDdbLeMissoes(t *testing.T) {
	agora := time.Date(2026, 10, 30, 13, 0, 0, 0, time.UTC)
	p, err := DeDdb([]byte(estadoMissoes), agora)
	if err != nil {
		t.Fatal(err)
	}
	if p.Principal.Nome != "Genguinha" || p.Principal.Workspace != "w1" || p.Principal.Gengar != "gengar-1030-0900" {
		t.Errorf("principal=%+v", p.Principal)
	}
	if len(p.Missoes) != 2 {
		t.Fatalf("missoes=%+v", p.Missoes)
	}
	m := p.Missoes[0]
	if m.Slug != "placar" || m.Nome != "Placar por unidade" || m.Workspace != "w15" || m.Origem != "issue" ||
		m.OrigemRef != "GH-123" || len(m.Cards) != 2 || m.Gengar != "g-placar" || m.Objetivo != "Filtrar o placar." {
		t.Errorf("missao=%+v", m)
	}
	if c := m.Contagem; c.Trabalhando != 2 || c.Parados != 1 || c.Pendentes != 1 || c.PipesAtivos != 1 || c.PrsVermelhos != nil {
		t.Errorf("contagem=%+v", c)
	}
	if !m.UltimaAtividade.Equal(agora.Add(-3 * time.Minute)) {
		t.Errorf("ultima=%v", m.UltimaAtividade)
	}
	if !p.Missoes[1].UltimaAtividade.IsZero() {
		t.Error("ultima nula deveria ficar zero")
	}
	if p.Tasks[0].Missao != "placar" || p.Tasks[1].Missao != "" || p.Esperando[0].Missao != "placar" {
		t.Errorf("tasks=%+v esperando=%+v", p.Tasks, p.Esperando)
	}
	if p.Sidequests[0].Missao != "placar" || len(p.Mural) != 1 || p.Mural[0].Missao != "placar" {
		t.Errorf("runs=%+v mural=%+v", p.Sidequests, p.Mural)
	}
}

func TestPRsVermelhosPorCard(t *testing.T) {
	ms := []Missao{{Slug: "placar", Cards: []string{"VT-1203"}}, {Slug: "listas", Cards: []string{"VT-9"}}}
	prs := []Travado{
		{Title: "[VT-1203] filtro", CI: "FAILURE"},
		{Title: "[VT-1203] outro", CI: "SUCCESS"},
		{Title: "[VT-90] outro", CI: "FAILURE"},
	}
	PreencherPRsVermelhos(ms, prs)
	if ms[0].Contagem.PrsVermelhos == nil || *ms[0].Contagem.PrsVermelhos != 1 {
		t.Errorf("placar=%v", ms[0].Contagem.PrsVermelhos)
	}
	if ms[1].Contagem.PrsVermelhos == nil || *ms[1].Contagem.PrsVermelhos != 0 {
		t.Errorf("listas=%v", ms[1].Contagem.PrsVermelhos)
	}
}

func TestLinhaGanhaMissaoPelaTaskOuWorkspace(t *testing.T) {
	e := Estado{
		Missoes: []Missao{{Slug: "placar", Workspace: "w15"}},
		Tasks:   []Linha{{Chave: "/r/.delegate/a", Missao: "placar"}},
		Linhas: []LinhaViva{
			{Nome: "c-a", Pasta: "/r/.delegate/a"},
			{Nome: "c-x", Pasta: "/r/.delegate/x", Workspace: "w15"},
			{Nome: "c-y", Pasta: "/r/.delegate/y", Workspace: "w1"},
		},
	}
	AtribuirMissoes(&e)
	if e.Linhas[0].Missao != "placar" || e.Linhas[1].Missao != "placar" || e.Linhas[2].Missao != "" {
		t.Errorf("linhas=%+v", e.Linhas)
	}
}

func TestWorkspaceDaLinhaVemDoMetaOuDaAba(t *testing.T) {
	raiz := t.TempDir()
	d := filepath.Join(raiz, ".delegate", "a")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "meta"), []byte("name=c-a\nworkspace=w15\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ws := WorkspaceDe(d, "w1:t2"); ws != "w15" {
		t.Errorf("meta: %q", ws)
	}
	if ws := WorkspaceDe(filepath.Join(raiz, "nada"), "w7:t2"); ws != "w7" {
		t.Errorf("aba: %q", ws)
	}
}

func TestLinhaCoordenadorLeTranscript(t *testing.T) {
	claude := t.TempDir()
	cwd := "/r"
	id := "sess-g"
	dst := CaminhoTranscript(claude, cwd, id)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("testdata/transcript-bash.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatal(err)
	}
	agora := time.Date(2026, 10, 26, 19, 16, 0, 0, time.UTC)
	l := LinhaCoordenador("g-placar", Agente{Status: "working", Sessao: id, Aba: "w15:t1"}, cwd, claude, agora)
	if l.Nome != "g-placar" || l.Etapa != "gengar" || l.Estado != Trabalhando || l.Agora == "—" || l.CtxPct < 0 || l.Aba != "w15:t1" {
		t.Errorf("linha=%+v", l)
	}
}
