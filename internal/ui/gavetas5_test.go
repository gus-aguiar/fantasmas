package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"
)

func TestCincoGavetasETeclas(t *testing.T) {
	quer := []string{"agora", "missoes", "prs", "maquina", "comandos"}
	if len(Gavetas) != 5 {
		t.Fatalf("gavetas: %d", len(Gavetas))
	}
	m := modelo(fixture(), &registro{})
	if Gavetas[m.gaveta].ID != "agora" {
		t.Fatalf("abre em %s", Gavetas[m.gaveta].ID)
	}
	for i, id := range quer {
		m = tecla(m, string(rune('1'+i)))
		if Gavetas[m.gaveta].ID != id || Gavetas[m.gaveta].Tecla != string(rune('1'+i)) {
			t.Fatalf("tecla %d: %s", i+1, Gavetas[m.gaveta].ID)
		}
	}
	if m = tecla(m, "6"); Gavetas[m.gaveta].ID != "comandos" {
		t.Fatal("6 não é gaveta")
	}
	if m = tecla(m, "right"); Gavetas[m.gaveta].ID != "agora" {
		t.Fatal("→ devia dar a volta")
	}
}

func TestApelidosDeGaveta(t *testing.T) {
	for velho, novo := range map[string]string{"crew": "agora", "ambiente": "maquina", "worktrees": "maquina", "configuracoes": "comandos"} {
		if !GavetaValida(velho) || Gavetas[indiceGaveta(velho)].ID != novo {
			t.Fatalf("%s não leva a %s", velho, novo)
		}
	}
	for _, id := range IDsGavetas() {
		if _, ok := apelidosGaveta[id]; ok {
			t.Fatalf("%s listado como gaveta", id)
		}
	}
}

func TestTopoUmaLinhaEMascoteSome(t *testing.T) {
	m := modelo(fixture(), &registro{})
	ls, _ := m.linhasTopo(140, AlturaSemMascote-1)
	if len(ls) != 1 {
		t.Fatalf("abaixo de %d linhas o topo tem %d", AlturaSemMascote, len(ls))
	}
	p := render(m, 140, 20)
	if !strings.Contains(p, "▲ 1 sua") || strings.Contains(p, "r atualiza") {
		t.Fatalf("topo:\n%s", p)
	}
	if ls, _ = m.linhasTopo(140, 42); len(ls) > 4 {
		t.Fatalf("mascote com %d linhas", len(ls))
	}
}

func TestTopoAlertaSoQuandoHa(t *testing.T) {
	e := fixture()
	e.Maquina = []dados.Linha{{Chave: "mem", Rotulo: "memória", Extra: "ok"}}
	if strings.Contains(render(modelo(e, &registro{}), 140, 20), "⚠") {
		t.Fatal("alerta sem motivo")
	}
	e.Maquina[0].Extra = "erro"
	if !strings.Contains(render(modelo(e, &registro{}), 140, 20), "⚠ memória") {
		t.Fatal("sem alerta de memória")
	}
}

func TestSemContadorZerado(t *testing.T) {
	e := fixture()
	e.Linhas = []dados.LinhaViva{{Nome: "c-x", Estado: dados.Terminou, Agora: "entregou relatório"}}
	m := modelo(e, &registro{})
	for i := range Gavetas {
		if p := render(m.irGaveta(i), 140, 40); strings.Contains(p, "(0)") || strings.Contains(p, "fantasminhas (") {
			t.Fatalf("%s com contador zerado:\n%s", Gavetas[i].ID, p)
		}
	}
}

func TestCaidoAntigoVaiParaAntigos(t *testing.T) {
	e := fixture()
	e.Caidos = []dados.Caido{{Nome: "c-velho", Pane: "p1", SessionID: "s", Em: e.At.Add(-72 * time.Hour)}}
	m := modelo(e, &registro{})
	if p := render(m, 140, 40); strings.Contains(p, "c-velho") || !strings.Contains(p, "antigos (") {
		t.Fatalf("caído velho em destaque:\n%s", p)
	}
	m.antigas = true
	if !strings.Contains(render(m, 140, 40), "c-velho") {
		t.Fatal("caído velho sumiu dos antigos")
	}
}

func TestComandoPedeEnterDuasVezes(t *testing.T) {
	var pedidos []string
	m := modelo(fixture(), &registro{})
	m.acoes.Disparar = func(id string) (dados.Disparo, error) {
		pedidos = append(pedidos, id)
		return dados.Disparo{ID: id, Aba: "w2:t9", Nome: "dia-1106-0900", Inicio: time.Now()}, nil
	}
	m = tecla(m, "5")
	if a, _ := m.selecionado(); a.comando != "dia-abrir" {
		t.Fatalf("primeiro item: %+v", a)
	}
	n, cmd := m.tecla("enter")
	m = n.(Model)
	if cmd != nil || m.armadoCmd != "dia-abrir" {
		t.Fatal("primeiro enter devia só armar")
	}
	n, _ = m.tecla("j")
	if n.(Model).armadoCmd != "" {
		t.Fatal("outra tecla devia desarmar")
	}
	n, _ = n.(Model).tecla("k")
	n, _ = n.(Model).tecla("enter")
	n, cmd = n.(Model).tecla("enter")
	if cmd == nil {
		t.Fatal("segundo enter devia disparar")
	}
	n, _ = n.(Model).Update(cmd())
	m = n.(Model)
	if len(pedidos) != 1 || pedidos[0] != "dia-abrir" {
		t.Fatalf("pedidos %v", pedidos)
	}
	m.relogio = func() time.Time { return m.disparos["dia-abrir"].Inicio.Add(7 * time.Minute) }
	if p := render(m, 140, 40); !strings.Contains(p, "rodando há 7 min") {
		t.Fatalf("sem tempo do disparo:\n%s", p)
	}
	if a, _ := m.selecionado(); a.aba != "w2:t9" {
		t.Fatalf("g não iria à aba: %+v", a)
	}
}

func TestMCPSemSublista(t *testing.T) {
	e := fixture()
	e.MCP = dados.ProgressoMCP{Inicio: e.At, Sessoes: []dados.SessaoMCP{{Nome: "sessao-a", Tom: "ok"}, {Nome: "sessao-b", Tom: "ok"}}}
	p := render(modelo(e, &registro{}).irGaveta(indiceGaveta("comandos")), 140, 40)
	if strings.Contains(p, "sessao-a") || !strings.Contains(p, "2 ok") {
		t.Fatalf("reconectar MCPs devia ser uma linha só:\n%s", p)
	}
}
