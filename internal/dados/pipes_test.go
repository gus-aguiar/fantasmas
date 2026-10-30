package dados

import (
	"context"
	"strings"
	"testing"
	"time"
)

const estadoRuns = `{"tasks":[],"pendentes":[],"opcoes":{},"mural":[],"runs":[
{"id":2,"pipeline":"implementar-homologar","title":"Filtro de frota","slug_base":"PL-2","status":"running","coordinator":"g","inputs":{"card":"PL-2"},"created_at":"2026-10-30T10:00:00","idade_min":42,"espera_voce":1,
 "stages":[{"id":"implementar","role":"implementer","skill":"writing-plans","status":"needs_input","reason":"gate:plano","slug":"2026-10-30-PL-2","attempt":0,"after":[],"gates":[{"item":"plano","by":"human"}],"agent":"c-pl-2","tab":"w1:t2","gone":false,"started_at":null,"ended_at":null,"report":null},
  {"id":"homologar","role":"homolog","skill":"homolog","status":"pending","reason":null,"slug":"2026-10-30-PL-2-homolog","attempt":0,"after":["implementar"],"gates":[],"agent":null,"tab":null,"gone":false,"started_at":null,"ended_at":null,"report":{"status":"done","links":["https://x/pr/1"]}}]}]}`

func TestDeDdbLeRuns(t *testing.T) {
	p, err := DeDdb([]byte(estadoRuns), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Runs) != 1 {
		t.Fatalf("runs=%+v", p.Runs)
	}
	r := p.Runs[0]
	if r.ID != 2 || r.Pipeline != "implementar-homologar" || r.EsperaVoce != 1 || r.IdadeMin != 42 || len(r.Etapas) != 2 {
		t.Fatalf("run=%+v", r)
	}
	e := r.Etapas[0]
	if e.Status != "needs_input" || e.Reason != "gate:plano" || e.Agent != "c-pl-2" || e.Tab != "w1:t2" || len(e.Gates) != 1 {
		t.Fatalf("etapa=%+v", e)
	}
	if h := r.Etapas[1]; len(h.After) != 1 || h.After[0] != "implementar" {
		t.Fatalf("after=%v", h.After)
	}
	if ls := LinksDoReport(r.Etapas[1].Report); len(ls) != 1 || ls[0] != "https://x/pr/1" {
		t.Fatalf("links=%v", ls)
	}
}

func TestProximoDoTick(t *testing.T) {
	saida := "[dry-run] nada é gravado nem lançado; para valer: DDB_TICK=on ddb tick --executar\n" +
		"2026-10-30-PL-1: ready\n2026-10-30-PL-2: ready\n" +
		"lançaria 2026-10-30-PL-1 (implementer, skill writing-plans)\n" +
		"não lança 2026-10-30-PL-2: max_running (1/1)\n"
	p := ProximoDoTick(saida)
	if p["2026-10-30-PL-1"] != "lançaria (implementer, skill writing-plans)" {
		t.Errorf("PL-1=%q", p["2026-10-30-PL-1"])
	}
	if p["2026-10-30-PL-2"] != "não lança: max_running (1/1)" {
		t.Errorf("PL-2=%q", p["2026-10-30-PL-2"])
	}
	if _, ok := p["[dry-run]"]; ok {
		t.Error("cabeçalho do dry-run virou etapa")
	}
}

func TestColetaComDdbTrazRunsETick(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "ddb")
	f := novoFake()
	f.saida[" state --json"] = estadoRuns
	f.saida[" tick --dry-run"] = "lançaria 2026-10-30-PL-2-homolog (homolog, skill homolog)\n"
	e := (&Coletor{Raiz: raizDdb(t), R: f}).Coletar(context.Background())
	if !e.PipesLigado || len(e.Runs) != 1 {
		t.Fatalf("ligado=%v runs=%v", e.PipesLigado, e.Runs)
	}
	if e.Proximo["2026-10-30-PL-2-homolog"] == "" {
		t.Errorf("proximo=%v", e.Proximo)
	}
	for _, c := range f.calls {
		if strings.Contains(c, "--executar") {
			t.Errorf("coleta executou o tick: %s", c)
		}
	}
}

func TestSemChavePipesDesligado(t *testing.T) {
	t.Setenv("FANTASMAS_FONTE", "")
	e := (&Coletor{Raiz: raizDdb(t), R: novoFake()}).Coletar(context.Background())
	if e.PipesLigado || len(e.Runs) != 0 {
		t.Errorf("ligado=%v runs=%v", e.PipesLigado, e.Runs)
	}
}

func TestPipesMontaComandos(t *testing.T) {
	f := novoFake()
	f.saida[" pipeline list --json"] = `[{"name":"implementar","arquivo":"/s/implementar.yaml","version":1,"inputs":["card","repos"],"stages":["implementar"],"erros":[]}]`
	f.saida[" pipeline validate /s/implementar.yaml"] = "ok implementar v1 (1 etapas)\n"
	p := Pipes{R: f, Raiz: "/r", Script: "/r/ddb.py"}
	ps, err := p.Listar()
	if err != nil || len(ps) != 1 || ps[0].Arquivo != "/s/implementar.yaml" || len(ps[0].Inputs) != 2 {
		t.Fatalf("ps=%+v err=%v", ps, err)
	}
	if _, err := p.Novo("implementar", [][2]string{{"card", "PL-9"}, {"repos", "WS React"}}, "Meu título"); err != nil {
		t.Fatal(err)
	}
	if err := p.Etapa(3, "homologar", "retry"); err != nil {
		t.Fatal(err)
	}
	if out, err := p.Validar("/s/implementar.yaml"); err != nil || !strings.HasPrefix(out, "ok implementar") {
		t.Fatalf("validar=%q %v", out, err)
	}
	quer := []string{
		"python3 /r/ddb.py --raiz /r pipeline list --json",
		"python3 /r/ddb.py --raiz /r run new implementar --input card=PL-9 --input repos=WS React --title Meu título",
		"python3 /r/ddb.py --raiz /r stage 3:homologar retry",
		"python3 /r/ddb.py --raiz /r pipeline validate /s/implementar.yaml",
	}
	if strings.Join(f.calls, "\n") != strings.Join(quer, "\n") {
		t.Fatalf("calls:\n%s", strings.Join(f.calls, "\n"))
	}
	if err := p.Etapa(3, "homologar", "veto"); err == nil {
		t.Error("ação fora de advance|skip|retry aceita")
	}
}
