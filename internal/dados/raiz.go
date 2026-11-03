package dados

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// marcaRaiz é o script do Spark de que as gavetas mais dependem: com ele, a pasta é uma raiz utilizável.
var marcaRaiz = filepath.Join("shared", "mods", "control-panel", "scripts", "delegate.py")

// EhRaizSpark diz se dir tem os scripts do Spark que o fantasmas lê.
func EhRaizSpark(dir string) bool {
	st, err := os.Stat(filepath.Join(dir, marcaRaiz))
	return err == nil && !st.IsDir()
}

// pulaNaBusca são pastas grandes que nunca são raiz do Spark.
var pulaNaBusca = map[string]bool{"node_modules": true, "vendor": true, "snap": true, "go": true, "Library": true, "target": true, "dist": true, "build": true}

// AcharRaizes procura raízes do Spark até prof níveis abaixo de base, sem entrar em pastas ocultas
// (inclui .worktrees: worktree de task não é a raiz principal) nem nas de pulaNaBusca.
// Devolve da mais ativa para a menos ativa (mtime do .delegate/, ou da própria pasta).
func AcharRaizes(ctx context.Context, base string, prof int) []string {
	base = filepath.Clean(base)
	nivel := func(p string) int {
		r, err := filepath.Rel(base, p)
		if err != nil || r == "." {
			return 0
		}
		return strings.Count(r, string(filepath.Separator)) + 1
	}
	var achadas []string
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return filepath.SkipAll
		}
		if err != nil || !d.IsDir() {
			return nil
		}
		if p != base && (strings.HasPrefix(d.Name(), ".") || pulaNaBusca[d.Name()]) {
			return filepath.SkipDir
		}
		if EhRaizSpark(p) {
			achadas = append(achadas, p)
			return filepath.SkipDir
		}
		if nivel(p) >= prof {
			return filepath.SkipDir
		}
		return nil
	})
	atividade := func(p string) time.Time {
		for _, q := range []string{filepath.Join(p, ".delegate"), p} {
			if st, err := os.Stat(q); err == nil {
				return st.ModTime()
			}
		}
		return time.Time{}
	}
	sort.SliceStable(achadas, func(i, j int) bool { return atividade(achadas[i]).After(atividade(achadas[j])) })
	return achadas
}
