package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

const qDetalhe = `{
 "crew": "c-pl-17724", "key": "PL-17724", "area": "Edição ampliada",
 "items": [
  {"id": "plano", "kind": "gate", "title": "Aprova o plano?", "peso": "alto",
   "context": ["Contexto da decisão, com acento e travessão — para ver a quebra de linha em larguras pequenas sem cortar nada.", "- item da lista"],
   "explain": ["Por quê: a herança do motor precisa nascer atrás de flag por empresa, idêntica ao antigo com a flag desligada."],
   "options": [
     {"id": "A", "label": "Aprovar", "effect": "Segue a execução", "default": true},
     {"id": "B", "label": "Segurar", "effect": "Espera; o motivo vai na próxima mensagem ao fantasminha"}],
   "recommendation": {"option": "A", "text": "ignorado"},
   "coordinator": "A, plano cobre os ACs", "unlocks": "Execução",
   "diagram": "flowchart LR\n  A[\"Plano\"] --> B[\"Execução\"]", "diagram_caption": "Fluxo da task",
   "compare_title": "Lado a lado",
   "compare_head": ["Ponto", "Protótipo", "Hoje", ""],
   "compare": [{"item": "Botão salvar", "a": "azul primário no rodapé", "b": "cinza no topo da tela", "status": "diff", "tag": "difere"},
               {"item": "Título", "a": "Unidades", "b": "Unidades", "status": "ok", "tag": "bate"}],
   "plan": "docs/x/_plan.md",
   "links": [{"label": "Épico", "url": "https://jira/1"}]},
  {"id": "juntar", "kind": "gate", "title": "Libera juntar o PR #3589 no tronco?", "peso": "alto",
   "merge": {"repo": "PrologWebReact", "pr": 3589, "url": "https://github.com/o/PrologWebReact/pull/3589",
     "base": "feat/PL-17174-tronco", "spark": "aprovado · 2 comentários triados, 0 pendentes", "checks": "verdes (lint, test, build)"},
   "context": ["Tronco trazido às 14:02, sem conflito."],
   "options": [
     {"id": "liberar", "label": "Liberar merge", "effect": "O fantasminha junta no tronco e manda done.", "default": true},
     {"id": "segurar", "label": "Segurar", "effect": "O fantasminha espera; o motivo vai na próxima mensagem."}],
   "recommendation": "Liberar: Spark aprovado e checks verdes.", "unlocks": "Merge no tronco e fechamento da task."},
  {"id": "env", "kind": "action", "title": "Copia o .env do WS para o worktree?",
   "commands": ["cp /home/gustavo/Documents/PROLOG/prolog-claude/PrologWebService/.env.local /home/gustavo/Documents/PROLOG/prolog-claude/.worktrees/pl-17724/PrologWebService/"]}
 ]}`

const rDetalhe = `{"crew": "c-pl-5", "key": "PL-5", "area": "Herança", "status": "done",
 "summary": ["PR aberto na dev com os testes do motor verdes.", "Sem pendência de review."],
 "links": [{"label": "PR #12", "url": "https://github.com/o/r/pull/12"}],
 "report_path": ".delegate/pl-5/report.md"}`

type respondido struct{ chave, resposta string }

type cena struct {
	dir   string
	resps []respondido
	urls  []string
	abas  []string
	ent   dados.Entrega
	err   error
}

func novaCena(t *testing.T) *cena {
	t.Helper()
	dir := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(dir, "q"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, "r"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, "docs", "x"), 0o755))
	must(os.WriteFile(filepath.Join(dir, "q", "questions.json"), []byte(qDetalhe), 0o644))
	must(os.WriteFile(filepath.Join(dir, "r", "report.json"), []byte(rDetalhe), 0o644))
	must(os.WriteFile(filepath.Join(dir, "docs", "x", "_plan.md"), []byte("# Plano\n\n- passo 1\n- passo 2\n"), 0o644))
	return &cena{dir: dir, ent: dados.Entrega{Crew: "c-pl-17724", Entregue: true}}
}

func (c *cena) chave(id string) string { return filepath.Join(c.dir, "q", "questions.json") + "#" + id }

func (c *cena) estado(ids ...string) dados.Estado {
	e := dados.Estado{At: time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC), Raiz: c.dir,
		Recursos: dados.Recursos{GH: true, Herdr: true, Delegate: true, Painel: "/p.py", Autor: "eu"}}
	for _, id := range ids {
		l := dados.Linha{Chave: c.chave(id), Rotulo: "c-pl-17724 · " + id, Detalhe: id}
		if id == "juntar" {
			l.Merge = `{"repo":"PrologWebReact","pr":3589,"url":"https://github.com/o/PrologWebReact/pull/3589"}`
		}
		e.Esperando = append(e.Esperando, l)
	}
	return e
}

func (c *cena) modelo(e dados.Estado) Model {
	a := Acoes{
		AbrirURL:   func(u string) error { c.urls = append(c.urls, u); return nil },
		FocarAba:   func(s string) error { c.abas = append(c.abas, s); return nil },
		LerArquivo: os.ReadFile,
		Responder: func(k, r string) (dados.Entrega, error) {
			c.resps = append(c.resps, respondido{k, r})
			en := c.ent
			en.Resposta = r
			return en, c.err
		},
	}
	m := Novo(func(context.Context) dados.Estado { return e }, a, time.Second)
	m = m.ComEstado(e)
	m.largura, m.altura = 96, 40
	return m
}

// passo envia a tecla e, se houver comando, roda e entrega a mensagem resultante.
func passo(t *testing.T, m Model, k string) Model {
	t.Helper()
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "pgdown":
		msg = tea.KeyMsg{Type: tea.KeyPgDown}
	case "pgup":
		msg = tea.KeyMsg{Type: tea.KeyPgUp}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	n, cmd := m.Update(msg)
	m = n.(Model)
	if cmd != nil {
		if r, ok := cmd().(respondeuMsg); ok {
			n, _ = m.Update(r)
			m = n.(Model)
		}
	}
	return m
}

func abrir(t *testing.T, c *cena, id string) Model {
	t.Helper()
	m := c.modelo(c.estado(id))
	m = passo(t, m, "enter")
	if m.aberto == nil {
		t.Fatalf("detalhe de %s não abriu; status=%q", id, m.status)
	}
	return m
}

func abrirRelatorio(t *testing.T, c *cena) Model {
	t.Helper()
	e := c.estado()
	e.Tasks = []dados.Linha{{Chave: filepath.Join(c.dir, "r"), Rotulo: "c-pl-5", Detalhe: "done"}}
	m := c.modelo(e)
	m = passo(t, m, "enter")
	if m.aberto == nil || m.aberto.Tipo != "relatorio" {
		t.Fatalf("relatório não abriu; status=%q", m.status)
	}
	return m
}

func TestEnterAbreDetalheComOpcoes(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	out := render(m, 96, 60)
	for _, s := range []string{"Aprova o plano?", "Aprovar", "Segue a execução", "padrão", "c-pl-17724", "A, plano cobre os ACs", "Execução"} {
		if !strings.Contains(out, s) {
			t.Errorf("sem %q:\n%s", s, out)
		}
	}
}

func TestOpcaoComumResponde(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	passo(t, m, "enter")
	if len(c.resps) != 1 || c.resps[0] != (respondido{c.chave("plano"), "Aprovar"}) {
		t.Fatalf("resps=%v", c.resps)
	}
}

func TestOpcaoLivreAbreTexto(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	if len(c.resps) != 0 || m.rascunho == nil {
		t.Fatalf("resps=%v rascunho=%v", c.resps, m.rascunho != nil)
	}
	if !strings.Contains(render(m, 96, 40), "Segurar:") {
		t.Fatal("campo de texto não aparece")
	}
}

func TestLivreSoEnviaComTexto(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	m = passo(t, m, "enter")
	if len(c.resps) != 0 || m.status == "" {
		t.Fatalf("vazio não deveria enviar: resps=%v status=%q", c.resps, m.status)
	}
	m = passo(t, m, "x")
	passo(t, m, "enter")
	if len(c.resps) != 1 || c.resps[0].resposta != "Segurar: x" {
		t.Fatalf("resps=%v", c.resps)
	}
}

func TestEscFechaTextoDepoisDetalhe(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	m = passo(t, m, "esc")
	if m.rascunho != nil || m.aberto == nil {
		t.Fatal("esc deveria fechar só o texto")
	}
	m = passo(t, m, "esc")
	if m.aberto != nil {
		t.Fatal("esc deveria fechar o detalhe")
	}
}

func TestRespostaFechaEStatus(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "enter")
	if m.aberto != nil {
		t.Fatal("detalhe deveria fechar")
	}
	if m.status != "entregue a c-pl-17724: Aprovar" {
		t.Fatalf("status=%q", m.status)
	}
	if len(m.estado.Esperando) != 0 {
		t.Fatalf("pergunta deveria sair do Esperando: %v", m.estado.Esperando)
	}
}

func TestNaoEntregueMostraMotivo(t *testing.T) {
	c := novaCena(t)
	c.ent = dados.Entrega{Crew: "c-pl-17724", Motivo: "herdr falhou (sem sessão)"}
	m := abrir(t, c, "plano")
	m = passo(t, m, "enter")
	if m.status != "gravado; não entregue a c-pl-17724: herdr falhou (sem sessão)" {
		t.Fatalf("status=%q", m.status)
	}
}

func TestErroMantemAberto(t *testing.T) {
	c := novaCena(t)
	c.err = errors.New("não está pendente")
	m := abrir(t, c, "plano")
	m = passo(t, m, "enter")
	if m.aberto == nil || !strings.Contains(m.status, "não está pendente") {
		t.Fatalf("aberto=%v status=%q", m.aberto != nil, m.status)
	}
}

func TestGateJuntarCartao(t *testing.T) {
	c := novaCena(t)
	e := c.estado("juntar")
	e.Travados = []dados.Travado{{Repo: "PrologWebReact", Number: 3589, Motivo: "aprovado · mergear", URL: "https://github.com/o/PrologWebReact/pull/3589"}}
	m := c.modelo(e)
	m = passo(t, m, "enter")
	out := render(m, 96, 60)
	for _, s := range []string{"PrologWebReact#3589", "feat/PL-17174-tronco", "aprovado · 2 comentários", "verdes (lint, test, build)", "aprovado · mergear", "Liberar merge", "╭"} {
		if !strings.Contains(out, s) {
			t.Errorf("sem %q:\n%s", s, out)
		}
	}
	// Liberar merge confirma em 2 passos.
	m = passo(t, m, "enter")
	m = passo(t, m, "enter")
	if len(c.resps) != 1 || c.resps[0].resposta != "Liberar merge" {
		t.Fatalf("resps=%v", c.resps)
	}
}

func TestAcaoFeitoENaoVouFazer(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "env")
	out := render(m, 96, 60)
	for _, s := range []string{"Feito", "Não vou fazer", "$ cp "} {
		if !strings.Contains(out, s) {
			t.Errorf("sem %q:\n%s", s, out)
		}
	}
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	m = passo(t, m, "enter")
	if m.rascunho == nil {
		t.Fatal("Não vou fazer deveria pedir texto")
	}
}

func TestRelatorioResponde(t *testing.T) {
	c := novaCena(t)
	m := abrirRelatorio(t, c)
	out := render(m, 96, 60)
	for _, s := range []string{"PR aberto na dev", "PR #12", ".delegate/pl-5/report.md", "Lido", "Lido, e tenho pedido"} {
		if !strings.Contains(out, s) {
			t.Errorf("sem %q:\n%s", s, out)
		}
	}
	passo(t, m, "enter")
	if len(c.resps) != 1 || c.resps[0] != (respondido{"c-pl-5#report", "Lido"}) {
		t.Fatalf("resps=%v", c.resps)
	}
	m = abrirRelatorio(t, c)
	m = passo(t, m, "down")
	m = passo(t, m, "enter")
	m = passo(t, m, "y")
	passo(t, m, "enter")
	if len(c.resps) != 2 || c.resps[1] != (respondido{"c-pl-5#report", "Lido, e tenho pedido: y"}) {
		t.Fatalf("resps=%v", c.resps)
	}
}

func TestDiagramaRecolhidoExpande(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	out := render(m, 96, 80)
	if !strings.Contains(out, "d para ver") || strings.Contains(out, "--> B") {
		t.Fatalf("diagrama deveria estar recolhido:\n%s", out)
	}
	m = passo(t, m, "d")
	out = render(m, 96, 80)
	if !strings.Contains(out, "--> B") || !strings.Contains(out, "Fluxo da task") {
		t.Fatalf("diagrama não expandiu:\n%s", out)
	}
}

func TestComparacaoTabelaEEmpilhada(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	m.diagrama = true
	largo := render(m, 96, 80)
	cab := false
	for _, l := range strings.Split(largo, "\n") {
		if strings.Contains(l, "Protótipo") && strings.Contains(l, "Hoje") {
			cab = true
		}
	}
	if !cab || !strings.Contains(largo, "difere") {
		t.Fatalf("96 col sem tabela:\n%s", largo)
	}
	estreito := render(m, 56, 80)
	if !strings.Contains(estreito, "Botão salvar") || !strings.Contains(estreito, "difere") || !strings.Contains(estreito, "Protótipo:") {
		t.Fatalf("56 col sem lista empilhada:\n%s", estreito)
	}
}

func TestPreviaDoPlano(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	if out := render(m, 96, 80); !strings.Contains(out, "passo 2") {
		t.Fatalf("sem prévia:\n%s", out)
	}
	c := novaCena(t)
	os.Remove(filepath.Join(c.dir, "docs", "x", "_plan.md"))
	m = abrir(t, c, "plano")
	if out := render(m, 96, 80); !strings.Contains(out, "(arquivo não encontrado nesta máquina)") {
		t.Fatalf("sem aviso de ausente:\n%s", out)
	}
}

func TestRolagemMantemOpcoes(t *testing.T) {
	m := abrir(t, novaCena(t), "plano")
	antes := render(m, 56, 24)
	m = passo(t, m, "pgdown")
	depois := render(m, 56, 24)
	if antes == depois {
		t.Fatal("pgdown não rolou")
	}
	if !strings.Contains(depois, "Aprovar") || !strings.Contains(antes, "Aprovar") {
		t.Fatal("opções deveriam ficar fixas embaixo")
	}
	m = passo(t, m, "pgup")
	if render(m, 56, 24) != antes {
		t.Fatal("pgup não voltou ao topo")
	}
}

func TestDetalheCabe(t *testing.T) {
	c := novaCena(t)
	casos := map[string]func() Model{
		"pergunta":  func() Model { m := abrir(t, c, "plano"); m.diagrama = true; return m },
		"juntar":    func() Model { return abrir(t, c, "juntar") },
		"acao":      func() Model { return abrir(t, c, "env") },
		"relatorio": func() Model { return abrirRelatorio(t, c) },
		"livre": func() Model {
			m := abrir(t, c, "plano")
			m = passo(t, m, "down")
			return passo(t, m, "enter")
		},
	}
	for nome, f := range casos {
		for _, w := range []int{46, 56, 96, 140} {
			for _, h := range []int{12, 20, 40, 200} {
				m := f()
				for paginas := 0; paginas < 3; paginas++ {
					out := render(m, w, h)
					ls := strings.Split(out, "\n")
					if len(ls) != h {
						t.Fatalf("%s %dx%d: %d linhas", nome, w, h, len(ls))
					}
					for i, l := range ls {
						if lw := lipgloss.Width(l); lw > w {
							t.Fatalf("%s %dx%d linha %d larga (%d): %q", nome, w, h, i, lw, l)
						}
					}
					topo := len(soTopo(m, w))
					if h >= 40 && strings.Contains(ansi.Strip(strings.Join(ls[topo:h-1], "\n")), "…") {
						t.Fatalf("%s %dx%d: algo cortado com …:\n%s", nome, w, h, out)
					}
					m.largura, m.altura = w, h
					m = passo(t, m, "pgdown")
				}
			}
		}
	}
}

// TestPrints grava quadros coloridos para o relatório; só roda com FANTASMAS_PRINTS=1.
func TestPrints(t *testing.T) {
	if os.Getenv("FANTASMAS_PRINTS") != "1" {
		t.Skip("FANTASMAS_PRINTS=1 para gerar os prints")
	}
	destino := "/home/gustavo/Documents/PROLOG/prolog-claude/.delegate/2026-10-16-fantasmas-tui"
	antes := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(antes)
	c := novaCena(t)
	for _, w := range []int{56, 96} {
		m := abrir(t, c, "plano")
		if err := os.WriteFile(filepath.Join(destino, "t7-pergunta-"+strconv.Itoa(w)+".txt"), []byte(m.Render(w, 44)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		e := c.estado("juntar")
		e.Travados = []dados.Travado{{Repo: "PrologWebReact", Number: 3589, Motivo: "aprovado · mergear"}}
		mj := c.modelo(e)
		mj = passo(t, mj, "enter")
		if err := os.WriteFile(filepath.Join(destino, "t7-juntar-"+strconv.Itoa(w)+".txt"), []byte(mj.Render(w, 44)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if w == 96 {
			mj = passo(t, mj, "enter")
			if err := os.WriteFile(filepath.Join(destino, "usab-juntar-armado-96.ansi"), []byte(mj.Render(96, 44)+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func render(m Model, w, h int) string { return ansi.Strip(m.Render(w, h)) }
