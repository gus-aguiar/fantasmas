package dados

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// VivosDe lê o JSON do `herdr agent list`; nil quando a saída não é legível.
func VivosDe(saida []byte) map[string]bool {
	ags := AgentesDe(saida)
	if ags == nil {
		return nil
	}
	v := map[string]bool{}
	for n := range ags {
		v[n] = true
	}
	return v
}

// Vivo: sem leitura do Herdr, ninguém é dado como morto.
func Vivo(e Estado, nome string) bool { return e.Vivos == nil || e.Vivos[nome] }

func crewDe(p Linha) string {
	crew, _, _ := strings.Cut(p.Rotulo, " · ")
	return crew
}

// Arquivar grava a resposta "(arquivado)" no answers.json sem entregar a ninguém.
func Arquivar(chave string) error {
	arquivo, id := Separar(chave)
	if id == "" {
		return os.ErrInvalid
	}
	p := filepath.Join(filepath.Dir(arquivo), "answers.json")
	a := map[string]json.RawMessage{}
	if b, err := os.ReadFile(p); err == nil {
		if err := json.Unmarshal(b, &a); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	v, _ := json.Marshal(map[string]string{"answer": "(arquivado)", "at": time.Now().Format("2006-01-02T15:04"), "via": "arquivado"})
	a[id] = v
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return escreveAtomico(p, append(b, '\n'))
}

func escreveAtomico(p string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(p), ".tmp-"+filepath.Base(p)+"-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if _, err := f.Write(b); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func dirCache() string {
	if d := os.Getenv("FANTASMAS_CACHE_DIR"); d != "" {
		return d
	}
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "fantasmas")
	}
	return ""
}

// ChaveVisto = pasta + mtime do relatório; vazio quando não há relatório.
func ChaveVisto(pasta string) string {
	for _, n := range []string{"report.json", "report.md"} {
		if st, err := os.Stat(filepath.Join(pasta, n)); err == nil {
			return pasta + "@" + st.ModTime().UTC().Format(time.RFC3339Nano)
		}
	}
	return ""
}

func CarregarVistos() map[string]bool {
	v := map[string]bool{}
	d := dirCache()
	if d == "" {
		return v
	}
	b, err := os.ReadFile(filepath.Join(d, "vistos.json"))
	if err != nil {
		return v
	}
	var l []string
	if json.Unmarshal(b, &l) == nil {
		for _, k := range l {
			v[k] = true
		}
	}
	return v
}

// MarcarVisto devolve a chave gravada; erro de cache não impede o uso em memória.
func MarcarVisto(pasta string) (string, error) {
	k := ChaveVisto(pasta)
	if k == "" {
		return "", nil
	}
	d := dirCache()
	if d == "" {
		return k, os.ErrNotExist
	}
	v := CarregarVistos()
	v[k] = true
	l := make([]string, 0, len(v))
	for x := range v {
		l = append(l, x)
	}
	b, _ := json.Marshal(l)
	if err := os.MkdirAll(d, 0o755); err != nil {
		return k, err
	}
	return k, escreveAtomico(filepath.Join(d, "vistos.json"), b)
}
