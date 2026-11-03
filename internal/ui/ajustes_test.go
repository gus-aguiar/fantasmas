package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fantasmas/internal/config"
	"fantasmas/internal/dados"

	tea "github.com/charmbracelet/bubbletea"
)

// modeloAjustes abre uma loja num arquivo temporário; o ambiente fica limpo e é restaurado no fim.
// shell são as chaves que o "shell" já define antes da abertura.
func modeloAjustes(t *testing.T, shell map[string]string) (Model, *config.Loja, string) {
	t.Helper()
	for _, o := range config.Opcoes {
		t.Setenv(o.Chave, "")
		os.Unsetenv(o.Chave)
	}
	for k, v := range shell {
		t.Setenv(k, v)
	}
	t.Cleanup(func() { UsarTema("roxo") })
	arq := filepath.Join(t.TempDir(), "fantasmas", "ajustes.json")
	l, err := config.Abrir(arq)
	if err != nil {
		t.Fatal(err)
	}
	m := modelo(fixture(), &registro{})
	m.acoes.Ajustes = l
	m.largura, m.altura = 110, 50
	return tecla(m, ","), l, arq
}

// selAjuste é a posição do ajuste entre os itens da aba (os botões de ações vêm antes).
func selAjuste(t *testing.T, m Model, chave string) int {
	t.Helper()
	for i, a := range m.alvos() {
		if a.ajuste == chave {
			return i
		}
	}
	t.Fatalf("sem %s", chave)
	return 0
}

func TestVirgulaAbreSettings(t *testing.T) {
	m, _, _ := modeloAjustes(t, nil)
	if Gavetas[m.gaveta].ID != "settings" {
		t.Fatalf("gaveta %s", Gavetas[m.gaveta].ID)
	}
	out := render(m, 110, 50)
	for _, s := range []string{"aparência", "variáveis de ambiente", "FANTASMAS_TEMA", "FANTASMAS_EDITOR", "FANTASMAS_CTX_JANELA",
		"1000000", "~/.claude", "delegate.py", "vale ao reabrir"} {
		if !strings.Contains(out, s) {
			t.Fatalf("settings sem %q:\n%s", s, out)
		}
	}
	for _, s := range []string{"FANTASMAS_IMAGEM", "FANTASMAS_SPRITE", "padrão "} {
		if strings.Contains(out, s) {
			t.Fatalf("settings não deveria ter %q:\n%s", s, out)
		}
	}
}

func TestEnterTrocaOTemaESalva(t *testing.T) {
	m, l, arq := modeloAjustes(t, nil)
	m.sel = selAjuste(t, m, config.ChaveTema)
	m = tecla(m, "enter")
	if T.Base != temaOrca.Base || l.Valor(config.ChaveTema) != "orca" {
		t.Fatalf("primeiro enter deveria ir ao orca: base %q valor %q", T.Base, l.Valor(config.ChaveTema))
	}
	if b, _ := os.ReadFile(arq); !strings.Contains(string(b), `"FANTASMAS_TEMA": "orca"`) {
		t.Fatalf("arquivo: %s", b)
	}
	m = tecla(m, "enter")
	if T.Base != "" || l.Valor(config.ChaveTema) != "terminal" {
		t.Fatalf("segundo enter deveria ir ao terminal: base %q", T.Base)
	}
	m = tecla(m, "enter")
	if T.Base != temaRoxo.Base || l.Origem(config.ChaveTema) != config.Padrao {
		t.Fatalf("terceiro enter deveria voltar ao roxo padrão: base %q", T.Base)
	}
	if !strings.Contains(render(m, 110, 50), "FANTASMAS_TEMA: roxo") {
		t.Fatal("toast sem o tema")
	}
}

func TestTextoLivreNuncaAparece(t *testing.T) {
	const segredo = "segredo-xyz-123"
	m, l, _ := modeloAjustes(t, nil)
	m.sel = selAjuste(t, m, "FANTASMAS_EDITOR")
	m = tecla(m, "enter")
	if m.ajuste == nil {
		t.Fatal("enter no texto livre deveria abrir a barra")
	}
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(segredo)})
	if out := render(m, 110, 50); strings.Contains(out, segredo) || !strings.Contains(out, "•••") {
		t.Fatalf("barra mostrou o valor ou sem máscara:\n%s", out)
	}
	m = tecla(m, "enter")
	if m.ajuste != nil || l.Valor("FANTASMAS_EDITOR") != segredo || os.Getenv("FANTASMAS_EDITOR") != segredo {
		t.Fatal("enter deveria salvar e aplicar")
	}
	out := render(m, 110, 50)
	if strings.Contains(out, segredo) || !strings.Contains(out, Mascara) {
		t.Fatalf("lista ou toast mostrou o valor:\n%s", out)
	}

	// reabrir a barra nunca traz o valor salvo; vazio volta ao padrão
	m = tecla(m, "enter")
	if m.ajuste.campo.Value() != "" {
		t.Fatal("barra veio preenchida")
	}
	m = tecla(m, "enter")
	if l.Origem("FANTASMAS_EDITOR") != config.Padrao {
		t.Fatal("vazio deveria voltar ao padrão")
	}
}

func TestEscCancelaEdicao(t *testing.T) {
	m, l, _ := modeloAjustes(t, nil)
	m.sel = selAjuste(t, m, "FANTASMAS_DEBUG")
	m = tecla(m, "enter")
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/tmp/x.log")})
	m = teclaMsg(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.ajuste != nil || l.Origem("FANTASMAS_DEBUG") != config.Padrao {
		t.Fatal("esc deveria cancelar sem salvar")
	}
}

func TestShellNaoEditavel(t *testing.T) {
	m, l, _ := modeloAjustes(t, map[string]string{"FANTASMAS_MURAL": "off", "FANTASMAS_CLAUDE_DIR": "/segredo/claude"})
	m.sel = selAjuste(t, m, "FANTASMAS_MURAL")
	m = tecla(m, "enter")
	out := render(m, 110, 50)
	if !strings.Contains(out, "vem do shell") || l.Origem("FANTASMAS_MURAL") != config.Ambiente {
		t.Fatalf("deveria avisar que vem do shell:\n%s", out)
	}
	if strings.Contains(out, "/segredo/claude") {
		t.Fatal("valor do shell apareceu na tela")
	}
}

func TestSettingsSemLoja(t *testing.T) {
	m := tecla(modelo(dados.Estado{}, &registro{}), ",")
	if !strings.Contains(render(m, 96, 40), "settings indisponível") {
		t.Fatalf("sem aviso de loja ausente:\n%s", render(m, 96, 40))
	}
}
