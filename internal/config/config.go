// Package config guarda os ajustes do fantasmas: as variáveis FANTASMAS_* escolhidas na aba configurações.
//
// O arquivo fica fora do repo (~/.config/fantasmas/ajustes.json, permissão 0600) e é aplicado ao
// ambiente do processo na abertura, antes de qualquer os.Getenv. O ambiente real sempre manda:
// uma chave já definida no shell não é sobrescrita nem editável pela aba.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Opcao é um ajuste da aba configurações.
type Opcao struct {
	Chave     string
	Grupo     string
	Descricao string
	// Valores são as escolhas fixas; vazio = texto livre, que a aba nunca mostra.
	Valores []string
	// Reabrir indica que o valor só é lido na abertura do fantasmas.
	Reabrir bool
	// Padrao é o valor que vale sem nada definido, como aparece na tela (pode ser um dos Valores).
	Padrao string
}

// Livre diz se o valor é texto livre (mascarado na tela).
func (o Opcao) Livre() bool { return len(o.Valores) == 0 }

// ChaveTema escolhe o tema das cores (roxo, orca ou terminal).
const ChaveTema = "FANTASMAS_TEMA"

// ChaveFigura escolhe a figura do topo; as escolhas vêm do catálogo (DefinirValores).
const ChaveFigura = "FANTASMAS_FIGURA"

// Opcoes, na ordem da aba. Padrao é o valor que vale sem nada definido, como aparece na tela.
var Opcoes = []Opcao{
	{ChaveTema, "aparência", "cores do app", []string{"roxo", "orca", "terminal"}, false, "roxo"},
	{ChaveFigura, "aparência", "figura do topo; GIF ou PNG na pasta figuras", []string{"gengar"}, false, "gengar"},
	{"FANTASMAS_RAIZ", "variáveis de ambiente", "raiz do workspace Spark (aceita ~/); --raiz manda", nil, true, "pasta atual"},
	{"FANTASMAS_FONTE", "variáveis de ambiente", "ddb liga as gavetas missões e pipes", []string{"ddb"}, false, "delegate.py"},
	{"FANTASMAS_DDB_SCRIPT", "variáveis de ambiente", "caminho do ddb.py, relativo à raiz do Spark", nil, true, "scripts/delegate/ddb.py"},
	{"FANTASMAS_MURAL", "variáveis de ambiente", "off esconde o mural", []string{"off"}, false, "ligado"},
	{"FANTASMAS_CLAUDE_DIR", "variáveis de ambiente", "pasta do Claude (transcripts)", nil, false, "~/.claude"},
	{"FANTASMAS_CACHE_DIR", "variáveis de ambiente", "pasta de cache do fantasmas", nil, false, padraoCache()},
	{"FANTASMAS_EDITOR", "variáveis de ambiente", "editor de arquivos e pastas", nil, false, padraoEditor()},
	{"FANTASMAS_CTX_JANELA", "variáveis de ambiente", "janela de contexto usada no %", nil, false, "1000000"},
	{"FANTASMAS_DEBUG", "variáveis de ambiente", "arquivo de log do terminal e da imagem", nil, true, "desligado"},
}

// padraoCache é a pasta que dados.dirCache usa sem FANTASMAS_CACHE_DIR, com ~ no lugar da home.
func padraoCache() string {
	d, err := os.UserCacheDir()
	if err != nil {
		return "nenhuma"
	}
	d = filepath.Join(d, "fantasmas")
	if h, err := os.UserHomeDir(); err == nil && strings.HasPrefix(d, h+string(os.PathSeparator)) {
		d = "~" + d[len(h):]
	}
	return d
}

// padraoEditor é o abridor do sistema que main.comandoAbrir usa sem FANTASMAS_EDITOR.
func padraoEditor() string {
	switch runtime.GOOS {
	case "darwin":
		return "open"
	case "linux":
		return "xdg-open"
	}
	return "start"
}

// DefinirValores troca as escolhas de uma opção (o catálogo de figuras é lido na abertura).
func DefinirValores(chave string, valores []string) {
	for i := range Opcoes {
		if Opcoes[i].Chave == chave {
			Opcoes[i].Valores = valores
		}
	}
}

// Achar devolve a opção da chave.
func Achar(chave string) (Opcao, bool) {
	for _, o := range Opcoes {
		if o.Chave == chave {
			return o, true
		}
	}
	return Opcao{}, false
}

// Origem diz de onde vem o valor efetivo de uma chave.
type Origem int

const (
	Padrao   Origem = iota // nada definido
	Arquivo                // salvo pela aba configurações
	Ambiente               // definido no shell; a aba não mexe
)

// Caminho é FANTASMAS_AJUSTES ou <config do usuário>/fantasmas/ajustes.json.
func Caminho() (string, error) {
	if p := os.Getenv("FANTASMAS_AJUSTES"); p != "" {
		return p, nil
	}
	d, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("sem pasta de configuração: %w", err)
	}
	return filepath.Join(d, "fantasmas", "ajustes.json"), nil
}

// Loja guarda os ajustes do arquivo e sabe quais chaves vieram do shell.
type Loja struct {
	caminho  string
	valores  map[string]string
	ambiente map[string]bool
}

type arquivo struct {
	Env map[string]string `json:"env"`
}

// Abrir lê o arquivo (ausente = sem ajustes) e aplica ao ambiente do processo as chaves que o shell
// não definiu. Com arquivo ilegível devolve a loja vazia, ainda utilizável, e o erro.
func Abrir(caminho string) (*Loja, error) {
	l := &Loja{caminho: caminho, valores: map[string]string{}, ambiente: map[string]bool{}}
	for _, o := range Opcoes {
		if _, ok := os.LookupEnv(o.Chave); ok {
			l.ambiente[o.Chave] = true
		}
	}
	b, err := os.ReadFile(caminho)
	if errors.Is(err, fs.ErrNotExist) {
		return l, nil
	}
	if err != nil {
		return l, fmt.Errorf("ajustes: %w", err)
	}
	var a arquivo
	if err := json.Unmarshal(b, &a); err != nil {
		return l, fmt.Errorf("ajustes ilegíveis em %s: %w", caminho, err)
	}
	for k, v := range a.Env {
		if _, ok := Achar(k); ok && v != "" {
			l.valores[k] = v
		}
	}
	for k, v := range l.valores {
		if !l.ambiente[k] {
			os.Setenv(k, v)
		}
	}
	return l, nil
}

// Caminho do arquivo desta loja.
func (l *Loja) Caminho() string { return l.caminho }

// Origem da chave.
func (l *Loja) Origem(chave string) Origem {
	switch {
	case l.ambiente[chave]:
		return Ambiente
	case l.valores[chave] != "":
		return Arquivo
	}
	return Padrao
}

// Valor efetivo da chave. A aba só o mostra para opções com Valores fixos.
func (l *Loja) Valor(chave string) string { return os.Getenv(chave) }

// Definir salva o valor ("" volta ao padrão) e já o aplica ao processo.
func (l *Loja) Definir(chave, valor string) error {
	o, ok := Achar(chave)
	if !ok {
		return fmt.Errorf("ajuste desconhecido: %s", chave)
	}
	if l.ambiente[chave] {
		return fmt.Errorf("%s vem do shell; tire de lá para mudar aqui", chave)
	}
	if valor != "" && !o.Livre() && !contem(o.Valores, valor) {
		return fmt.Errorf("%s não aceita esse valor", chave)
	}
	novo := make(map[string]string, len(l.valores)+1)
	for k, v := range l.valores {
		novo[k] = v
	}
	if valor == "" {
		delete(novo, chave)
	} else {
		novo[chave] = valor
	}
	if err := gravar(l.caminho, novo); err != nil {
		return err
	}
	l.valores = novo
	if valor == "" {
		return os.Unsetenv(chave)
	}
	return os.Setenv(chave, valor)
}

// gravar escreve o arquivo com permissão 0600 numa pasta 0700, trocando-o atomicamente.
func gravar(caminho string, valores map[string]string) error {
	dir := filepath.Dir(caminho)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("ajustes: %w", err)
	}
	b, err := json.MarshalIndent(arquivo{Env: valores}, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".ajustes-*")
	if err != nil {
		return fmt.Errorf("ajustes: %w", err)
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return fmt.Errorf("ajustes: %w", err)
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		f.Close()
		return fmt.Errorf("ajustes: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("ajustes: %w", err)
	}
	if err := os.Rename(tmp, caminho); err != nil {
		return fmt.Errorf("ajustes: %w", err)
	}
	return nil
}

func contem(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
