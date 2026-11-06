package ui

import (
	"strings"
	"testing"
	"time"

	"fantasmas/internal/dados"
)

func TestGavetaMaquinaMostraFaxina(t *testing.T) {
	e := fixture()
	e.Faxina = &dados.Faxina{Em: time.Date(2026, 11, 6, 14, 35, 0, 0, time.Local), Acoes: []string{"a", "b"}, MB: 1324}
	out := render(tecla(modelo(e, &registro{}), "4"), 140, 60)
	if !strings.Contains(out, "Faxina: 14:35 · 2 ações · ~1324 MB liberados") {
		t.Fatalf("linha da faxina ausente na gaveta máquina:\n%s", out)
	}
	e.Faxina = nil
	if out := render(tecla(modelo(e, &registro{}), "4"), 140, 60); strings.Contains(out, "Faxina:") {
		t.Fatal("sem faxina não deveria aparecer a linha")
	}
}
