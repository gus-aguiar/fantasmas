package dados

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type runAtu struct {
	sujo      string
	heads     []string
	contagem  string
	buildErro bool
	chamadas  []string
}

func (f *runAtu) Run(_ context.Context, argv ...string) (string, error) {
	j := strings.Join(argv, " ")
	f.chamadas = append(f.chamadas, j)
	switch {
	case strings.Contains(j, "status --porcelain"):
		return f.sujo, nil
	case strings.Contains(j, "rev-parse HEAD"):
		h := f.heads[0]
		f.heads = f.heads[1:]
		return h + "\n", nil
	case strings.Contains(j, "rev-list --count"):
		return f.contagem + "\n", nil
	case strings.Contains(j, "rev-parse --is-inside-work-tree"):
		return "true\n", nil
	case strings.Contains(j, " build "):
		if f.buildErro {
			return "", errors.New("go: erro de compilação")
		}
		return "", os.WriteFile(argv[len(argv)-2], []byte("novo"), 0o755)
	}
	return "", nil
}

func repoFake(t *testing.T) string {
	d := t.TempDir()
	os.MkdirAll(filepath.Join(d, "bin"), 0o755)
	os.WriteFile(filepath.Join(d, "bin", "fantasmas"), []byte("antigo"), 0o755)
	os.WriteFile(filepath.Join(d, "go.mod"), []byte("module fantasmas\n\ngo 1.22\n"), 0o644)
	return d
}

func binario(t *testing.T, d string) string {
	b, _ := os.ReadFile(filepath.Join(d, "bin", "fantasmas"))
	return string(b)
}

func TestAtualizarMudancasLocais(t *testing.T) {
	r := &runAtu{sujo: " M main.go"}
	_, err := Atualizar(context.Background(), r, repoFake(t))
	if err == nil || err.Error() != "mudanças locais, não atualizei" {
		t.Fatalf("err = %v", err)
	}
	if len(r.chamadas) != 1 {
		t.Errorf("não devia seguir: %v", r.chamadas)
	}
}

func TestAtualizarJaAtualizado(t *testing.T) {
	d := repoFake(t)
	r := &runAtu{heads: []string{"a", "a"}}
	res, err := Atualizar(context.Background(), r, d)
	if err != nil || res.Trocou || res.Mensagem != "já está na última versão" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	for _, c := range r.chamadas {
		if strings.Contains(c, "build") {
			t.Errorf("recompilou: %s", c)
		}
	}
}

func TestAtualizarNovosCommitsBuildOk(t *testing.T) {
	d := repoFake(t)
	r := &runAtu{heads: []string{"a", "b"}, contagem: "3"}
	res, err := Atualizar(context.Background(), r, d)
	if err != nil || !res.Trocou || res.Novos != 3 {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if binario(t, d) != "novo" {
		t.Errorf("binário não trocado")
	}
	if _, err := os.Stat(filepath.Join(d, "bin", "fantasmas.novo")); !os.IsNotExist(err) {
		t.Errorf(".novo sobrou")
	}
}

func TestAtualizarBuildFalha(t *testing.T) {
	d := repoFake(t)
	os.WriteFile(filepath.Join(d, "bin", "fantasmas.novo"), []byte("lixo"), 0o755)
	r := &runAtu{heads: []string{"a", "b"}, contagem: "2", buildErro: true}
	res, err := Atualizar(context.Background(), r, d)
	if err == nil || !strings.Contains(err.Error(), "build falhou") || res.Trocou {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if binario(t, d) != "antigo" {
		t.Errorf("binário antigo mexido")
	}
	if _, err := os.Stat(filepath.Join(d, "bin", "fantasmas.novo")); !os.IsNotExist(err) {
		t.Errorf(".novo não apagado")
	}
}

func TestAcharRepo(t *testing.T) {
	d := repoFake(t)
	got, err := AcharRepo(context.Background(), &runAtu{}, filepath.Join(d, "bin", "fantasmas"))
	if err != nil || got != d {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if _, err := AcharRepo(context.Background(), &runAtu{}, filepath.Join(t.TempDir(), "fantasmas")); err == nil || !strings.Contains(err.Error(), "repo não encontrado") {
		t.Errorf("fora de bin/: err=%v", err)
	}
	vazio := t.TempDir()
	os.MkdirAll(filepath.Join(vazio, "bin"), 0o755)
	if _, err := AcharRepo(context.Background(), &runAtu{}, filepath.Join(vazio, "bin", "x")); err == nil {
		t.Errorf("sem go.mod devia falhar")
	}
}
