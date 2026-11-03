package dados

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// spark cria uma raiz falsa com o delegate.py, com o .delegate/ de mtime t.
func spark(t *testing.T, dir string, ativa time.Time) string {
	t.Helper()
	s := filepath.Join(dir, "shared", "mods", "control-panel", "scripts")
	if err := os.MkdirAll(s, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(s, "delegate.py"), []byte("#"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".delegate"), 0o755)
	os.Chtimes(filepath.Join(dir, ".delegate"), ativa, ativa)
	return dir
}

func TestAcharRaizes(t *testing.T) {
	home := t.TempDir()
	agora := time.Now()
	velha := spark(t, filepath.Join(home, "orca", "spark"), agora.Add(-48*time.Hour))
	nova := spark(t, filepath.Join(home, "orca", "workspaces", "spark", "feat-x"), agora)
	// não contam: fundo demais, dentro de pasta oculta, de .worktrees e de node_modules
	spark(t, filepath.Join(home, "a", "b", "c", "d", "e", "f"), agora)
	spark(t, filepath.Join(home, ".cache", "spark"), agora)
	spark(t, filepath.Join(nova, ".worktrees", "task", "spark"), agora)
	spark(t, filepath.Join(home, "proj", "node_modules", "spark"), agora)
	// raiz dentro de raiz: só a de fora
	spark(t, filepath.Join(velha, "sub"), agora)

	got := AcharRaizes(context.Background(), home, 5)
	if !reflect.DeepEqual(got, []string{nova, velha}) {
		t.Fatalf("raízes = %v", got)
	}
	if !EhRaizSpark(nova) || EhRaizSpark(home) {
		t.Fatal("EhRaizSpark")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := AcharRaizes(ctx, home, 5); len(got) != 0 {
		t.Fatalf("ctx cancelado deveria parar: %v", got)
	}
}
