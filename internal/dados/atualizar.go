package dados

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Resultado da atualização: Trocou indica binário novo em bin/fantasmas.
type Resultado struct {
	Novos    int
	Trocou   bool
	Binario  string
	Mensagem string
}

const TempoAtualizar = 3 * time.Minute

// AcharRepo sobe do binário (bin/fantasmas) até a pasta do app e confere go.mod e git.
func AcharRepo(ctx context.Context, r Runner, exe string) (string, error) {
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	bin := filepath.Dir(exe)
	if filepath.Base(bin) != "bin" {
		return "", fmt.Errorf("repo não encontrado: %s não está em bin/", exe)
	}
	repo := filepath.Dir(bin)
	mod, err := os.ReadFile(filepath.Join(repo, "go.mod"))
	if err != nil || !strings.Contains(string(mod), "module fantasmas\n") {
		return "", fmt.Errorf("repo não encontrado: %s sem go.mod do fantasmas", repo)
	}
	if out, err := r.Run(ctx, "git", "-C", repo, "rev-parse", "--is-inside-work-tree"); err != nil || strings.TrimSpace(out) != "true" {
		return "", fmt.Errorf("repo não encontrado: %s não é repo git", repo)
	}
	return repo, nil
}

// Atualizar faz pull --ff-only e, havendo commit novo, recompila e troca bin/fantasmas atomicamente.
func Atualizar(ctx context.Context, r Runner, repo string) (Resultado, error) {
	ctx, cancel := context.WithTimeout(ctx, TempoAtualizar)
	defer cancel()
	git := func(args ...string) (string, error) {
		out, err := r.Run(ctx, append([]string{"git", "-C", repo}, args...)...)
		return strings.TrimSpace(out), err
	}
	sujo, err := git("status", "--porcelain", "--", ".")
	if err != nil {
		return Resultado{}, err
	}
	if sujo != "" {
		return Resultado{}, errors.New("mudanças locais, não atualizei")
	}
	antes, err := git("rev-parse", "HEAD")
	if err != nil {
		return Resultado{}, err
	}
	if _, err := git("pull", "--ff-only"); err != nil {
		return Resultado{}, fmt.Errorf("pull falhou: %w", err)
	}
	depois, err := git("rev-parse", "HEAD")
	if err != nil {
		return Resultado{}, err
	}
	n := 0
	if depois != antes {
		c, err := git("rev-list", "--count", antes+".."+depois)
		if err != nil {
			return Resultado{}, err
		}
		n, _ = strconv.Atoi(c)
	}
	if n == 0 {
		return Resultado{Mensagem: "já está na última versão"}, nil
	}
	novo := filepath.Join(repo, "bin", "fantasmas.novo")
	destino := filepath.Join(repo, "bin", "fantasmas")
	if _, err := rodaComEnv(ctx, r, []string{"GOTOOLCHAIN=local"}, "go", "-C", repo, "build", "-o", novo, "."); err != nil {
		os.Remove(novo)
		return Resultado{Novos: n}, fmt.Errorf("build falhou: %s", curto(err.Error()))
	}
	if err := os.Rename(novo, destino); err != nil {
		os.Remove(novo)
		return Resultado{Novos: n}, fmt.Errorf("não troquei o binário: %w", err)
	}
	return Resultado{Novos: n, Trocou: true, Binario: destino,
		Mensagem: fmt.Sprintf("atualizado: %d commits novos", n)}, nil
}

func curto(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 160 {
		return s[:160] + "…"
	}
	return s
}
