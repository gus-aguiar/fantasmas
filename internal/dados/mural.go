package dados

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"time"
)

// Aviso é uma linha do .delegate/_mural.jsonl.
type Aviso struct {
	At    time.Time
	Autor string
	Texto string
	Ate   time.Time
	// Missao é o slug da missão do aviso; vazio é global.
	Missao string
}

const (
	janelaMural = 24 * time.Hour
	caudaMural  = 64 * 1024
)

// MuralDe lê a cauda do arquivo do mural e devolve até max avisos, do mais recente ao mais antigo,
// só das últimas 24 h e que não passaram do "ate". Arquivo ausente, linha ilegível
// e FANTASMAS_MURAL=off dão lista vazia ou pulam a linha.
func MuralDe(caminho string, agora time.Time, max int) []Aviso {
	if os.Getenv("FANTASMAS_MURAL") == "off" || max <= 0 {
		return nil
	}
	buf := LerCauda(caminho, caudaMural)
	if len(buf) == 0 {
		return nil
	}
	var out []Aviso
	r := bufio.NewReader(bytes.NewReader(buf))
	for {
		linha, err := r.ReadBytes('\n')
		if len(linha) > 0 {
			if a, ok := avisoDe(linha, agora); ok {
				out = append(out, a)
			}
		}
		if err != nil {
			break
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if len(out) > max {
		out = out[:max]
	}
	return out
}

func avisoDe(linha []byte, agora time.Time) (Aviso, bool) {
	var raw struct {
		At     string `json:"at"`
		Autor  string `json:"autor"`
		Texto  string `json:"texto"`
		Ate    string `json:"ate"`
		Missao string `json:"missao"`
	}
	if json.Unmarshal(linha, &raw) != nil || raw.Texto == "" {
		return Aviso{}, false
	}
	at, err := time.Parse(time.RFC3339, raw.At)
	if err != nil || agora.Sub(at) > janelaMural {
		return Aviso{}, false
	}
	a := Aviso{At: at, Autor: raw.Autor, Texto: raw.Texto, Missao: raw.Missao}
	if raw.Ate != "" {
		ate, err := time.Parse(time.RFC3339, raw.Ate)
		if err != nil {
			ate, err = time.Parse("2006-01-02T15:04Z07:00", raw.Ate)
		}
		if err == nil {
			if !ate.After(agora) {
				return Aviso{}, false
			}
			a.Ate = ate
		}
	}
	return a, true
}
