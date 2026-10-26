package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	SlugDaydreamer   = "daydreamer"
	timeoutSpawn     = 4 * time.Minute
	limiteReconectar = 30 * time.Minute
)

// Daydreamer é o estado do fantasminha que mantém o DayDreaming.
type Daydreamer struct {
	Nome   string
	Status string // working, idle… ou "" sem aba
	Ultima time.Time
	Erro   string
}

func NomeDaydreamer(raiz string) string {
	b, err := os.ReadFile(filepath.Join(raiz, ".delegate", SlugDaydreamer, "meta"))
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(l), "name="); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func UltimaDaydreamer(raiz string) time.Time {
	b, err := os.ReadFile(filepath.Join(raiz, ".delegate", SlugDaydreamer, "ultima"))
	if err != nil {
		return time.Time{}
	}
	t, _ := time.ParseInLocation("2006-01-02T15:04", strings.TrimSpace(string(b)), time.Local)
	return t
}

// StatusAgente lê .result.agent.agent_status do `herdr agent get`; "" quando não existe.
func StatusAgente(saida []byte) string {
	var r struct {
		Result struct {
			Agent struct {
				Status string `json:"agent_status"`
			} `json:"agent"`
		} `json:"result"`
	}
	if json.Unmarshal(saida, &r) != nil {
		return ""
	}
	return r.Result.Agent.Status
}

func LerDaydreamer(ctx context.Context, r Runner, raiz string) Daydreamer {
	d := Daydreamer{Nome: NomeDaydreamer(raiz), Ultima: UltimaDaydreamer(raiz)}
	if d.Nome == "" {
		return d
	}
	cctx, cancel := context.WithTimeout(ctx, timeoutProbe)
	defer cancel()
	if out, err := r.Run(cctx, "herdr", "agent", "get", d.Nome); err == nil {
		d.Status = StatusAgente([]byte(out))
	}
	return d
}

// AtualizarDaydreaming pede uma rodada: prompt se a aba está viva, spawn se não. Pode levar ~2 min.
func AtualizarDaydreaming(ctx context.Context, r Runner, raiz, brief string) (string, error) {
	d := LerDaydreamer(ctx, r, raiz)
	switch d.Status {
	case "working":
		return d.Nome + " já está numa rodada", nil
	case "":
	default:
		cctx, cancel := context.WithTimeout(ctx, timeoutProbe)
		defer cancel()
		if _, err := r.Run(cctx, "herdr", "agent", "prompt", d.Nome, "/daydreaming"); err != nil {
			return "", err
		}
		return "rodada pedida a " + d.Nome, nil
	}
	if os.Getenv("HERDR_ENV") != "1" || os.Getenv("HERDR_WORKSPACE_ID") == "" {
		return "", fmt.Errorf("abra o fantasmas dentro do Herdr para criar a aba do daydreamer")
	}
	cctx, cancel := context.WithTimeout(ctx, timeoutSpawn)
	defer cancel()
	cmd := exec.CommandContext(cctx, "bash", filepath.Join(raiz, "scripts", "delegate", "delegate-spawn.sh"),
		SlugDaydreamer, brief, "--model", "opus", "--effort", "medium")
	cmd.Dir = raiz
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("spawn: %w: %s", err, Corta(strings.TrimSpace(string(out)), 160))
	}
	return "aba do daydreamer aberta", nil
}

func Corta(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

// ArquivoMCP é onde o mcp_reconectar.py grava o progresso.
func ArquivoMCP(raiz string) string { return filepath.Join(raiz, ".tmp", "mcp-reconectar.tsv") }

// ReconectarMCPs dispara o mcp_reconectar.py destacado: sobrevive ao fechar o fantasmas.
func ReconectarMCPs(raiz string) error {
	if p := LerProgressoMCP(raiz, time.Now()); p.Rodando {
		return fmt.Errorf("já reconectando (%d/%d)", len(p.Sessoes), p.Total)
	}
	saida := ArquivoMCP(raiz)
	if err := os.MkdirAll(filepath.Dir(saida), 0o755); err != nil {
		return err
	}
	// grava o '# inicio' antes do script, para a UI já ver a rodada em andamento
	if err := os.WriteFile(saida, []byte(fmt.Sprintf("# inicio\t%d\n", time.Now().Unix())), 0o644); err != nil {
		return err
	}
	script := filepath.Join(raiz, "shared", "mods", "control-panel", "scripts", "mcp_reconectar.py")
	cmd := exec.Command("python3", script, "--tsv", "--saida", saida)
	cmd.Dir = raiz
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

type SessaoMCP struct{ Pane, Nome, Detalhe, Tom string }

type ProgressoMCP struct {
	Inicio       time.Time
	Total        int
	Sessoes      []SessaoMCP
	Rodando      bool
	Interrompida bool
	Nota         string // texto do '# fim', ex.: herdr indisponível
}

func LerProgressoMCP(raiz string, agora time.Time) ProgressoMCP {
	b, err := os.ReadFile(ArquivoMCP(raiz))
	if err != nil {
		return ProgressoMCP{}
	}
	return ProgressoMCPDe(string(b), agora)
}

func ProgressoMCPDe(s string, agora time.Time) ProgressoMCP {
	var p ProgressoMCP
	fim := false
	for _, l := range strings.Split(s, "\n") {
		c := strings.Split(strings.TrimSuffix(l, "\r"), "\t")
		switch c[0] {
		case "":
			continue
		case "# inicio":
			if len(c) > 1 {
				if n, err := strconv.ParseInt(c[1], 10, 64); err == nil {
					p.Inicio = time.Unix(n, 0)
				}
			}
		case "# sessoes":
			if len(c) > 1 {
				p.Total, _ = strconv.Atoi(c[1])
			}
		case "# fim":
			fim = true
			if len(c) > 2 {
				p.Nota = c[2]
			}
		default:
			if len(c) >= 4 && !strings.HasPrefix(c[0], "#") {
				p.Sessoes = append(p.Sessoes, SessaoMCP{c[0], c[1], c[2], c[3]})
			}
		}
	}
	if !fim && !p.Inicio.IsZero() {
		if agora.Sub(p.Inicio) > limiteReconectar {
			p.Interrompida = true
		} else {
			p.Rodando = true
		}
	}
	return p
}
