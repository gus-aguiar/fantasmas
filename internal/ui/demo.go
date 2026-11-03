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
		dia + "-PL-19597-homolog": "não lança: aguarda o login do humano (gate)",
		dia + "-PL-19848-pr":      "não lança: aguarda corrigir (after)",
	}
	missoesDemo(&e, agora)
	return e
}

const (
	demoCustos   = "multas-custos"
	demoEconomia = "economia-multas"
)

// missoesDemo: o Genguinha e as duas missões do rascunho (uma espera você, outra trabalhando).
func missoesDemo(e *dados.Estado, agora time.Time) {
	um, zero := 1, 0
	e.MissoesLigado = true
	e.Principal = dados.Principal{Nome: "Genguinha", Workspace: "w1", Gengar: "gengar-1030-0900"}
	e.Missoes = []dados.Missao{
		{Slug: demoEconomia, Nome: "economia-multas", Workspace: "w16", Origem: "produto", OrigemRef: "pedido do CS",
			Cards: []string{"PL-19826"}, Estado: "aberta", Gengar: "g-economia-multas", Sidequests: []int{11, 14},
			Objetivo: "Mostrar quanto a frota economiza com multas evitadas.", Prototipo: "https://claude.ai/design/economia-multas",
			Contagem:        dados.Contagem{Trabalhando: 2, PipesAtivos: 1, PrsVermelhos: &zero},
			UltimaAtividade: agora.Add(-2 * time.Minute),
			Decisoes: []dados.Decisao{{At: agora.Add(-3 * time.Minute).Format(time.RFC3339), Via: "regra",
				Slug: agora.Format("2006-01-02") + "-PL-19826", Item: "plano", Motivo: "plano de 1 repo, sem migration"}}},
		{Slug: demoCustos, Nome: "multas-custos", Workspace: "w15", Origem: "PDY", OrigemRef: "PDY-123",
			Cards: []string{"PL-19597", "PL-19598"}, Estado: "aberta", Gengar: "g-multas-custos", Sidequests: []int{13},
			Objetivo:        "Somar o custo das multas por veículo.",
			Contagem:        dados.Contagem{Trabalhando: 1, Parados: 1, Pendentes: 1, PipesAtivos: 1, PrsVermelhos: &um},
			UltimaAtividade: agora.Add(-1 * time.Minute)},
	}
	e.Coordenadores = map[string]dados.LinhaViva{
		"gengar-1030-0900": {Nome: "gengar-1030-0900", Aba: "w1:t1", Etapa: "gengar", Estado: dados.Trabalhando, CtxPct: 41,
			Agora: "revisando o PR do React", Hora: agora.Add(-1 * time.Minute)},
		"g-multas-custos": {Nome: "g-multas-custos", Aba: "w15:t1", Etapa: "gengar", Estado: dados.EsperaVoce, CtxPct: 38,
			Agora: "1 sua: login para homologar (c-pl-2)", Hora: agora.Add(-1 * time.Minute)},
		"g-economia-multas": {Nome: "g-economia-multas", Aba: "w16:t1", Etapa: "gengar", Estado: dados.Trabalhando, CtxPct: 17,
			Agora: "Bash: herdr agent wait h-pl-5 --until blocked", Hora: agora.Add(-2 * time.Minute)},
	}
	ws := map[string][2]string{
		"c-pl-1": {"w15", demoCustos}, "c-pl-2": {"w15", demoCustos}, "c-pl-3": {"w15", demoCustos},
		"h-pl-5": {"w16", demoEconomia}, "c-pl-4": {"w16", demoEconomia},
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
	e.Mural = append([]dados.Aviso{{At: agora.Add(-4 * time.Minute), Autor: "g-multas-custos",
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
	login := homol("PL-19597", "needs_input")
	login.Reason = "gate:login"
	pr := etapa("PR", "PL-19826-pr", "failed", "", "implementar.done")
	pr.Reason = "falhou"
	feita := etapa("pr", "PL-19830-pr", "done", "c-pl-19830", "corrigir.done")
	feita.Report = []byte(`{"status":"done","links":["https://github.com/prologapp/PrologWebReact/pull/3120"]}`)
	return []dados.Sidequest{
		{ID: 13, Pipeline: "implementar-homologar", Title: "Custo das multas por veículo", SlugBase: "PL-19597", Status: "running",
			Coordinator: "g-multas-custos", Inputs: map[string]string{"card": "PL-19597", "repos": "WS React", "homolog_alvo": "hml"},
			IdadeMin: 42, EsperaVoce: 1, Prontidao: "pronto",
			Etapas: []dados.Etapa{impl("PL-19597", "done", "c-pl-19597", "w15:t3"), login,
				{ID: "juntar", Role: "gengar", Status: "pending", Slug: dia + "-PL-19597-juntar", After: []string{"homologar.done"},
					Gates: []dados.Gate{{Item: "juntar", By: "human"}}}}},
		{ID: 11, Pipeline: "bugfix", Title: "Valor da multa arredondado", SlugBase: "PL-19830", Status: "done",
			Coordinator: "g-economia-multas", Inputs: map[string]string{"card": "PL-19830", "repos": "WS"}, IdadeMin: 26 * 60,
			Etapas: []dados.Etapa{etapa("debug", "PL-19830", "done", "c-pl-19830"),
				etapa("corrigir", "PL-19830-corrigir", "done", "c-pl-19830", "debug.done"), feita}},
		{ID: 14, Pipeline: "implementar-homologar", Title: "Economia com multas evitadas", SlugBase: "PL-19826", Status: "failed",
			Coordinator: "g-economia-multas", Prontidao: "aguardando prontidão",
			Inputs: map[string]string{"card": "PL-19826", "repos": "React"}, IdadeMin: 70,
			Etapas: []dados.Etapa{impl("PL-19826", "done", "c-pl-19826", ""), pr, homol("PL-19826", "pending")}},
		{ID: 12, Pipeline: "bugfix", Title: "Km travado no abastecimento", SlugBase: "PL-19848", Status: "running",
			Coordinator: "gengar-1030-0900", Inputs: map[string]string{"card": "PL-19848", "repos": "WS"}, IdadeMin: 95,
			Etapas: []dados.Etapa{etapa("debug", "PL-19848", "done", "c-pl-19848"),
				etapa("corrigir", "PL-19848-corrigir", "running", "c-pl-19848", "debug.done"),
				etapa("pr", "PL-19848-pr", "pending", "", "corrigir.done")}},
	}
}

func estadoDemoBase(agora time.Time) dados.Estado {
	return dados.Estado{
		At:   agora,
		Raiz: "/demo/spark",
		Mural: []dados.Aviso{
			{At: agora.Add(-8 * time.Minute), Autor: "gengar-1030-0900", Texto: "vou mexer no delegate-spawn até 12h"},
			{At: agora.Add(-95 * time.Minute), Autor: "gustavo", Texto: "HML fora até 15h, sem homologação por enquanto"},
		},
		Vivos:    map[string]bool{"c-pl-1": true, "c-pl-2": true, "c-pl-3": true, "c-pl-4": true},
		Recursos: dados.Recursos{GH: true, Herdr: true, Delegate: true, Painel: "/demo/spark/scripts/delegate/painel.py", Autor: "demo"},
		Tasks: []dados.Linha{
			{Chave: "t-working", Rotulo: "c-pl-1", Detalhe: "working · PL-1001 · implementando", Extra: "aba-1"},
			{Chave: "t-espera", Rotulo: "c-pl-2", Detalhe: "esperando você · PL-1002", Extra: "aba-2"},
			{Chave: "t-blocked", Rotulo: "c-pl-3", Detalhe: "blocked · PL-1003 · CI falhou", Extra: "aba-3"},
			{Chave: "t-done", Rotulo: "c-pl-4", Detalhe: "done · PL-1004 · relatório pronto", Extra: "aba-4"},
		},
		Linhas: []dados.LinhaViva{
			{Nome: "c-pl-2", Pasta: "/demo/spark/.delegate/c-pl-2", Aba: "aba-2", Etapa: "implementar", Agora: "1 sua(s): plano", Estado: dados.EsperaVoce, CtxPct: 41, Hora: agora.Add(-3 * time.Minute)},
			{Nome: "c-pl-1", Pasta: "/demo/spark/.delegate/c-pl-1", Aba: "aba-1", Etapa: "implementar", Agora: "rodando a suíte de integração do módulo de pneus com Testcontainers e conferindo a migração de índices — ação válida até o fim da conferência", Estado: dados.Trabalhando, CtxPct: 55, Hora: agora.Add(-1 * time.Minute)},
			{Nome: "h-pl-5", Pasta: "/demo/spark/.delegate/h-pl-5", Aba: "aba-5", Etapa: "homologar", Agora: "Playwright: navegando em /frota/pneus", Estado: dados.Trabalhando, CtxPct: 86, Hora: agora.Add(-2 * time.Minute)},
			{Nome: "c-pl-3", Pasta: "/demo/spark/.delegate/c-pl-3", Aba: "aba-3", Etapa: "implementar", Agora: "parado com tarefa há 14 min", Estado: dados.Parado, CtxPct: -1, Hora: agora.Add(-14 * time.Minute)},
			{Nome: "c-pl-4", Pasta: "/demo/spark/.delegate/c-pl-4", Aba: "aba-4", Etapa: "implementar", Agora: "entregou relatório", Estado: dados.Terminou, CtxPct: 100, Hora: agora.Add(-20 * time.Minute)},
		},
		Caidos: []dados.Caido{
			{Nome: "c-pl-7", Pane: "w1:p7", SessionID: "9f3c2a71-5b0e-4d8a-a1c4-2e7b60d1f5aa", Pasta: "/demo/spark/.delegate/c-pl-7"},
		},
		Esperando: []dados.Linha{
			{Chave: "t-espera/q1", Rotulo: "c-pl-2 · plano", Detalhe: "Aprovar o plano de migração da tela de frota?"},
		},
		Travados: []dados.Travado{
			{Repo: "PrologWebReact", Number: 3101, Motivo: "CI vermelho", Title: "[PL-1003] Filtro de frota", URL: "https://github.com/prologapp/PrologWebReact/pull/3101", Branch: "feature/filtro-frota-PL-1003", CI: "FAILURE", Falhos: []string{"lint", "vitest"}},
			{Repo: "PrologWebService", Number: 4202, Motivo: "aguardando revisão", Title: "[PL-1001] Endpoint de pneus", URL: "https://github.com/prologapp/PrologWebService/pull/4202", Branch: "feature/pneus-PL-1001", CI: "PENDING"},
			{Repo: "PrologFlutter", Number: 1303, Motivo: "aprovado · mergear", Title: "[PL-1005] Ajuste de tema", URL: "https://github.com/prologapp/PrologFlutter/pull/1303", Branch: "feature/tema-PL-1005", CI: "SUCCESS"},
			{Repo: "PrologE2E", Number: 504, Motivo: "aprovado · mergear", Title: "[PL-1006] Spec de checklist", URL: "https://github.com/prologapp/PrologE2E/pull/504", Branch: "feature/e2e-PL-1006", CI: "SUCCESS"},
			{Repo: "GlorpDesignSystem", Number: 605, Motivo: "aprovado · mergear", Title: "[PL-1007] Token de borda", URL: "https://github.com/prologapp/GlorpDesignSystem/pull/605", Branch: "feature/borda-PL-1007", CI: "SUCCESS"},
			{Repo: "PrologMessaging", Number: 706, Motivo: "aguardando revisão", Title: "[PL-1008] Evento de notificação", URL: "https://github.com/prologapp/PrologMessaging/pull/706", Branch: "feature/evento-PL-1008", CI: "SUCCESS"},
			{Repo: "PrologExportService", Number: 807, Motivo: "aguardando revisão", Title: "[PL-1009] Exportação sem CI", URL: "https://github.com/prologapp/PrologExportService/pull/807", Branch: "feature/export-PL-1009"},
		},
		Servicos: []dados.Linha{
			{Chave: "PrologWebReact", Rotulo: "no ar", Detalhe: ":3000"},
			{Chave: "PrologWebService", Rotulo: "parado", Detalhe: ""},
		},
		Maquina: []dados.Linha{
			{Chave: "trava", Rotulo: "trava de testes pesados", Detalhe: "livre", Extra: "ok"},
		},
	}
}
