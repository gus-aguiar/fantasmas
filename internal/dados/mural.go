package dados

import (
	"bufio"
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
}

const janelaMural = 24 * time.Hour

// MuralDe lê o arquivo do mural e devolve até max avisos, do mais recente ao mais antigo,
// só das últimas 24 h e que não passaram do "ate". Arquivo ausente, linha ilegível
// e FANTASMAS_MURAL=off dão lista vazia ou pulam a linha.
func MuralDe(caminho string, agora time.Time, max int) []Aviso {
	if os.Getenv("FANTASMAS_MURAL") == "off" || max <= 0 {
		return nil
	}
	f, err := os.Open(caminho)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []Aviso
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var raw struct {
			At    string `json:"at"`
			Autor string `json:"autor"`
			Texto string `json:"texto"`
			Ate   string `json:"ate"`
		}
		if json.Unmarshal(sc.Bytes(), &raw) != nil || raw.Texto == "" {
			continue
		}
		at, err := time.Parse(time.RFC3339, raw.At)
		if err != nil || agora.Sub(at) > janelaMural {
			continue
		}
		a := Aviso{At: at, Autor: raw.Autor, Texto: raw.Texto}
		if raw.Ate != "" {
			ate, err := time.Parse(time.RFC3339, raw.Ate)
			if err != nil {
				ate, err = time.Parse("2006-01-02T15:04Z07:00", raw.Ate)
			}
			if err == nil {
				if !ate.After(agora) {
					continue
				}
				a.Ate = ate
			}
		}
		out = append(out, a)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if len(out) > max {
		out = out[:max]
	}
	return out
}
