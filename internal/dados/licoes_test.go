package dados

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLicoesDe(t *testing.T) {
	agora := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC)
	arq := filepath.Join(t.TempDir(), "_licoes.jsonl")
	jsonl := `{"at":"2026-11-25T10:00:00Z","crew":"c-1","erro":"lint: import não usado na linha 12","correcao":"rodar o lint antes do commit"}
{"at":"2026-11-26T10:00:00Z","crew":"c-2","erro":"Lint:  import não usado na linha 40","correcao":"lint no pre-commit"}
lixo
{"at":"2026-09-01T10:00:00Z","crew":"c-3","erro":"velho demais","correcao":"x"}
{"at":"2026-11-26T11:00:00Z","crew":"c-4","erro":"teste de integração sem banco","correcao":"subir o Testcontainers"}
`
	if err := os.WriteFile(arq, []byte(jsonl), 0o600); err != nil {
		t.Fatal(err)
	}
	sq := []Sidequest{{Etapas: []Etapa{
		{Attempt: 2, Reason: "teste de integração sem banco"},
		{Attempt: 1, Reason: ""},
		{Attempt: 0, Reason: "CI vermelho"},
	}}}
	ls := LicoesDe(arq, sq, agora)
	if len(ls) != 2 {
		t.Fatalf("esperava 2 lições (linha velha e lixo fora): %+v", ls)
	}
	if ls[0].Erro != "teste de integração sem banco" || ls[0].Vezes != 3 || ls[0].Correcao != "subir o Testcontainers" {
		t.Fatalf("registro + 2 retries da sidequest: %+v", ls[0])
	}
	if ls[1].Vezes != 2 || ls[1].Correcao != "lint no pre-commit" || ls[1].Erro != "lint: import não usado na linha 12" {
		t.Fatalf("variações do mesmo erro juntas, correção mais nova: %+v", ls[1])
	}
	if got := LicoesDe(filepath.Join(t.TempDir(), "nada"), nil, agora); len(got) != 0 {
		t.Fatalf("sem fonte, sem lição: %+v", got)
	}
}
