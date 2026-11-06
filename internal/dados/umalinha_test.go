package dados

import "testing"

func TestUmaLinhaTiraQuebras(t *testing.T) {
	got := UmaLinha("_[🤖 Texto de IA]_\n\nO administrador\t  pediu")
	if got != "_[🤖 Texto de IA]_ O administrador pediu" {
		t.Fatalf("UmaLinha = %q", got)
	}
}
