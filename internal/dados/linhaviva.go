package dados

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type EstadoLinha int

const (
	EsperaVoce EstadoLinha = iota
	Trabalhando
	Parado
	Terminou
)

const limiarParado = 10 * time.Minute

type LinhaViva struct {
	Nome, Pasta, Aba, Etapa, Agora string
	Estado                         EstadoLinha
	CtxPct                         int
	Hora                           time.Time
}

func etapaDe(nome string) string {
	switch {
	case strings.HasPrefix(nome, "c-"):
		return "implementar"
	case strings.HasPrefix(nome, "h-"):
		return "homologar"
	case strings.HasPrefix(nome, "g-"):
		return "gengar"
	}
	return "—"
}

func MontarLinhaViva(nome string, ag Agente, ses *SessaoClaude, at Atividade, pendentes []string, temReport, visto bool, agora time.Time) LinhaViva {
	l := LinhaViva{Nome: nome, Aba: ag.Aba, Etapa: etapaDe(nome), CtxPct: CtxPct(at), Hora: at.Quando}
	ref := at.Quando
	if ref.IsZero() && ses != nil {
		ref = ses.Atualizado
	}
	switch {
	case len(pendentes) > 0:
		l.Estado, l.Agora = EsperaVoce, fmt.Sprintf("%d sua(s): %s", len(pendentes), strings.Join(pendentes, ", "))
	case ag.Status == "working" || (ses != nil && (ses.Status == "busy" || ses.Status == "shell")):
		l.Estado, l.Agora = Trabalhando, at.Acao
	case temReport && !visto:
		l.Estado, l.Agora = Terminou, "entregou relatório"
	case !temReport && !ref.IsZero() && agora.Sub(ref) >= limiarParado:
		l.Estado, l.Agora = Parado, fmt.Sprintf("parado com tarefa há %d min", int(agora.Sub(ref).Minutes()))
	default:
		l.Estado, l.Agora = Trabalhando, at.Acao
	}
	if l.Agora == "" {
		l.Agora = "—"
	}
	return l
}

// LinhasVivas junta agentes do Herdr com meta, sessão e transcript; agente sem meta fica de fora.
func LinhasVivas(e Estado, raiz, claudeDir string, agora time.Time) []LinhaViva {
	metas := MetasDe(raiz)
	sessoes := SessoesClaude(filepath.Join(claudeDir, "sessions"))
	var out []LinhaViva
	for nome, ag := range e.Agentes {
		pasta, ok := metas[nome]
		if !ok {
			continue
		}
		ses := sessaoDe(sessoes, nome, ag.Sessao)
		var at Atividade
		if ses != nil {
			id := ag.Sessao
			if id == "" {
				id = idDe(sessoes, ses)
			}
			at = AtividadeDe(LerCauda(CaminhoTranscript(claudeDir, ses.Cwd, id), maxCauda))
		}
		var pend []string
		for _, p := range e.Esperando {
			if crewDe(p) == nome {
				if _, id := Separar(p.Chave); id != "" {
					pend = append(pend, id)
				}
			}
		}
		k := ChaveVisto(pasta)
		l := MontarLinhaViva(nome, ag, ses, at, pend, k != "", k != "" && e.Vistos[k], agora)
		l.Pasta = pasta
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Estado != out[j].Estado {
			return out[i].Estado < out[j].Estado
		}
		return out[i].Nome < out[j].Nome
	})
	return out
}

func idDe(m map[string]SessaoClaude, s *SessaoClaude) string {
	for id, x := range m {
		if x == *s {
			return id
		}
	}
	return ""
}
