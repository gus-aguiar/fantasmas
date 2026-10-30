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

// EstadoDemo é um estado determinístico para a olhada renderizada (--demo).
func EstadoDemo(agora time.Time) dados.Estado {
	e := estadoDemoBase(agora)
	e.PipesLigado = true
	e.Runs = runsDemo(agora)
	e.Proximo = map[string]string{
		"2026-10-30-PL-1210":         "lançaria (implementer, skill writing-plans)",
		"2026-10-30-PL-1203-homolog": "não lança: aguarda implementar (after)",
	}
	return e
}

func runsDemo(agora time.Time) []dados.Run {
	dia := agora.Format("2006-01-02")
	impl := func(slug, status, reason, agent, tab string) dados.Etapa {
		return dados.Etapa{ID: "implementar", Role: "implementer", Skill: "writing-plans", Status: status, Reason: reason,
			Slug: dia + "-" + slug, Agent: agent, Tab: tab,
			Gates: []dados.Gate{{Item: "plano", By: "human"}, {Item: "juntar", By: "human"}}}
	}
	homol := func(slug, status string) dados.Etapa {
		return dados.Etapa{ID: "homologar", Role: "homolog", Skill: "homolog", Status: status, Slug: dia + "-" + slug + "-homolog",
			After: []string{"implementar.done"}, Gates: []dados.Gate{{Item: "login", By: "human"}}}
	}
	feita := impl("PL-1150", "done", "", "c-pl-1150", "")
	feita.Report = []byte(`{"status":"done","links":["https://github.com/prologapp/PrologWebReact/pull/3120"]}`)
	hfeita := homol("PL-1150", "done")
	hfeita.Report = []byte(`{"status":"done","links":["https://github.com/prologapp/PrologE2E/pull/512"]}`)
	return []dados.Run{
		{ID: 13, Pipeline: "implementar-homologar", Title: "Filtro de frota por unidade", SlugBase: "PL-1203", Status: "running",
			Coordinator: "gengar-1030-0900", Inputs: map[string]string{"card": "PL-1203", "repos": "WS React", "homolog_alvo": "hml"},
			IdadeMin: 42, EsperaVoce: 1,
			Etapas: []dados.Etapa{impl("PL-1203", "needs_input", "gate:plano", "c-pl-1203", "w1:t3"), homol("PL-1203", "pending")}},
		{ID: 12, Pipeline: "bugfix", Title: "Km travado no abastecimento", SlugBase: "PL-1188", Status: "running",
			Coordinator: "gengar-1030-0900", Inputs: map[string]string{"card": "PL-1188", "repos": "WS"}, IdadeMin: 95,
			Etapas: []dados.Etapa{{ID: "implementar", Role: "implementer", Skill: "debug-for-fix", Status: "running",
				Slug: dia + "-PL-1188", Agent: "c-pl-1188", Tab: "w1:t4", Gates: []dados.Gate{{Item: "plano", By: "human"}}}}},
		{ID: 11, Pipeline: "implementar-homologar", Title: "Exportar checklist em PDF", SlugBase: "PL-1150", Status: "done",
			Coordinator: "gengar-1030-0900", Inputs: map[string]string{"card": "PL-1150", "repos": "React", "homolog_alvo": "local"},
			IdadeMin: 26 * 60, Etapas: []dados.Etapa{feita, hfeita}},
		{ID: 10, Pipeline: "implementar", Title: "Tela de multas por veículo", SlugBase: "PL-1210", Status: "intake",
			Inputs: map[string]string{"card": "PL-1210", "repos": "React"}, IdadeMin: 3,
			Etapas: []dados.Etapa{impl("PL-1210", "ready", "", "", "")}},
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
			{Chave: "slot-1", Detalhe: "mvn há 4 min · KpisIT", Extra: "ok"},
		},
	}
}
