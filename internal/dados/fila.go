package dados

import (
	"os"
	"path/filepath"
)

// Recomendacao é o que o agente recomenda numa pergunta, para a fila de decisões da gaveta agora.
type Recomendacao struct {
	// Opcao é o id da opção recomendada; vazio quando a recomendação é só texto (não dá para aprovar com uma tecla).
	Opcao string
	// Texto é o rótulo da opção recomendada ou, sem opção, o texto da recomendação.
	Texto string
	// Regra é a regra do vision que motivou a dúvida; vazia = fora do vision.
	Regra string
}

// Aprovavel diz se a recomendação aponta para uma opção que se responde sem texto livre.
func (r Recomendacao) Aprovavel() bool { return r.Opcao != "" }

// RecomendacaoDe resolve a recomendação de uma pergunta aberta: o id vira o rótulo da opção.
func RecomendacaoDe(a *Aberto) Recomendacao {
	r := Recomendacao{Texto: a.Recomendacao, Regra: a.Vision}
	for _, o := range a.Opcoes {
		if a.Recomendacao != "" && (o.ID == a.Recomendacao || o.Label == a.Recomendacao) {
			if !o.Livre {
				r.Opcao = o.ID
				if r.Opcao == "" {
					r.Opcao = o.Label
				}
			}
			r.Texto = o.Label
			return r
		}
	}
	return r
}

// LerRecomendacoes lê o questions.json de cada pergunta pendente e devolve a recomendação por chave.
// Pergunta sem arquivo legível fica de fora; a fila mostra só o que a pergunta tem.
func LerRecomendacoes(raiz string, esperando []Linha, ler func(string) ([]byte, error)) map[string]Recomendacao {
	if ler == nil {
		ler = os.ReadFile
	}
	out := map[string]Recomendacao{}
	cache := map[string][]byte{}
	for _, p := range esperando {
		arquivo, id := Separar(p.Chave)
		if id == "" {
			continue
		}
		if !filepath.IsAbs(arquivo) && raiz != "" {
			arquivo = filepath.Join(raiz, arquivo)
		}
		b, ok := cache[arquivo]
		if !ok {
			b, _ = ler(arquivo)
			cache[arquivo] = b
		}
		if b == nil {
			continue
		}
		if a, err := PerguntaDe(p.Chave, b, id); err == nil && a != nil {
			out[p.Chave] = RecomendacaoDe(a)
		}
	}
	return out
}
