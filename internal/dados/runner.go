package dados

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, argv ...string) (stdout string, err error)
}

// EnvRunner é o Runner que aceita variáveis extras no ambiente do processo.
type EnvRunner interface {
	RunEnv(ctx context.Context, env []string, argv ...string) (string, error)
}

// rodaComEnv usa o env do runner quando ele suporta; senão prefixa o comando com `env`.
func rodaComEnv(ctx context.Context, r Runner, env []string, argv ...string) (string, error) {
	if e, ok := r.(EnvRunner); ok {
		return e.RunEnv(ctx, env, argv...)
	}
	return r.Run(ctx, append(append([]string{"env"}, env...), argv...)...)
}

type ExecRunner struct{}

func (e ExecRunner) Run(ctx context.Context, argv ...string) (string, error) {
	return e.RunEnv(ctx, nil, argv...)
}

func (ExecRunner) RunEnv(ctx context.Context, env []string, argv ...string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("comando vazio")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(errb.String()); msg != "" {
			return out.String(), fmt.Errorf("%s: %w: %s", argv[0], err, msg)
		}
		return out.String(), fmt.Errorf("%s: %w", argv[0], err)
	}
	return out.String(), nil
}
