package ui

import (
	"time"

	"fantasmas/internal/dados"
)

// EstadoDemoVerde é a demo com o CI de todos os PRs verde, para a olhada do "tudo verde".
func EstadoDemoVerde(agora time.Time) dados.Estado {
	e := EstadoDemo(agora)
	prs := append([]dados.Travado(nil), e.Travados...)
	for i := range prs {
		prs[i].CI, prs[i].Falhos = "SUCCESS", nil
	}
	e.Travados = prs
	return e
}

// EstadoDemo é um estado determinístico para a olhada renderizada (--demo): os dados do rascunho aprovado.
func EstadoDemo(agora time.Time) dados.Estado {
	e := estadoDemoBase(agora)
	e.PipesLigado = true
	e.Sidequests = runsDemo(agora)
	dia := agora.Format("2006-01-02")
	e.Proximo = map[string]string{
		dia + "-VT-19597-homolog": "não lança: aguarda o login do humano (gate)",
		dia + "-VT-19848-pr":      "não lança: aguarda corrigir (after)",
	}
	missoesDemo(&e, agora)
	filaDemo(&e)
	e.Mergeados = mergeadosDemo(agora)
	e.Licoes = licoesDemo(agora)
	return e
}

// filaDemo: duas decisões na fila, uma com regra do vision e outra fora dele, ambas com recomendação.
func filaDemo(e *dados.Estado) {
	e.Esperando = append(e.Esperando, dados.Linha{Chave: "t-working/q2", Rotulo: "c-vt-1 · testes", Detalhe: "Trocar o Testcontainers por banco em memória nos testes de listas?"})
	e.Recs = map[string]dados.Recomendacao{
		"t-espera/q1":  {Opcao: "aprovar", Texto: "Aprovar o plano", Regra: "migração de dados só roda com plano aprovado"},
		"t-working/q2": {Opcao: "manter", Texto: "Manter o Testcontainers"},
	}
}

// mergeadosDemo: o que foi mergeado hoje nas duas missões e fora delas.
func mergeadosDemo(agora time.Time) []dados.Mergeado {
	return []dados.Mergeado{
		{Repo: "d21", Number: 4198, Title: "[VT-19597] Placar ao vivo", Branch: "feature/placar-VT-19597", Commits: 6, Em: agora.Add(-40 * time.Minute)},
		{Repo: "d21", Number: 4190, Title: "[VT-19598] Tabela de pontos", Branch: "feature/pontos-VT-19598", Commits: 3, Em: agora.Add(-3 * time.Hour)},
		{Repo: "mtgNoPapel", Number: 1299, Title: "[VT-19826] Preço por carta", Branch: "feature/preco-VT-19826", Commits: 4, Em: agora.Add(-2 * time.Hour)},
		{Repo: "vitrine", Number: 3099, Title: "Ficha do fantasmas", Branch: "docs/ficha-fantasmas", Commits: 1, Em: agora.Add(-1 * time.Hour)},
	}
}

// licoesDemo: erros que se repetiram nos fantasminhas e o que os corrigiu.
func licoesDemo(agora time.Time) []dados.Licao {
	return []dados.Licao{
		{Erro: "lint: import não usado", Vezes: 4, Correcao: "rodar o lint antes do commit", Ultima: agora.Add(-25 * time.Minute)},
		{Erro: "teste de integração sem banco", Vezes: 3, Correcao: "subir o Testcontainers antes da suíte", Ultima: agora.Add(-2 * time.Hour)},
		{Erro: "snapshot de tela desatualizado", Vezes: 1, Correcao: "regravar só com a mudança aprovada", Ultima: agora.Add(-5 * time.Hour)},
	}
}

const (
	demoCustos   = "d21-placar"
	demoEconomia = "mtg-preco-deck"
)

// missoesDemo: o coordenador principal e as duas missões do rascunho (uma espera você, outra trabalhando).
func missoesDemo(e *dados.Estado, agora time.Time) {
	um, zero := 1, 0
	e.MissoesLigado = true
	e.Principal = dados.Principal{Nome: "Central", Workspace: "w1", Gengar: "coord-1030-0900"}
	e.Missoes = []dados.Missao{
		{Slug: demoEconomia, Nome: "mtg-preco-deck", Workspace: "w16", Origem: "produto", OrigemRef: "pedido do CS",
			Cards: []string{"VT-19826"}, Estado: "aberta", Gengar: "g-mtg-preco-deck", Sidequests: []int{11, 14},
			Objetivo: "Mostrar quanto o deck custa com preços de hoje.", Prototipo: "https://claude.ai/design/mtg-preco-deck",
			Contagem:        dados.Contagem{Trabalhando: 2, PipesAtivos: 1, PrsVermelhos: &zero},
			UltimaAtividade: agora.Add(-2 * time.Minute),
			Decisoes: []dados.Decisao{{At: agora.Add(-3 * time.Minute).Format(time.RFC3339), Via: "regra",
				Slug: agora.Format("2006-01-02") + "-VT-19826", Item: "plano", Motivo: "plano de 1 repo, sem migration"}}},
		{Slug: demoCustos, Nome: "d21-placar", Workspace: "w15", Origem: "issue", OrigemRef: "GH-123",
			Cards: []string{"VT-19597", "VT-19598"}, Estado: "aberta", Gengar: "g-d21-placar", Sidequests: []int{13},
			Objetivo:        "Somar o placar de cada jogador da mesa.",
			Contagem:        dados.Contagem{Trabalhando: 1, Parados: 1, Pendentes: 1, PipesAtivos: 1, PrsVermelhos: &um},
			UltimaAtividade: agora.Add(-1 * time.Minute)},
	}
	e.Coordenadores = map[string]dados.LinhaViva{
		"coord-1030-0900": {Nome: "coord-1030-0900", Aba: "w1:t1", Etapa: "coordenar", Estado: dados.Trabalhando, CtxPct: 41,
			Agora: "revisando o PR do React", Hora: agora.Add(-1 * time.Minute)},
		"g-d21-placar": {Nome: "g-d21-placar", Aba: "w15:t1", Etapa: "coordenar", Estado: dados.EsperaVoce, CtxPct: 38,
			Agora: "1 sua: login para homologar (c-vt-2)", Hora: agora.Add(-1 * time.Minute)},
		"g-mtg-preco-deck": {Nome: "g-mtg-preco-deck", Aba: "w16:t1", Etapa: "coordenar", Estado: dados.Trabalhando, CtxPct: 17,
			Agora: "Bash: herdr agent wait h-vt-5 --until blocked", Hora: agora.Add(-2 * time.Minute)},
	}
	ws := map[string][2]string{
		"c-vt-1": {"w15", demoCustos}, "c-vt-2": {"w15", demoCustos}, "c-vt-3": {"w15", demoCustos},
		"h-vt-5": {"w16", demoEconomia}, "c-vt-4": {"w16", demoEconomia},
	}
	for i, l := range e.Linhas {
		e.Linhas[i].Workspace, e.Linhas[i].Missao = ws[l.Nome][0], ws[l.Nome][1]
	}
	for i, t := range e.Tasks {
		e.Tasks[i].Missao = ws[t.Rotulo][1]
	}
	for i := range e.Esperando {
		e.Esperando[i].Missao = demoCustos
	}
	porRun := map[int]string{13: demoCustos, 11: demoEconomia, 14: demoEconomia}
	for i, r := range e.Sidequests {
		e.Sidequests[i].Missao, e.Sidequests[i].Solta = porRun[r.ID], porRun[r.ID] == ""
	}
	e.Mural = append([]dados.Aviso{{At: agora.Add(-4 * time.Minute), Autor: "g-d21-placar",
		Texto: "homologação pronta, esperando o login do humano", Missao: demoCustos}}, e.Mural...)
}

// runsDemo: as quatro sidequests do rascunho (opção B), uma delas solta.
func runsDemo(agora time.Time) []dados.Sidequest {
	dia := agora.Format("2006-01-02")
	impl := func(slug, status, agent, tab string) dados.Etapa {
		return dados.Etapa{ID: "implementar", Role: "implementer", Skill: "writing-plans", Status: status,
			Slug: dia + "-" + slug, Agent: agent, Tab: tab,
			Gates: []dados.Gate{{Item: "plano", By: "human"}, {Item: "juntar", By: "human"}}}
	}
	homol := func(slug, status string) dados.Etapa {
		return dados.Etapa{ID: "homologar", Role: "homolog", Skill: "homolog", Status: status, Slug: dia + "-" + slug + "-homolog",
			After: []string{"implementar.done"}, Gates: []dados.Gate{{Item: "login", By: "human"}}}
	}
	etapa := func(id, slug, status, agent string, after ...string) dados.Etapa {
		return dados.Etapa{ID: id, Role: "implementer", Skill: "debug-for-fix", Status: status, Slug: dia + "-" + slug,
			Agent: agent, After: after}
	}
	login := homol("VT-19597", "needs_input")
	login.Reason = "gate:login"
	pr := etapa("PR", "VT-19826-pr", "failed", "", "implementar.done")
	pr.Reason = "falhou"
	feita := etapa("pr", "VT-19830-pr", "done", "c-vt-19830", "corrigir.done")
	feita.Report = []byte(`{"status":"done","links":["https://github.com/gus-aguiar/vitrine/pull/3120"]}`)
	return []dados.Sidequest{
		{ID: 13, Pipeline: "implementar-homologar", Title: "Placar por jogador", SlugBase: "VT-19597", Status: "running",
			Coordinator: "g-d21-placar", Inputs: map[string]string{"card": "VT-19597", "repos": "WS React", "homolog_alvo": "vps"},
			IdadeMin: 42, EsperaVoce: 1, Prontidao: "pronto",
			Etapas: []dados.Etapa{impl("VT-19597", "done", "c-vt-19597", "w15:t3"), login,
				{ID: "juntar", Role: "coordenador", Status: "pending", Slug: dia + "-VT-19597-juntar", After: []string{"homologar.done"},
					Gates: []dados.Gate{{Item: "juntar", By: "human"}}}}},
		{ID: 11, Pipeline: "bugfix", Title: "Valor da carta arredondado", SlugBase: "VT-19830", Status: "done",
			Coordinator: "g-mtg-preco-deck", Inputs: map[string]string{"card": "VT-19830", "repos": "WS"}, IdadeMin: 26 * 60,
			Etapas: []dados.Etapa{etapa("debug", "VT-19830", "done", "c-vt-19830"),
				etapa("corrigir", "VT-19830-corrigir", "done", "c-vt-19830", "debug.done"), feita}},
		{ID: 14, Pipeline: "implementar-homologar", Title: "Preço do deck atualizado", SlugBase: "VT-19826", Status: "failed",
			Coordinator: "g-mtg-preco-deck", Prontidao: "aguardando prontidão",
			Inputs: map[string]string{"card": "VT-19826", "repos": "React"}, IdadeMin: 70,
			Etapas: []dados.Etapa{impl("VT-19826", "done", "c-vt-19826", ""), pr, homol("VT-19826", "pending")}},
		{ID: 12, Pipeline: "bugfix", Title: "Item duplicado na lista", SlugBase: "VT-19848", Status: "running",
			Coordinator: "coord-1030-0900", Inputs: map[string]string{"card": "VT-19848", "repos": "WS"}, IdadeMin: 95,
			Etapas: []dados.Etapa{etapa("debug", "VT-19848", "done", "c-vt-19848"),
				etapa("corrigir", "VT-19848-corrigir", "running", "c-vt-19848", "debug.done"),
				etapa("pr", "VT-19848-pr", "pending", "", "corrigir.done")}},
	}
}

func estadoDemoBase(agora time.Time) dados.Estado {
	return dados.Estado{
		At:   agora,
		Raiz: "/demo/workspace",
		Mural: []dados.Aviso{
			{At: agora.Add(-8 * time.Minute), Autor: "coord-1030-0900", Texto: "vou mexer no delegate-spawn até 12h"},
			{At: agora.Add(-95 * time.Minute), Autor: "gustavo", Texto: "VPS fora até 15h, sem homologação por enquanto"},
		},
		Vivos:    map[string]bool{"c-vt-1": true, "c-vt-2": true, "c-vt-3": true, "c-vt-4": true},
		Recursos: dados.Recursos{GH: true, Herdr: true, Delegate: true, Painel: "/demo/workspace/scripts/delegate/painel.py", Autor: "demo"},
		Tasks: []dados.Linha{
			{Chave: "t-working", Rotulo: "c-vt-1", Detalhe: "working · VT-1001 · implementando", Extra: "aba-1"},
			{Chave: "t-espera", Rotulo: "c-vt-2", Detalhe: "esperando você · VT-1002", Extra: "aba-2"},
			{Chave: "t-blocked", Rotulo: "c-vt-3", Detalhe: "blocked · VT-1003 · CI falhou", Extra: "aba-3"},
			{Chave: "t-done", Rotulo: "c-vt-4", Detalhe: "done · VT-1004 · relatório pronto", Extra: "aba-4"},
		},
		Linhas: []dados.LinhaViva{
			{Nome: "c-vt-2", Pasta: "/demo/workspace/.delegate/c-vt-2", Aba: "aba-2", Etapa: "implementar", Agora: "1 sua(s): plano", Estado: dados.EsperaVoce, CtxPct: 41, Hora: agora.Add(-3 * time.Minute)},
			{Nome: "c-vt-1", Pasta: "/demo/workspace/.delegate/c-vt-1", Aba: "aba-1", Etapa: "implementar", Agora: "rodando a suíte de integração do módulo de listas com Testcontainers e conferindo a migração de índices — ação válida até o fim da conferência", Estado: dados.Trabalhando, CtxPct: 55, Hora: agora.Add(-1 * time.Minute)},
			{Nome: "h-vt-5", Pasta: "/demo/workspace/.delegate/h-vt-5", Aba: "aba-5", Etapa: "homologar", Agora: "Playwright: navegando em /listas/compartilhar", Estado: dados.Trabalhando, CtxPct: 86, Hora: agora.Add(-2 * time.Minute)},
			{Nome: "c-vt-3", Pasta: "/demo/workspace/.delegate/c-vt-3", Aba: "aba-3", Etapa: "implementar", Agora: "parado com tarefa há 14 min", Estado: dados.Parado, CtxPct: -1, Hora: agora.Add(-14 * time.Minute)},
			{Nome: "c-vt-4", Pasta: "/demo/workspace/.delegate/c-vt-4", Aba: "aba-4", Etapa: "implementar", Agora: "entregou relatório", Estado: dados.Terminou, CtxPct: 100, Hora: agora.Add(-20 * time.Minute)},
		},
		Caidos: []dados.Caido{
			{Nome: "c-vt-7", Pane: "w1:p7", SessionID: "9f3c2a71-5b0e-4d8a-a1c4-2e7b60d1f5aa", Pasta: "/demo/workspace/.delegate/c-vt-7"},
		},
		Esperando: []dados.Linha{
			{Chave: "t-espera/q1", Rotulo: "c-vt-2 · plano", Detalhe: "Aprovar o plano de migração da tela de listas?"},
		},
		Travados: []dados.Travado{
			{Repo: "vitrine", Number: 3101, Motivo: "CI vermelho", Title: "[VT-1003] Filtro de projetos", URL: "https://github.com/gus-aguiar/vitrine/pull/3101", Branch: "feature/filtro-projetos-VT-1003", CI: "FAILURE", Falhos: []string{"lint", "vitest"}},
			{Repo: "d21", Number: 4202, Motivo: "aguardando revisão", Title: "[VT-1001] Endpoint de placar", URL: "https://github.com/gus-aguiar/d21/pull/4202", Branch: "feature/placar-VT-1001", CI: "PENDING"},
			{Repo: "mtgNoPapel", Number: 1303, Motivo: "aprovado · mergear", Title: "[VT-1005] Ajuste de tema", URL: "https://github.com/gus-aguiar/mtgNoPapel/pull/1303", Branch: "feature/tema-VT-1005", CI: "SUCCESS"},
			{Repo: "lista-livre", Number: 504, Motivo: "aprovado · mergear", Title: "[VT-1006] Spec de checklist", URL: "https://github.com/gus-aguiar/lista-livre/pull/504", Branch: "feature/e2e-VT-1006", CI: "SUCCESS"},
			{Repo: "fantasmas", Number: 605, Motivo: "aprovado · mergear", Title: "[VT-1007] Token de borda", URL: "https://github.com/gus-aguiar/fantasmas/pull/605", Branch: "feature/borda-VT-1007", CI: "SUCCESS"},
			{Repo: "vitrine-kit", Number: 706, Motivo: "aguardando revisão", Title: "[VT-1008] Evento de notificação", URL: "https://github.com/gus-aguiar/vitrine-kit/pull/706", Branch: "feature/evento-VT-1008", CI: "SUCCESS"},
			{Repo: "gus-setup", Number: 807, Motivo: "aguardando revisão", Title: "[VT-1009] Exportação sem CI", URL: "https://github.com/gus-aguiar/gus-setup/pull/807", Branch: "feature/export-VT-1009"},
		},
		Servicos: []dados.Linha{
			{Chave: "vitrine", Rotulo: "no ar", Detalhe: ":3000"},
			{Chave: "d21", Rotulo: "parado", Detalhe: ""},
		},
		Maquina: []dados.Linha{
			{Chave: "trava", Rotulo: "trava de testes pesados", Detalhe: "livre", Extra: "ok"},
		},
	}
}

// nomeCoordenador é como a interface chama o coordenador e a figura padrão; o --demo troca por um nome neutro.
var nomeCoordenador = "gengar"

// ModoDemo troca os nomes próprios da interface por neutros, para gravações públicas. Só o --demo chama.
func ModoDemo() { nomeCoordenador = "coordenador" }

// rotuloFigura mostra a figura padrão com o nome da interface; as demais ficam como estão.
func rotuloFigura(v string) string {
	if v == FiguraPadrao && nomeCoordenador != FiguraPadrao {
		return "fantasma"
	}
	return v
}
