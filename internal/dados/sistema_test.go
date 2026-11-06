package dados

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// meminfoFalso grava um meminfo com os valores em kB e aponta FANTASMAS_MEMINFO para ele.
func meminfoFalso(t *testing.T, disponivelKB, swapTotalKB, swapLivreKB int) {
	t.Helper()
	arq := filepath.Join(t.TempDir(), "meminfo")
	txt := "MemTotal:       32000000 kB\nMemAvailable:   " + strconv.Itoa(disponivelKB) + " kB\nSwapTotal:      " + strconv.Itoa(swapTotalKB) + " kB\nSwapFree:       " + strconv.Itoa(swapLivreKB) + " kB\n"
	if err := os.WriteFile(arq, []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ChaveMeminfo, arq)
}

func TestLerMemoriaEDecisaoDoSwap(t *testing.T) {
	meminfoFalso(t, 3*1024*1024, 4*1024*1024, 2*1024*1024) // 3 GB livres, 2 GB de swap usado
	m, err := LerMemoria()
	if err != nil || m.DisponivelMB != 3072 || m.SwapUsadoMB != 2048 {
		t.Fatalf("memória %+v %v", m, err)
	}
	if m.SwapVazio() || m.CabeNaRAM() {
		t.Fatalf("3 GB não recebem 2 GB + folga: %+v", m)
	}
	if !(Memoria{DisponivelMB: 8000, SwapUsadoMB: 2048}).CabeNaRAM() {
		t.Fatal("8 GB recebem 2 GB")
	}
	if !(Memoria{DisponivelMB: 8000, SwapUsadoMB: 99}).SwapVazio() {
		t.Fatal("99 MB é swap vazio")
	}
}

func TestDispararSwapNoWorkspaceDaTUI(t *testing.T) {
	agora := time.Date(2026, 11, 6, 14, 30, 0, 0, time.UTC)
	k, ok := ComandoPorID(CmdSwap)
	if !ok {
		t.Fatal("sem o swap no catálogo")
	}
	t.Setenv("HERDR_WORKSPACE_ID", "w7")
	r := &runnerHerdr{}
	if _, err := Disparar(context.Background(), r, "/spark", k, agora); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 2 || strings.Join(r.calls[0], " ") != "herdr tab create --workspace w7 --cwd /spark --label swap-1106-1430 --no-focus" {
		t.Fatalf("chamadas %v", r.calls)
	}
	if run := r.calls[1]; run[4] != "sudo swapoff -a && sudo swapon -a && free -h" {
		t.Fatalf("linha %v", run)
	}
	t.Setenv("HERDR_WORKSPACE_ID", "")
	r = &runnerHerdr{}
	if _, err := Disparar(context.Background(), r, "/spark", k, agora); err != nil {
		t.Fatal(err)
	}
	if strings.Join(r.calls[1][:5], " ") != "herdr tab create --workspace w2" {
		t.Fatalf("sem HERDR_WORKSPACE_ID devia cair no GENGAR: %v", r.calls)
	}
}

type runnerFaxina struct {
	raiz  string
	calls [][]string
	em    string
}

func (r *runnerFaxina) Run(_ context.Context, argv ...string) (string, error) {
	r.calls = append(r.calls, argv)
	if r.em != "" {
		_ = os.MkdirAll(filepath.Join(r.raiz, ".tmp"), 0o755)
		_ = os.WriteFile(filepath.Join(r.raiz, ".tmp", "faxina-ultima.json"),
			[]byte(`{"em":"`+r.em+`","liberado_mb_estimado":900,"acoes":[{"texto":"a"},{"texto":"b"}]}`), 0o644)
	}
	return "", nil
}

func TestRodarFaxina(t *testing.T) {
	inicio := time.Date(2026, 11, 6, 14, 30, 0, 0, time.UTC)
	r := &runnerFaxina{raiz: t.TempDir(), em: "2026-11-06T14:30:05Z"}
	f, err := RodarFaxina(context.Background(), r, r.raiz, inicio)
	if err != nil || f == nil || f.Curta() != "2 ações · ~900 MB" {
		t.Fatalf("faxina %+v %v", f, err)
	}
	if len(r.calls) != 1 || strings.Join(r.calls[0], " ") != "python3 "+r.raiz+"/scripts/delegate/faxina.py rodar --uma-vez" {
		t.Fatalf("chamadas %v", r.calls)
	}
	r.em = "2026-11-06T10:00:00Z" // arquivo velho: a rodada não agiu
	if f, err := RodarFaxina(context.Background(), r, r.raiz, inicio); err != nil || f != nil {
		t.Fatalf("arquivo anterior ao início não conta: %+v %v", f, err)
	}
}
