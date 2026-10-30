package ui

import (
	"time"

	"fantasmas/internal/dados"
)

// EstadoDemo é um estado determinístico para a olhada renderizada (--demo).
func EstadoDemo(agora time.Time) dados.Estado {
	return dados.Estado{
		At:       agora,
		Raiz:     "/demo/spark",
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
		Esperando: []dados.Linha{
			{Chave: "t-espera/q1", Rotulo: "c-pl-2 · plano", Detalhe: "Aprovar o plano de migração da tela de frota?"},
		},
		Travados: []dados.Travado{
			{Repo: "PrologWebReact", Number: 3101, Motivo: "CI vermelho", Title: "[PL-1003] Filtro de frota", URL: "https://github.com/prologapp/PrologWebReact/pull/3101", Branch: "feature/filtro-frota-PL-1003"},
			{Repo: "PrologWebService", Number: 4202, Motivo: "aguardando revisão", Title: "[PL-1001] Endpoint de pneus", URL: "https://github.com/prologapp/PrologWebService/pull/4202", Branch: "feature/pneus-PL-1001"},
			{Repo: "PrologFlutter", Number: 1303, Motivo: "aprovado · mergear", Title: "[PL-1005] Ajuste de tema", URL: "https://github.com/prologapp/PrologFlutter/pull/1303", Branch: "feature/tema-PL-1005"},
			{Repo: "PrologE2E", Number: 504, Motivo: "aprovado · mergear", Title: "[PL-1006] Spec de checklist", URL: "https://github.com/prologapp/PrologE2E/pull/504", Branch: "feature/e2e-PL-1006"},
			{Repo: "GlorpDesignSystem", Number: 605, Motivo: "aprovado · mergear", Title: "[PL-1007] Token de borda", URL: "https://github.com/prologapp/GlorpDesignSystem/pull/605", Branch: "feature/borda-PL-1007"},
			{Repo: "PrologMessaging", Number: 706, Motivo: "aguardando revisão", Title: "[PL-1008] Evento de notificação", URL: "https://github.com/prologapp/PrologMessaging/pull/706", Branch: "feature/evento-PL-1008"},
			{Repo: "PrologExportService", Number: 807, Motivo: "aguardando revisão", Title: "[PL-1009] Exportação sem CI", URL: "https://github.com/prologapp/PrologExportService/pull/807", Branch: "feature/export-PL-1009"},
		},
		CI: []dados.Linha{
			{Chave: "https://github.com/prologapp/PrologWebReact/actions/runs/1", Rotulo: "PrologWebReact · CI", Detalhe: "completed/failure · lint e testes"},
			{Chave: "https://github.com/prologapp/PrologWebService/actions/runs/2", Rotulo: "PrologWebService · CI", Detalhe: "in_progress/- · build e testes"},
			{Chave: "https://github.com/prologapp/PrologFlutter/actions/runs/3", Rotulo: "PrologFlutter · CI", Detalhe: "completed/success · analyze"},
			{Chave: "https://github.com/prologapp/PrologE2E/actions/runs/4", Rotulo: "PrologE2E · CI", Detalhe: "completed/success · smoke"},
			{Chave: "https://github.com/prologapp/GlorpDesignSystem/actions/runs/5", Rotulo: "GlorpDesignSystem · CI", Detalhe: "completed/success · build"},
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
