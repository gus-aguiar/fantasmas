package ui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
)

// meminfo grava um /proc/meminfo falso (MB) e aponta FANTASMAS_MEMINFO para ele.
func meminfo(t *testing.T, disponivelMB, swapUsadoMB int) {
	t.Helper()
	arq := filepath.Join(t.TempDir(), "meminfo")
	kb := func(mb int) string { return strconv.Itoa(mb*1024) + " kB\n" }
	txt := "MemAvailable: " + kb(disponivelMB) + "SwapTotal: " + kb(8192) + "SwapFree: " + kb(8192-swapUsadoMB)
	if err := os.WriteFile(arq, []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(dados.ChaveMeminfo, arq)
}

func paletaSistema(t *testing.T) (Model, *[]string) {
	t.Helper()
	var pedidos []string
	m := modelo(fixture(), &registro{}).ComBotoes()
	m.acoes.Memoria = dados.LerMemoria
	m.acoes.Disparar = func(id string) (dados.Disparo, error) {
		pedidos = append(pedidos, id)
		return dados.Disparo{ID: id, Aba: "w2:t9", Inicio: time.Now()}, nil
	}
	return m, &pedidos
}

func aperta(m Model, k rune) (Model, tea.Cmd) {
	n, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{k}})
	return n.(Model), cmd
}

func TestFaxinaAgoraRodaCom2xEMostraResumo(t *testing.T) {
	meminfo(t, 16000, 0)
	m, _ := paletaSistema(t)
	agora := time.Date(2026, 11, 6, 15, 0, 0, 0, time.UTC)
	m.relogio = func() time.Time { return agora }
	vezes := 0
	m.acoes.Faxina = func() (*dados.Faxina, error) {
		vezes++
		return &dados.Faxina{Acoes: []string{"a", "b", "c"}, MB: 1200}, nil
	}
	m, cmd := aperta(m, '5')
	if cmd != nil || m.armadoCmd != dados.CmdFaxina || vezes != 0 {
		t.Fatalf("primeiro 5 só arma: %q", m.armadoCmd)
	}
	m, cmd = aperta(m, '5')
	if cmd == nil || !m.faxinando || !strings.Contains(render(m, 140, 40), "rodando…") {
		t.Fatal("segundo 4 roda e mostra rodando…")
	}
	m = manda(m, cmd())
	if vezes != 1 {
		t.Fatalf("faxina rodou %d vezes", vezes)
	}
	if p := render(m, 140, 40); !strings.Contains(p, "3 ações · ~1200 MB") {
		t.Fatalf("sem o resumo:\n%s", p)
	}
	agora = agora.Add(DuraResumoFaxina + time.Second)
	if strings.Contains(render(m, 140, 40), "3 ações") {
		t.Fatal("resumo devia sumir depois de 2 min")
	}
}

func TestSwapRecusaComPoucaRAM(t *testing.T) {
	meminfo(t, 3000, 2000) // 3000 não supera 2000 + 2048
	m, pedidos := paletaSistema(t)
	for range 2 {
		m, _ = aperta(m, '6')
	}
	if len(*pedidos) != 0 || m.armadoCmd != "" {
		t.Fatalf("com pouca RAM não dispara: %v %q", *pedidos, m.armadoCmd)
	}
	if !strings.Contains(m.toast, "pouca RAM livre para receber o swap") {
		t.Fatalf("aviso: %q", m.toast)
	}
}

func TestSwapDisparaCom2x(t *testing.T) {
	meminfo(t, 9000, 2000)
	m, pedidos := paletaSistema(t)
	m, cmd := aperta(m, '6')
	if cmd != nil || m.armadoCmd != dados.CmdSwap {
		t.Fatalf("primeiro 6 só arma: %q", m.armadoCmd)
	}
	_, cmd = aperta(m, '6')
	if cmd == nil {
		t.Fatal("segundo 6 devia disparar")
	}
	cmd()
	if len(*pedidos) != 1 || (*pedidos)[0] != dados.CmdSwap {
		t.Fatalf("pedidos %v", *pedidos)
	}
	if k, _ := dados.ComandoPorID(dados.CmdSwap); k.Linha != "sudo swapoff -a && sudo swapon -a && free -h" || k.Workspace != "" {
		t.Fatalf("comando do swap: %+v", k)
	}
}

func TestSwapVazioNaoDispara(t *testing.T) {
	meminfo(t, 9000, 50)
	m, pedidos := paletaSistema(t)
	if !strings.Contains(render(m, 140, 40), "swap já vazio") {
		t.Fatal("item devia dizer swap já vazio")
	}
	for range 2 {
		m, _ = aperta(m, '6')
	}
	if len(*pedidos) != 0 {
		t.Fatalf("swap vazio não dispara: %v", *pedidos)
	}
}
