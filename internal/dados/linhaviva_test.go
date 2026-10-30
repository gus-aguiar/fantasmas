package dados

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMontarLinhaVivaRegras(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	recente := Atividade{Acao: "roda testes", Quando: agora.Add(-time.Minute), Tokens: 100_000}
	antiga := Atividade{Acao: "lê x.go", Quando: agora.Add(-25 * time.Minute)}
	casos := []struct {
		nome      string
		ag        Agente
		ses       *SessaoClaude
		at        Atividade
		pend      []string
		report    bool
		visto     bool
		estado    EstadoLinha
		agoraWant string
	}{
		{"regra1", Agente{Status: "idle"}, nil, recente, []string{"q1", "q2"}, false, false, EsperaVoce, "2 sua(s): q1, q2"},
		{"pendente vence working", Agente{Status: "working"}, nil, recente, []string{"q1"}, false, false, EsperaVoce, "1 sua(s): q1"},
		{"regra2 herdr", Agente{Status: "working"}, nil, recente, nil, false, false, Trabalhando, "roda testes"},
		{"regra2 sessao", Agente{Status: "idle"}, &SessaoClaude{Status: "busy"}, recente, nil, false, false, Trabalhando, "roda testes"},
		{"regra2 shell", Agente{Status: "idle"}, &SessaoClaude{Status: "shell"}, recente, nil, false, false, Trabalhando, "roda testes"},
		{"regra3", Agente{Status: "done"}, nil, recente, nil, true, false, Terminou, "entregou relatório"},
		{"regra3 visto cai na 5", Agente{Status: "done"}, nil, recente, nil, true, true, Trabalhando, "roda testes"},
		{"regra4", Agente{Status: "idle"}, nil, antiga, nil, false, false, Parado, "parado com tarefa há 25 min"},
		{"regra5", Agente{Status: "idle"}, nil, recente, nil, false, false, Trabalhando, "roda testes"},
		{"sem acao", Agente{Status: "idle"}, nil, Atividade{}, nil, false, false, Trabalhando, "—"},
	}
	for _, c := range casos {
		l := MontarLinhaViva("c-x", c.ag, c.ses, c.at, c.pend, c.report, c.visto, agora)
		if l.Estado != c.estado || l.Agora != c.agoraWant {
			t.Errorf("%s: estado=%v agora=%q", c.nome, l.Estado, l.Agora)
		}
	}
	l := MontarLinhaViva("c-x", Agente{Aba: "w1:t9"}, nil, recente, nil, false, false, agora)
	if l.Aba != "w1:t9" || l.CtxPct != 10 || !l.Hora.Equal(recente.Quando) {
		t.Error(l)
	}
	for nome, etapa := range map[string]string{"c-a": "implementar", "h-a": "homologar", "g-a": "gengar", "x-a": "—"} {
		if e := MontarLinhaViva(nome, Agente{}, nil, Atividade{}, nil, false, false, agora).Etapa; e != etapa {
			t.Error(nome, e)
		}
	}
}

func fixtureViva(t *testing.T) (raiz, claude string) {
	raiz, claude = t.TempDir(), t.TempDir()
	escreve(t, filepath.Join(raiz, ".delegate", "s-a", "meta"), "name=c-a\n")
	return
}

func TestLinhasVivasIgnoraSemMeta(t *testing.T) {
	raiz, claude := fixtureViva(t)
	e := Estado{Agentes: map[string]Agente{"gengar-1026": {Status: "idle"}, "c-a": {Status: "working", Aba: "w:t1"}}}
	ls := LinhasVivas(e, raiz, claude, time.Now())
	if len(ls) != 1 || ls[0].Nome != "c-a" || ls[0].Pasta != filepath.Join(raiz, ".delegate", "s-a") || ls[0].Aba != "w:t1" {
		t.Error(ls)
	}
}

func TestLinhasVivasArquivoQuebrado(t *testing.T) {
	raiz, claude := fixtureViva(t)
	escreve(t, filepath.Join(claude, "sessions", "1.json"), "lixo")
	e := Estado{Agentes: map[string]Agente{"c-a": {Status: "idle", Sessao: "nao-existe"}}}
	ls := LinhasVivas(e, raiz, claude, time.Now())
	if len(ls) != 1 || ls[0].Agora != "—" || ls[0].CtxPct != -1 {
		t.Error(ls)
	}
}

func TestLinhasVivasLeTranscriptEPendentes(t *testing.T) {
	raiz, claude := fixtureViva(t)
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	escreve(t, filepath.Join(claude, "sessions", "1.json"), `{"sessionId":"s1","name":"c-a","cwd":"/w/x","status":"idle","updatedAt":1}`)
	linha := `{"type":"assistant","timestamp":"` + agora.Add(-time.Minute).Format(time.RFC3339Nano) + `","message":{"model":"m","usage":{"input_tokens":500000},"content":[{"type":"tool_use","name":"Bash","input":{"description":"roda o build"}}]}}` + "\n"
	escreve(t, CaminhoTranscript(claude, "/w/x", "s1"), linha)
	e := Estado{Agentes: map[string]Agente{"c-a": {Status: "idle", Sessao: "s1"}}}
	ls := LinhasVivas(e, raiz, claude, agora)
	if len(ls) != 1 || ls[0].Agora != "roda o build" || ls[0].CtxPct != 50 || ls[0].Estado != Trabalhando {
		t.Error(ls)
	}
	e.Esperando = []Linha{{Chave: "/p/questions.json#q7", Rotulo: "c-a · pergunta"}}
	ls = LinhasVivas(e, raiz, claude, agora)
	if ls[0].Estado != EsperaVoce || !strings.Contains(ls[0].Agora, "q7") {
		t.Error(ls)
	}
}

func TestLinhasVivasTranscriptSemSessionsCaiNaRaiz(t *testing.T) {
	raiz, claude := fixtureViva(t)
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	linha := `{"type":"assistant","timestamp":"` + agora.Add(-time.Minute).Format(time.RFC3339Nano) + `","message":{"model":"m","usage":{"input_tokens":200000},"content":[{"type":"tool_use","name":"Bash","input":{"description":"compila"}}]}}` + "\n"
	escreve(t, CaminhoTranscript(claude, raiz, "s9"), linha)
	e := Estado{Agentes: map[string]Agente{"c-a": {Status: "idle", Sessao: "s9"}}}
	ls := LinhasVivas(e, raiz, claude, agora)
	if len(ls) != 1 || ls[0].Agora != "compila" || ls[0].CtxPct != 20 {
		t.Error(ls)
	}
}

func TestLinhasVivasIdDoHerdrVenceFallbackPorNome(t *testing.T) {
	raiz, claude := fixtureViva(t)
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	mk := func(acao string) string {
		return `{"type":"assistant","timestamp":"` + agora.Add(-time.Minute).Format(time.RFC3339Nano) + `","message":{"content":[{"type":"tool_use","name":"Bash","input":{"description":"` + acao + `"}}]}}` + "\n"
	}
	escreve(t, filepath.Join(claude, "sessions", "1.json"), `{"sessionId":"antigo","name":"c-a","cwd":"/w/x","status":"idle","updatedAt":1}`)
	escreve(t, CaminhoTranscript(claude, "/w/x", "antigo"), mk("da sessão antiga"))
	escreve(t, CaminhoTranscript(claude, "/w/x", "do-herdr"), mk("da sessão do herdr"))
	e := Estado{Agentes: map[string]Agente{"c-a": {Status: "idle", Sessao: "do-herdr"}}}
	ls := LinhasVivas(e, raiz, claude, agora)
	if len(ls) != 1 || ls[0].Agora != "da sessão do herdr" {
		t.Error(ls)
	}
}
