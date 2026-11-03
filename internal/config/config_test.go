package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// limpo zera as chaves do catálogo no processo e as restaura no fim do teste.
func limpo(t *testing.T) {
	t.Helper()
	for _, o := range Opcoes {
		t.Setenv(o.Chave, "")
		os.Unsetenv(o.Chave)
	}
}

func TestAbrirSemArquivo(t *testing.T) {
	limpo(t)
	l, err := Abrir(filepath.Join(t.TempDir(), "nao", "existe.json"))
	if err != nil {
		t.Fatal(err)
	}
	if l.Origem(ChaveTema) != Padrao || l.Valor(ChaveTema) != "" {
		t.Fatalf("origem %v valor %q", l.Origem(ChaveTema), l.Valor(ChaveTema))
	}
}

func TestDefinirGravaAplicaERelê(t *testing.T) {
	limpo(t)
	arq := filepath.Join(t.TempDir(), "fantasmas", "ajustes.json")
	l, _ := Abrir(arq)
	if err := l.Definir("FANTASMAS_EDITOR", "code --wait"); err != nil {
		t.Fatal(err)
	}
	if err := l.Definir(ChaveTema, "orca"); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("FANTASMAS_EDITOR") != "code --wait" || os.Getenv(ChaveTema) != "orca" {
		t.Fatal("Definir não aplicou ao processo")
	}
	st, err := os.Stat(arq)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("arquivo com permissão %o, quer 600", st.Mode().Perm())
	}
	if d, _ := os.Stat(filepath.Dir(arq)); d.Mode().Perm() != 0o700 {
		t.Fatalf("pasta com permissão %o, quer 700", d.Mode().Perm())
	}
	if sobra, _ := filepath.Glob(filepath.Join(filepath.Dir(arq), ".ajustes-*")); len(sobra) > 0 {
		t.Fatalf("temporário ficou para trás: %v", sobra)
	}

	limpo(t)
	l2, err := Abrir(arq)
	if err != nil {
		t.Fatal(err)
	}
	if l2.Origem("FANTASMAS_EDITOR") != Arquivo || os.Getenv("FANTASMAS_EDITOR") != "code --wait" {
		t.Fatal("reabrir não aplicou o arquivo")
	}
	if err := l2.Definir("FANTASMAS_EDITOR", ""); err != nil {
		t.Fatal(err)
	}
	if _, ok := os.LookupEnv("FANTASMAS_EDITOR"); ok || l2.Origem("FANTASMAS_EDITOR") != Padrao {
		t.Fatal("vazio deveria voltar ao padrão")
	}
	b, _ := os.ReadFile(arq)
	if strings.Contains(string(b), "EDITOR") || !strings.Contains(string(b), "orca") {
		t.Fatalf("arquivo: %s", b)
	}
}

func TestShellMandaSobreOArquivo(t *testing.T) {
	limpo(t)
	arq := filepath.Join(t.TempDir(), "ajustes.json")
	os.WriteFile(arq, []byte(`{"env":{"FANTASMAS_MURAL":"off","FANTASMAS_TEMA":"orca"}}`), 0o600)
	t.Setenv("FANTASMAS_MURAL", "")
	l, err := Abrir(arq)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("FANTASMAS_MURAL") != "" || l.Origem("FANTASMAS_MURAL") != Ambiente {
		t.Fatal("o shell deveria mandar, mesmo com valor vazio")
	}
	if err := l.Definir("FANTASMAS_MURAL", "off"); err == nil {
		t.Fatal("chave do shell não deveria ser editável")
	}
	if os.Getenv(ChaveTema) != "orca" {
		t.Fatal("chave fora do shell deveria vir do arquivo")
	}
}

func TestDefinirRecusa(t *testing.T) {
	limpo(t)
	l, _ := Abrir(filepath.Join(t.TempDir(), "a.json"))
	if l.Definir("FANTASMAS_NAO_EXISTE", "x") == nil {
		t.Fatal("chave desconhecida aceita")
	}
	if l.Definir(ChaveTema, "verde") == nil {
		t.Fatal("valor fora da lista aceito")
	}
}

func TestArquivoIlegivel(t *testing.T) {
	limpo(t)
	arq := filepath.Join(t.TempDir(), "a.json")
	os.WriteFile(arq, []byte("{"), 0o600)
	l, err := Abrir(arq)
	if err == nil || l == nil {
		t.Fatalf("quer loja vazia e erro, veio %v %v", l, err)
	}
	if err := l.Definir(ChaveTema, "orca"); err != nil {
		t.Fatalf("loja vazia deveria seguir utilizável: %v", err)
	}
}

func TestCatalogo(t *testing.T) {
	vistas := map[string]bool{}
	for _, o := range Opcoes {
		if !strings.HasPrefix(o.Chave, "FANTASMAS_") || o.Grupo == "" || o.Descricao == "" {
			t.Errorf("opção incompleta: %+v", o)
		}
		if vistas[o.Chave] {
			t.Errorf("chave repetida: %s", o.Chave)
		}
		vistas[o.Chave] = true
	}
}
