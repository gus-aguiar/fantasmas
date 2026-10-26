package dados

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProgressoMCPDe(t *testing.T) {
	ini := time.Unix(1_000_000, 0)
	tsv := "# inicio\t1000000\t--tsv\n# sessoes\t3\nw1:p1\tgengar\tok · 12s\tok\nw1:p2\tc-x\ttrabalhando · 0s\tpulada\n"
	p := ProgressoMCPDe(tsv, ini.Add(time.Minute))
	if !p.Rodando || p.Total != 3 || len(p.Sessoes) != 2 || p.Sessoes[1].Tom != "pulada" {
		t.Errorf("%+v", p)
	}
	if p := ProgressoMCPDe(tsv, ini.Add(31*time.Minute)); p.Rodando || !p.Interrompida {
		t.Errorf("sem fim em 30 min deveria ser interrompida: %+v", p)
	}
	if p := ProgressoMCPDe(tsv+"# fim\t1000100\n", ini.Add(time.Hour)); p.Rodando || p.Interrompida {
		t.Errorf("com fim: %+v", p)
	}
	if p := ProgressoMCPDe("# inicio\t1000000\n# fim\t1000001\therdr indisponível\n", ini); p.Nota != "herdr indisponível" {
		t.Errorf("%+v", p)
	}
}

func TestStatusAgente(t *testing.T) {
	if s := StatusAgente([]byte(`{"result":{"agent":{"agent_status":"working"}}}`)); s != "working" {
		t.Error(s)
	}
	if StatusAgente([]byte("lixo")) != "" {
		t.Error("ilegível deveria dar vazio")
	}
}

func TestDaydreamerArquivos(t *testing.T) {
	raiz := t.TempDir()
	d := filepath.Join(raiz, ".delegate", SlugDaydreamer)
	os.MkdirAll(d, 0o755)
	os.WriteFile(filepath.Join(d, "meta"), []byte("slug=daydreamer\nname=c-sonho\n"), 0o644)
	os.WriteFile(filepath.Join(d, "ultima"), []byte("2026-10-23T09:05\n"), 0o644)
	if n := NomeDaydreamer(raiz); n != "c-sonho" {
		t.Error(n)
	}
	if u := UltimaDaydreamer(raiz); u.Format("02/01 15:04") != "23/10 09:05" {
		t.Error(u)
	}
}
