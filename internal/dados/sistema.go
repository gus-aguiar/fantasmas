package dados

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Memoria é o que o Esvaziar swap precisa de /proc/meminfo, em MB.
type Memoria struct {
	DisponivelMB, SwapUsadoMB int
}

const (
	// SwapVazioMB: abaixo disso o swap conta como vazio.
	SwapVazioMB = 100
	// FolgaSwapMB: a RAM disponível precisa superar o swap usado por essa folga.
	FolgaSwapMB = 2048
	// ChaveMeminfo troca o /proc/meminfo (testes).
	ChaveMeminfo = "FANTASMAS_MEMINFO"
)

// LerMemoria lê FANTASMAS_MEMINFO ou /proc/meminfo.
func LerMemoria() (Memoria, error) {
	arq := os.Getenv(ChaveMeminfo)
	if arq == "" {
		arq = "/proc/meminfo"
	}
	f, err := os.Open(arq)
	if err != nil {
		return Memoria{}, err
	}
	defer f.Close()
	kb := map[string]int{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		campos := strings.Fields(sc.Text())
		if len(campos) >= 2 {
			if n, err := strconv.Atoi(campos[1]); err == nil {
				kb[strings.TrimSuffix(campos[0], ":")] = n
			}
		}
	}
	if _, ok := kb["MemAvailable"]; !ok {
		return Memoria{}, fmt.Errorf("meminfo sem MemAvailable")
	}
	return Memoria{DisponivelMB: kb["MemAvailable"] / 1024, SwapUsadoMB: (kb["SwapTotal"] - kb["SwapFree"]) / 1024}, nil
}

// SwapVazio: nada a devolver para a RAM.
func (m Memoria) SwapVazio() bool { return m.SwapUsadoMB < SwapVazioMB }

// CabeNaRAM: a RAM disponível recebe o swap usado com folga.
func (m Memoria) CabeNaRAM() bool { return m.DisponivelMB > m.SwapUsadoMB+FolgaSwapMB }

// RodarFaxina roda uma rodada da faxina e devolve o resumo dela; nil quando ela não agiu
// (o faxina-ultima.json só é escrito quando há ação, então um arquivo anterior ao início não conta).
func RodarFaxina(ctx context.Context, r Runner, raiz string, inicio time.Time) (*Faxina, error) {
	if _, err := r.Run(ctx, "python3", filepath.Join(raiz, "scripts", "delegate", "faxina.py"), "rodar", "--uma-vez"); err != nil {
		return nil, err
	}
	f := FaxinaDe(raiz)
	if f == nil || f.Em.Before(inicio.Add(-time.Second)) {
		return nil, nil
	}
	return f, nil
}

// Curta é "N ações · ~X MB".
func (f *Faxina) Curta() string {
	acoes := "ações"
	if len(f.Acoes) == 1 {
		acoes = "ação"
	}
	return fmt.Sprintf("%d %s · ~%d MB", len(f.Acoes), acoes, f.MB)
}
