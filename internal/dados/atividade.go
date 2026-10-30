package dados

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const maxCauda = 64 << 10

// Atividade é o que o fantasminha fez por último, lido da cauda do transcript.
type Atividade struct {
	Acao   string
	Quando time.Time
	Tokens int
	Modelo string
}

type linhaTranscript struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Model string `json:"model"`
		Usage *struct {
			Input         int `json:"input_tokens"`
			CacheRead     int `json:"cache_read_input_tokens"`
			CacheCreation int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type blocoConteudo struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input struct {
		Description string `json:"description"`
		FilePath    string `json:"file_path"`
		Skill       string `json:"skill"`
	} `json:"input"`
}

// AtividadeDe devolve a última ação e o último usage; linhas ilegíveis são ignoradas.
func AtividadeDe(cauda []byte) Atividade {
	var a Atividade
	r := bufio.NewReader(bytes.NewReader(cauda))
	for {
		linha, err := r.ReadBytes(10)
		if len(linha) == 0 && err != nil {
			break
		}
		var l linhaTranscript
		if json.Unmarshal(linha, &l) != nil || l.Type != "assistant" {
			continue
		}
		quando, _ := time.Parse(time.RFC3339Nano, l.Timestamp)
		if u := l.Message.Usage; u != nil {
			a.Tokens = u.Input + u.CacheRead + u.CacheCreation
			if l.Message.Model != "" {
				a.Modelo = l.Message.Model
			}
		}
		var blocos []blocoConteudo
		if json.Unmarshal(l.Message.Content, &blocos) != nil {
			continue
		}
		for _, b := range blocos {
			if b.Type == "tool_use" {
				a.Acao = descreverFerramenta(b)
				a.Quando = quando
			}
		}
	}
	return a
}

func descreverFerramenta(b blocoConteudo) string {
	in := b.Input
	if in.Description != "" && b.Name != "Agent" {
		return in.Description
	}
	switch {
	case b.Name == "Agent" && in.Description != "":
		return "agente: " + in.Description
	case b.Name == "Edit" && in.FilePath != "":
		return "edita " + filepath.Base(in.FilePath)
	case b.Name == "Read" && in.FilePath != "":
		return "lê " + filepath.Base(in.FilePath)
	case b.Name == "Skill" && in.Skill != "":
		return "skill " + in.Skill
	}
	return b.Name
}

// LerCauda lê os últimos max bytes e descarta a primeira linha se cortou no meio.
func LerCauda(caminho string, max int64) []byte {
	f, err := os.Open(caminho)
	if err != nil {
		return nil
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil
	}
	inicio := info.Size() - max
	if inicio < 0 {
		inicio = 0
	}
	// com inicio > 0 lê 1 byte a mais para saber se o corte caiu no começo de uma linha
	leitura := inicio
	if inicio > 0 {
		leitura--
	}
	if _, err := f.Seek(leitura, io.SeekStart); err != nil {
		return nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil
	}
	if inicio > 0 && len(b) > 0 {
		anterior := b[0]
		b = b[1:]
		if anterior != '\n' {
			i := bytes.IndexByte(b, '\n')
			if i < 0 {
				return nil
			}
			b = b[i+1:]
		}
	}
	return b
}

// CtxPct é o uso da janela de contexto em %, ou -1 sem medição.
func CtxPct(a Atividade) int {
	if a.Tokens == 0 {
		return -1
	}
	janela := 1_000_000
	if v, err := strconv.Atoi(os.Getenv("FANTASMAS_CTX_JANELA")); err == nil && v > 0 {
		janela = v
	}
	if p := a.Tokens * 100 / janela; p < 100 {
		return p
	}
	return 100
}
