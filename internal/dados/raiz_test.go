package dados

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// raizFalsa cria uma raiz falsa com o delegate.py, com o .delegate/ de mtime t.
func raizFalsa(t *testing.T, dir string, ativa time.Time) string {
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
	velha := raizFalsa(t, filepath.Join(home, "projetos", "ws"), agora.Add(-48*time.Hour))
	nova := raizFalsa(t, filepath.Join(home, "projetos", "workspaces", "ws", "feat-x"), agora)
	// não contam: fundo demais, dentro de pasta oculta, de .worktrees e de node_modules
	raizFalsa(t, filepath.Join(home, "a", "b", "c", "d", "e", "f"), agora)
	raizFalsa(t, filepath.Join(home, ".cache", "ws"), agora)
	raizFalsa(t, filepath.Join(nova, ".worktrees", "task", "ws"), agora)
	raizFalsa(t, filepath.Join(home, "proj", "node_modules", "ws"), agora)
	// raiz dentro de raiz: só a de fora
	raizFalsa(t, filepath.Join(velha, "sub"), agora)

	got := AcharRaizes(context.Background(), home, 5)
	if !reflect.DeepEqual(got, []string{nova, velha}) {
		t.Fatalf("raízes = %v", got)
	}
	if !EhRaiz(nova) || EhRaiz(home) {
		t.Fatal("EhRaiz")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := AcharRaizes(ctx, home, 5); len(got) != 0 {
		t.Fatalf("ctx cancelado deveria parar: %v", got)
	}
}
