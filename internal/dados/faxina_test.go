package dados

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func escreveFaxina(t *testing.T, conteudo string) string {
	t.Helper()
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, ".tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, ".tmp", "faxina-ultima.json"), []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func TestFaxinaDe(t *testing.T) {
	raiz := escreveFaxina(t, `{"em": "2026-11-06T14:35:00-03:00", "liberado_mb_estimado": 1324,
		"acoes": [{"texto": "serviço dev: ocioso"}, {"texto": "daemon GradleDaemon\npid 9"}]}`)
	f := FaxinaDe(raiz)
	if f == nil {
		t.Fatal("esperava a faxina")
	}
	if f.MB != 1324 || len(f.Acoes) != 2 || f.Acoes[1] != "daemon GradleDaemon pid 9" {
		t.Fatalf("faxina = %+v", f)
	}
	hora := time.Date(2026, 11, 6, 17, 35, 0, 0, time.UTC).Local().Format("15:04")
	if got, want := f.Linha(), "Faxina: "+hora+" · 2 ações · ~1324 MB liberados"; got != want {
		t.Fatalf("Linha() = %q, quer %q", got, want)
	}
}

func TestFaxinaUmaAcao(t *testing.T) {
	f := &Faxina{Em: time.Now(), Acoes: []string{"x"}, MB: 0}
	if got := f.Linha(); got[len(got)-len("1 ação · ~0 MB liberados"):] != "1 ação · ~0 MB liberados" {
		t.Fatalf("Linha() = %q", got)
	}
}

func TestFaxinaAusenteOuQuebrada(t *testing.T) {
	if FaxinaDe(t.TempDir()) != nil {
		t.Fatal("sem arquivo deveria dar nil")
	}
	if FaxinaDe(escreveFaxina(t, "{")) != nil {
		t.Fatal("json quebrado deveria dar nil")
	}
	if FaxinaDe(escreveFaxina(t, `{"em": "ontem"}`)) != nil {
		t.Fatal("data ilegível deveria dar nil")
	}
}
