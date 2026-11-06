package dados

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Faxina é o resumo da última rodada da faxina de memória que agiu (.tmp/faxina-ultima.json,
// escrito por scripts/delegate/faxina.py).
type Faxina struct {
	Em    time.Time
	Acoes []string
	MB    int
}

// FaxinaDe lê <raiz>/.tmp/faxina-ultima.json; ausente ou ilegível dá nil.
func FaxinaDe(raiz string) *Faxina {
	b, err := os.ReadFile(filepath.Join(raiz, ".tmp", "faxina-ultima.json"))
	if err != nil {
		return nil
	}
	var raw struct {
		Em    string `json:"em"`
		MB    int    `json:"liberado_mb_estimado"`
		Acoes []struct {
			Texto string `json:"texto"`
		} `json:"acoes"`
	}
	if json.Unmarshal(b, &raw) != nil {
		return nil
	}
	em, err := time.Parse(time.RFC3339, raw.Em)
	if err != nil {
		return nil
	}
	f := &Faxina{Em: em, MB: raw.MB}
	for _, a := range raw.Acoes {
		f.Acoes = append(f.Acoes, UmaLinha(a.Texto))
	}
	return f
}

// Linha é "Faxina: HH:MM · N ações · ~X MB liberados", no fuso local.
func (f *Faxina) Linha() string {
	n := len(f.Acoes)
	acoes := "ações"
	if n == 1 {
		acoes = "ação"
	}
	return fmt.Sprintf("Faxina: %s · %d %s · ~%d MB liberados", f.Em.Local().Format("15:04"), n, acoes, f.MB)
}
