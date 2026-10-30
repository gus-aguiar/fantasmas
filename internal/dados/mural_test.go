package dados

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func escreveMural(t *testing.T, linhas ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "_mural.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(linhas, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestMuralDe(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	p := escreveMural(t,
		`{"at":"2026-10-30T11:00:00Z","autor":"a","texto":"um"}`,
		`{"at":"2026-10-29T11:00:00Z","autor":"velho","texto":"passou de 24h"}`,
		`{"at":"2026-10-30T09:00:00Z","autor":"b","texto":"expirou","ate":"2026-10-30T11:30:00Z"}`,
		`{"at":"2026-10-30T10:00:00Z","autor":"c","texto":"vale","ate":"2026-10-30T15:00:00Z"}`,
		`{"at":"2026-10-30T11:30:00Z","autor":"d","texto":"quatro"}`,
		`{"at":"2026-10-30T08:00:00Z","autor":"e","texto":"cinco"}`,
	)
	got := MuralDe(p, agora, 3)
	var autores []string
	for _, a := range got {
		autores = append(autores, a.Autor)
	}
	if strings.Join(autores, ",") != "d,a,c" {
		t.Fatalf("autores = %v, quer d,a,c (mais recente primeiro, 3, sem velho nem expirado)", autores)
	}
	if got[0].Texto != "quatro" || !got[2].Ate.After(agora) {
		t.Fatalf("campos: %+v", got)
	}
}

func TestMuralLinhaQuebrada(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	p := escreveMural(t,
		`{"at":"2026-10-30T11:00:00Z","autor":"a","texto":"ok"}`,
		`{"at":"2026-10-30T11:0`,
		`lixo`,
		``,
		`{"at":"ontem","autor":"x","texto":"data ruim"}`,
		`{"at":"2026-10-30T11:10:00Z","autor":"b","texto":"outro"}`,
	)
	got := MuralDe(p, agora, 5)
	if len(got) != 2 || got[0].Autor != "b" || got[1].Autor != "a" {
		t.Fatalf("got = %+v", got)
	}
}

func TestMuralAusente(t *testing.T) {
	agora := time.Now()
	if got := MuralDe(filepath.Join(t.TempDir(), "nao-existe.jsonl"), agora, 3); len(got) != 0 {
		t.Fatalf("got = %+v", got)
	}
	t.Setenv("FANTASMAS_MURAL", "off")
	p := escreveMural(t, `{"at":"`+agora.Format(time.RFC3339)+`","autor":"a","texto":"x"}`)
	if got := MuralDe(p, agora, 3); len(got) != 0 {
		t.Fatalf("FANTASMAS_MURAL=off devia desligar: %+v", got)
	}
}

func TestMuralSoLeACauda(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	var linhas []string
	velho := `{"at":"2026-10-20T11:00:00Z","autor":"velho","texto":"` + strings.Repeat("x", 150) + `"}`
	for i := 0; i < 600; i++ {
		linhas = append(linhas, velho)
	}
	linhas = append(linhas, `{"at":"2026-10-30T11:55:00Z","autor":"novo","texto":"no fim"}`)
	p := escreveMural(t, linhas...)
	if st, _ := os.Stat(p); st.Size() <= 64*1024 {
		t.Fatalf("arquivo de teste pequeno demais: %d", st.Size())
	}
	got := MuralDe(p, agora, 3)
	if len(got) != 1 || got[0].Autor != "novo" {
		t.Fatalf("got = %+v", got)
	}
}

func TestMuralLinhaGiganteNaoAborta(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	p := escreveMural(t,
		`{"at":"2026-10-30T11:00:00Z","autor":"a","texto":"`+strings.Repeat("y", 40*1024)+`"}`,
		`{"at":"2026-10-30T11:30:00Z","autor":"b","texto":"depois"}`,
	)
	got := MuralDe(p, agora, 3)
	if len(got) != 2 || got[0].Autor != "b" {
		t.Fatalf("got = %+v", got)
	}
}

func TestMuralAteSemSegundos(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	p := escreveMural(t,
		`{"at":"2026-10-30T10:00:00Z","autor":"vale","texto":"x","ate":"2026-10-30T12:30Z"}`,
		`{"at":"2026-10-30T10:30:00Z","autor":"expirou","texto":"x","ate":"2026-10-30T11:45Z"}`,
	)
	got := MuralDe(p, agora, 3)
	if len(got) != 1 || got[0].Autor != "vale" {
		t.Fatalf("got = %+v", got)
	}
}

func TestMuralAtNoFuturoApareceNoTopo(t *testing.T) {
	agora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	p := escreveMural(t,
		`{"at":"2026-10-30T11:00:00Z","autor":"a","texto":"x"}`,
		`{"at":"2026-10-30T13:00:00Z","autor":"futuro","texto":"x"}`,
	)
	got := MuralDe(p, agora, 3)
	if len(got) != 2 || got[0].Autor != "futuro" {
		t.Fatalf("got = %+v", got)
	}
}
