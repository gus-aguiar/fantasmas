package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type runnerPipes struct {
	mu    sync.Mutex
	calls []string
}

func (r *runnerPipes) Run(_ context.Context, argv ...string) (string, error) {
	k := strings.Join(argv, " ")
	r.mu.Lock()
	r.calls = append(r.calls, k)
	r.mu.Unlock()
	switch {
	case strings.HasSuffix(k, "sidequest list --json"):
		return `[{"name":"implementar","arquivo":"/s/implementar.yaml","version":1,"inputs":["card","repos"],"stages":["implementar"],"erros":[]},
{"name":"bugfix","arquivo":"/s/bugfix.yaml","version":1,"inputs":["card"],"stages":["debug","corrigir"],"erros":[]}]`, nil
	case strings.Contains(k, " sidequest new "):
		return "7 2026-10-30-PL-9\n", nil
	}
	return "", nil
}

func estadoPipes() dados.Estado {
	e := EstadoDemo(time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC))
	e.Proximo = map[string]string{"2026-10-30-PL-19597-homolog": "lançaria (homolog, skill homolog)"}
	return e
}

func comPipes(e dados.Estado, r *runnerPipes) Model {
	a := Acoes{Pipes: &dados.Pipes{R: r, Raiz: "/r", Script: "/r/ddb.py"}}
	m := Novo(func(context.Context) dados.Estado { return e }, a, time.Second).ComEstado(e).ComGaveta("missoes")
	m.largura, m.altura = 140, 60
	return m
}

// selecionaSidequest põe a seleção na linha da sidequest id.
func selecionaSidequest(t *testing.T, m Model, id int) Model {
	t.Helper()
	for i, a := range m.alvos() {
		if a.pipe == id && a.vetoDe == "" && a.missao == "" {
			m.sel = i
			return m
		}
	}
	t.Fatalf("sidequest %d fora da gaveta", id)
	return m
}

// modeloPipes abre a gaveta missões da demo com a sidequest 13 selecionada.
func modeloPipes(t *testing.T, r *runnerPipes) Model {
	return selecionaSidequest(t, comPipes(estadoPipes(), r), 13)
}

func teclaMsg(m Model, msg tea.KeyMsg) Model {
	n, _ := m.Update(msg)
	return n.(Model)
}

func digitaPipe(m Model, s string) Model {
	return teclaMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)})
}

var t0SQ = time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)

// estadoSQ é o rascunho aprovado (opção B): duas missões, quatro sidequests, uma solta.
func estadoSQ() dados.Estado {
	et := func(id, status, reason, slug, agent string) dados.Etapa {
		return dados.Etapa{ID: id, Status: status, Reason: reason, Slug: slug, Agent: agent}
	}
	return dados.Estado{
		At: t0SQ, MissoesLigado: true, PipesLigado: true,
		Missoes: []dados.Missao{
			{Slug: "multas-custos", Nome: "multas-custos", Workspace: "w20", Origem: "PDY", Gengar: "g-multas-custos",
				Cards: []string{"PL-19597", "PL-19598"}, Sidequests: []int{1}, Contagem: dados.Contagem{Pendentes: 1}},
			{Slug: "economia-multas", Nome: "economia-multas", Workspace: "w21", Origem: "produto", Gengar: "g-economia",
				Cards: []string{"PL-19826"}, Sidequests: []int{2, 3}, Contagem: dados.Contagem{Trabalhando: 2},
				Decisoes: []dados.Decisao{
					{Via: "regra", Slug: "s-3-impl", Item: "plano", Motivo: "plano curto"},
					{Via: "humano", Slug: "s-2-corrigir", Item: "plano"},
				}},
		},
		Sidequests: []dados.Sidequest{
			{ID: 1, Pipeline: "implementar-homologar", Status: "running", EsperaVoce: 1, Missao: "multas-custos",
				Etapas: []dados.Etapa{et("implementar", "done", "", "s-1-impl", ""), et("homologar", "needs_input", "gate:login", "s-1-hom", ""), et("juntar", "pending", "", "s-1-j", "")}},
			{ID: 2, Pipeline: "bugfix", Status: "done", Missao: "economia-multas",
				Etapas: []dados.Etapa{et("debug", "done", "", "s-2-debug", ""), et("corrigir", "done", "", "s-2-corrigir", ""), et("pr", "done", "", "s-2-pr", "")}},
			{ID: 3, Pipeline: "implementar-homologar", Status: "failed", Missao: "economia-multas",
				Etapas: []dados.Etapa{et("implementar", "done", "", "s-3-impl", ""), et("PR", "failed", "falhou", "s-3-pr", ""), et("homologar", "pending", "", "s-3-hom", "")}},
			{ID: 4, Pipeline: "bugfix", SlugBase: "PL-19848", Status: "running", Solta: true,
				Etapas: []dados.Etapa{et("debug", "done", "", "s-4-debug", ""), et("corrigir", "running", "", "s-4-corrigir", "c-pl-19848"), et("pr", "pending", "", "s-4-pr", "")}},
		},
	}
}

func telaSQ(m Model, w int) string { return ansi.Strip(m.Render(w, 60)) }

func TestSemGavetaPipes(t *testing.T) {
	ids := strings.Join(IDsGavetas(), ",")
	if ids != "agora,missoes,prs,maquina,configuracoes" {
		t.Fatalf("ordem=%s", ids)
	}
	if GavetaValida("pipes") {
		t.Fatal("a gaveta pipes não existe mais")
	}
	for i, g := range Gavetas {
		if g.Tecla != fmt.Sprint(i) {
			t.Fatalf("%s com tecla %s na posição %d", g.ID, g.Tecla, i)
		}
	}
}

func TestCartaMostraSidequestComBlocos(t *testing.T) {
	p := telaSQ(comPipes(estadoSQ(), &runnerPipes{}), 140)
	for _, s := range []string{"implementar-homologar ■■□ homologar · login", "bugfix ■■■ concluída", "implementar-homologar ■■□ PR · falhou"} {
		if !strings.Contains(p, s) {
			t.Errorf("falta %q:\n%s", s, p)
		}
	}
	if strings.Contains(strings.ToLower(p), "pipe") {
		t.Errorf("texto de UI com pipe:\n%s", p)
	}
}

func TestBlocosEtapasUmPorEtapa(t *testing.T) {
	es := estadoSQ().Sidequests[3].Etapas
	if b := ansi.Strip(blocosEtapas(es)); b != "■■□" {
		t.Fatalf("blocos=%q", b)
	}
}

func TestSoltasNoFimDaGaveta(t *testing.T) {
	e := estadoSQ()
	p := telaSQ(comPipes(e, &runnerPipes{}), 140)
	i := strings.Index(p, "SIDEQUESTS SOLTAS")
	if i < 0 || !strings.Contains(p[i:], "bugfix · PL-19848 ■■□ corrigir · c-pl-19848") {
		t.Fatalf("soltas fora do fim:\n%s", p)
	}
	if strings.Index(p, "economia-multas") > i {
		t.Fatal("soltas antes das missões")
	}
	e.Sidequests = e.Sidequests[:3]
	if p := telaSQ(comPipes(e, &runnerPipes{}), 140); strings.Contains(p, "SIDEQUESTS SOLTAS") {
		t.Fatalf("seção sem soltas:\n%s", p)
	}
	// workspace sem missão: a sidequest sai em soltas, nunca some
	e = estadoSQ()
	e.Missoes = nil
	p = telaSQ(comPipes(e, &runnerPipes{}), 140)
	if !strings.Contains(p, "SIDEQUESTS SOLTAS") || strings.Count(p, "■") < 4 {
		t.Fatalf("sem missão, as sidequests sumiram:\n%s", p)
	}
}

func TestRodapeContagem(t *testing.T) {
	p := telaSQ(comPipes(estadoSQ(), &runnerPipes{}), 140)
	if !strings.Contains(p, "2 missões · 4 sidequests (1 solta)") || !strings.Contains(p, "m nova missão") {
		t.Fatalf("rodapé:\n%s", p)
	}
}

func TestLarguras56e96(t *testing.T) {
	longo := "implementar-homologar-com-um-nome-bem-comprido-de-verdade-que-nao-cabe-nem-em-noventa-e-seis-colunas"
	for _, w := range []int{56, 96} {
		e := estadoSQ()
		e.Sidequests[0].Pipeline = longo
		m := comPipes(e, &runnerPipes{})
		m.largura = w
		out := m.Render(w, 60)
		for i, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
		achou := false
		for _, l := range strings.Split(ansi.Strip(out), "\n") {
			if strings.Contains(l, "implementar-homologar-com") {
				achou = true
				if !strings.Contains(l, "…") || !strings.Contains(l, "■■□") || !strings.Contains(l, "▲ ") {
					t.Fatalf("w=%d: nome não cortado ou blocos/glifo cortados: %q", w, l)
				}
			}
		}
		if !achou {
			t.Fatalf("w=%d: sem a linha da sidequest:\n%s", w, ansi.Strip(out))
		}
	}
}

func TestEnterAbreDetalheDaSidequest(t *testing.T) {
	m := selecionaSidequest(t, comPipes(estadoSQ(), &runnerPipes{}), 3)
	m = tecla(m, "enter")
	if m.pipe != 3 {
		t.Fatalf("enter não abriu o detalhe: pipe=%d", m.pipe)
	}
	if p := ansi.Strip(m.Render(140, 40)); !strings.Contains(p, "sidequest 3 · implementar-homologar") {
		t.Fatalf("detalhe:\n%s", p)
	}
}

func preencheImplementar(m Model) Model {
	m = tecla(m, "enter")
	m = digitaPipe(m, "PL-9")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "WS")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	return tecla(m, "enter")
}

func TestTeclaNLancaNaMissaoEmFoco(t *testing.T) {
	r := &runnerPipes{}
	m := comPipes(estadoSQ(), r)
	m.missao = "economia-multas"
	m = preencheImplementar(tecla(m, "n"))
	quer := "python3 /r/ddb.py --raiz /r sidequest new implementar --input card=PL-9 --input repos=WS --workspace w21"
	if r.calls[len(r.calls)-1] != quer {
		t.Fatalf("com foco: calls=%v", r.calls)
	}
	r = &runnerPipes{}
	m = preencheImplementar(tecla(comPipes(estadoSQ(), r), "n"))
	if c := r.calls[len(r.calls)-1]; strings.Contains(c, "--workspace") || !strings.Contains(c, " sidequest new implementar ") {
		t.Fatalf("sem foco: %s", c)
	}
}

func TestTeclaNComFocoEmOutraGaveta(t *testing.T) {
	r := &runnerPipes{}
	m := comPipes(estadoSQ(), r)
	m.missao = "multas-custos"
	m = m.irGaveta(indiceGaveta("agora"))
	m = preencheImplementar(tecla(m, "n"))
	if c := r.calls[len(r.calls)-1]; !strings.HasSuffix(c, "--workspace w20") {
		t.Fatalf("n na gaveta crew com foco: %s", c)
	}
	r = &runnerPipes{}
	m = comPipes(estadoSQ(), r).irGaveta(indiceGaveta("agora"))
	if m = tecla(m, "n"); m.form != nil {
		t.Fatal("sem foco, n fora da gaveta missões não lança")
	}
}

func TestTeclaMAbreMissao(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(comPipes(estadoSQ(), r), "m")
	if m.form == nil || m.form.tipo != "missao" {
		t.Fatal("m não abriu o formulário da missão")
	}
	for _, v := range []string{"multas-novas", "Multas novas", "produto", "pedido do CS", "PL-1,PL-2"} {
		m = digitaPipe(m, v)
		m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	}
	m = tecla(m, "enter")
	if len(r.calls) != 0 || m.form == nil || !strings.Contains(m.status, "enter de novo") {
		t.Fatalf("1º enter deveria só armar: calls=%v status=%q", r.calls, m.status)
	}
	m = tecla(m, "enter")
	quer := "bash /r/scripts/delegate/missao-abrir.sh multas-novas --nome Multas novas --origem produto --origem-ref pedido do CS --cards PL-1,PL-2"
	if len(r.calls) == 0 || r.calls[0] != quer {
		t.Fatalf("calls=%v", r.calls)
	}
	if m.form != nil {
		t.Fatal("formulário não fechou")
	}
	m = tecla(comPipes(estadoSQ(), r), "m")
	m = digitaPipe(m, "x")
	for range 5 {
		m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	}
	m = tecla(m, "enter")
	if m.form == nil || !strings.Contains(m.status, "falta") {
		t.Fatalf("aceitou campos vazios: status=%q", m.status)
	}
}

func TestDecisoesComVetar(t *testing.T) {
	r := &runnerPipes{}
	m := comPipes(estadoSQ(), r)
	p := telaSQ(m, 140)
	if !strings.Contains(p, "regra aprovou plano") || !strings.Contains(p, "v vetar") || !strings.Contains(p, "você aprovou plano") {
		t.Fatalf("decisões:\n%s", p)
	}
	if strings.Count(p, "v vetar") != 1 {
		t.Fatalf("só a decisão da regra é vetável:\n%s", p)
	}
	idx := -1
	for i, a := range m.alvos() {
		if a.vetoDe != "" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("sem alvo de veto")
	}
	m.sel = idx
	m = tecla(m, "v")
	if m.form == nil || m.form.tipo != "veto" {
		t.Fatal("v não abriu o veto")
	}
	m = digitaPipe(m, "refaz o plano")
	m = tecla(m, "enter")
	if len(r.calls) != 0 || m.form == nil {
		t.Fatalf("1º enter deveria só armar: calls=%v", r.calls)
	}
	m = tecla(m, "enter")
	if quer := "python3 /r/ddb.py --raiz /r stage 3:implementar veto -- refaz o plano"; len(r.calls) == 0 || r.calls[len(r.calls)-1] != quer {
		t.Fatalf("calls=%v", r.calls)
	}
}

func TestDecisaoSemEtapaNaoVeta(t *testing.T) {
	e := estadoSQ()
	e.Missoes[1].Decisoes = []dados.Decisao{{Via: "regra", Slug: "A", Item: "pista rápida", Motivo: "pista rápida"}}
	r := &runnerPipes{}
	m := comPipes(e, r)
	if p := telaSQ(m, 140); !strings.Contains(p, "regra aprovou pista rápida") || strings.Contains(p, "vetar") {
		t.Fatalf("decisão sem etapa não pode oferecer veto:\n%s", p)
	}
	idx := -1
	for i, a := range m.alvos() {
		if a.decisao {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("sem alvo da decisão")
	}
	m.sel = idx
	m = tecla(m, "v")
	if m.form != nil || m.toast != "nada para vetar nesta decisão" || len(r.calls) != 0 {
		t.Fatalf("form=%v toast=%q calls=%v", m.form, m.toast, r.calls)
	}
}

func TestDecisaoUsaRunIDEStageID(t *testing.T) {
	e := estadoSQ()
	run := 2
	e.Missoes[1].Decisoes = []dados.Decisao{{Via: "regra", Slug: "fora", Item: "plano", RunID: &run, StageID: "corrigir"}}
	m := comPipes(e, &runnerPipes{})
	if !strings.Contains(telaSQ(m, 140), "v vetar") {
		t.Fatal("decisão com run_id/stage_id devia ser vetável")
	}
	for i, a := range m.alvos() {
		if a.vetoDe != "" {
			if a.vetoID != 2 || a.vetoEtapa != "corrigir" {
				t.Fatalf("alvo=%+v", a)
			}
			m.sel = i
		}
	}
	if m = tecla(m, "v"); m.form == nil || m.form.tipo != "veto" {
		t.Fatal("v não abriu o veto")
	}
}

func TestPipesSemChaveMostraMensagem(t *testing.T) {
	m := tecla(modelo(fixture(), &registro{}).ComGaveta("missoes"), "n")
	if !strings.Contains(m.toast, "sidequests: ligue FANTASMAS_FONTE=ddb") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestPipesDetalheEmTodasLarguras(t *testing.T) {
	for _, w := range larguras {
		m := modeloPipes(t, &runnerPipes{})
		m = tecla(m, "enter")
		if m.pipe == 0 {
			t.Fatalf("w=%d: enter não abriu o detalhe", w)
		}
		m.largura = w
		out := m.Render(w, 40)
		for i, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
		p := ansi.Strip(out)
		if w >= 70 {
			for _, s := range []string{"implementar", "homologar", "após implementar", "gate:login", "lançaria"} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d: falta %q:\n%s", w, s, p)
				}
			}
		}
		m = tecla(m, "esc")
		if m.pipe != 0 {
			t.Fatalf("w=%d: esc não fechou", w)
		}
	}
}

func TestPipesASTPedemConfirmacao(t *testing.T) {
	for _, c := range []struct{ k, acao string }{{"A", "advance"}, {"S", "skip"}, {"T", "retry"}} {
		r := &runnerPipes{}
		m := tecla(modeloPipes(t, r), "enter")
		m = tecla(m, c.k)
		if len(r.calls) != 0 {
			t.Fatalf("%s: rodou sem confirmar: %v", c.k, r.calls)
		}
		if !strings.Contains(m.status, c.k+" de novo") {
			t.Fatalf("%s: status=%q", c.k, m.status)
		}
		m = tecla(m, "j")
		m = tecla(m, c.k)
		if len(r.calls) != 0 {
			t.Fatalf("%s: mover a seleção deveria desarmar: %v", c.k, r.calls)
		}
		m = tecla(m, c.k)
		quer := "python3 /r/ddb.py --raiz /r stage " + fmt.Sprint(m.runAberto().ID) + ":juntar " + c.acao
		if len(r.calls) == 0 || r.calls[0] != quer {
			t.Fatalf("%s: calls=%v quer %q", c.k, r.calls, quer)
		}
	}
}

func TestPipesFormularioMontaComando(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(modeloPipes(t, r), "n")
	if m.form == nil {
		t.Fatal("n não abriu o formulário")
	}
	if p := ansi.Strip(m.Render(140, 42)); !strings.Contains(p, "implementar") || !strings.Contains(p, "bugfix") || !strings.Contains(p, "lançar sidequest") {
		t.Fatalf("lista de sidequests ausente:\n%s", p)
	}
	m = tecla(m, "enter")
	m = digitaPipe(m, "PL-9")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "WS React")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyTab})
	m = digitaPipe(m, "Filtro")
	m = tecla(m, "enter")
	quer := "python3 /r/ddb.py --raiz /r sidequest new implementar --input card=PL-9 --input repos=WS React --title Filtro"
	if r.calls[len(r.calls)-1] != quer {
		t.Fatalf("calls=%v", r.calls)
	}
	if m.form != nil {
		t.Fatal("formulário não fechou depois de lançar")
	}
	if !strings.Contains(m.toast, "sidequest 7") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestPipesFormularioRecusaEntradaVazia(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(tecla(modeloPipes(t, r), "n"), "enter")
	m = tecla(m, "enter")
	m = tecla(m, "enter")
	m = tecla(m, "enter")
	for _, c := range r.calls {
		if strings.Contains(c, "sidequest new") {
			t.Fatalf("lançou com entrada vazia: %v", r.calls)
		}
	}
	if m.form == nil {
		t.Fatal("formulário fechou")
	}
}

func TestPipesNMaiusculoPedeConfirmacaoEChamaGengar(t *testing.T) {
	r := &runnerPipes{}
	m := tecla(modeloPipes(t, r), "N")
	if m.form != nil || !strings.Contains(m.status, "N de novo") {
		t.Fatalf("primeiro N deveria só armar: form=%v status=%q", m.form, m.status)
	}
	m = tecla(m, "N")
	if m.form == nil {
		t.Fatal("segundo N não abriu o formulário")
	}
	m = preencheImplementar(m)
	n := len(r.calls)
	if n < 2 || !strings.Contains(r.calls[n-2], " sidequest new implementar ") {
		t.Fatalf("sidequest new ausente: %v", r.calls)
	}
	if quer := "bash /r/scripts/delegate/gengar-spawn.sh 7 --assunto PL-9"; r.calls[n-1] != quer {
		t.Fatalf("calls=%v quer %q", r.calls, quer)
	}
	if !strings.Contains(m.toast, "gengar") {
		t.Fatalf("toast=%q", m.toast)
	}
}

func TestPipesNMinusculoNaoChamaGengar(t *testing.T) {
	r := &runnerPipes{}
	preencheImplementar(tecla(modeloPipes(t, r), "n"))
	for _, c := range r.calls {
		if strings.Contains(c, "gengar-spawn") {
			t.Fatalf("n não deveria abrir gengar: %v", r.calls)
		}
	}
}

func TestPipesNMoverDesarma(t *testing.T) {
	m := tecla(modeloPipes(t, &runnerPipes{}), "N")
	m = tecla(m, "j")
	m = tecla(m, "N")
	if m.form != nil {
		t.Fatal("mover entre os dois N deveria desarmar")
	}
}

func TestPipesLinhaMostraGengarEProntidao(t *testing.T) {
	for _, w := range larguras {
		m := comPipes(estadoPipes(), &runnerPipes{})
		m.largura = w
		out := m.Render(w, 80)
		p := ansi.Strip(out)
		if w >= 46 {
			for _, s := range []string{"g-economia-multas", "aguardando prontidão", "g-multas-custos", "✓ pronto"} {
				if !strings.Contains(p, s) {
					t.Errorf("w=%d: falta %q:\n%s", w, s, p)
				}
			}
		}
		for i, l := range strings.Split(out, "\n") {
			if lipgloss.Width(l) > w {
				t.Fatalf("w=%d linha %d larga: %d", w, i, lipgloss.Width(l))
			}
		}
	}
}
