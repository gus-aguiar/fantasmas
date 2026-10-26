package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entrega struct {
	Slug, Crew, Titulo, Resposta string
	Entregue                     bool
	Motivo                       string
}

const timeoutResposta = 30 * time.Second

func Responder(ctx context.Context, r Runner, raiz string, rec Recursos, chave, resposta string) (Entrega, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutResposta)
		defer cancel()
	}
	if rec.Painel != "" {
		return respondePeloPainel(ctx, r, raiz, rec, chave, resposta)
	}
	return respondeSemPainel(ctx, r, raiz, rec, chave, resposta)
}

func respondePeloPainel(ctx context.Context, r Runner, raiz string, rec Recursos, chave, resposta string) (Entrega, error) {
	out, err := r.Run(ctx, "python3", rec.Painel, "responder", raiz, chave, resposta)
	if err != nil {
		return Entrega{}, err
	}
	cols := strings.Split(strings.TrimRight(out, "\r\n"), "\t")
	if len(cols) < 5 {
		return Entrega{}, fmt.Errorf("saída inesperada do painel: %q", strings.TrimSpace(out))
	}
	e := Entrega{Slug: cols[0], Crew: cols[1], Titulo: cols[2], Resposta: cols[3]}
	e.Entregue, e.Motivo = veredito(cols[4])
	return e, nil
}

func veredito(s string) (bool, string) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "sim") {
		return true, ""
	}
	return false, strings.TrimPrefix(s, "não: ")
}

func respondeSemPainel(ctx context.Context, r Runner, raiz string, rec Recursos, chave, resposta string) (Entrega, error) {
	arquivo, id := Separar(chave)
	if id == "report" && !strings.Contains(arquivo, "/") {
		if resposta != "Lido" {
			return Entrega{}, fmt.Errorf("sem painel.py o pedido do relatório não chega ao %s; mande pelo coordenador", arquivo)
		}
		return Entrega{Crew: arquivo, Titulo: "Relatório", Resposta: resposta, Motivo: "relatório só lido, nada a entregar"}, nil
	}
	if id == "" {
		return Entrega{}, fmt.Errorf("chave sem id de pergunta: %q", chave)
	}
	script := filepath.Join(raiz, "shared", "mods", "control-panel", "scripts", "delegate.py")
	out, err := r.Run(ctx, "python3", script, "responder", arquivo, id, resposta)
	if err != nil {
		return Entrega{}, err
	}
	cols := strings.Split(strings.TrimRight(out, "\r\n"), "\t")
	if len(cols) < 4 {
		return Entrega{}, fmt.Errorf("saída inesperada do delegate.py: %q", strings.TrimSpace(out))
	}
	e := Entrega{Slug: cols[0], Crew: cols[1], Titulo: cols[2], Resposta: cols[3]}

	pasta := filepath.Dir(arquivo)
	destino := e.Crew
	if e.Crew == "coordenador" {
		destino = coordenadorDe(pasta)
	}
	switch {
	case !rec.Herdr:
		e.Motivo = "herdr não encontrado"
	case destino == "":
		e.Motivo = "sem sessão de destino (o coordenador lê o answers.json)"
	default:
		msg := fmt.Sprintf("decision: %s: %s", id, resposta)
		if kindDe(arquivo, id) == "action" && resposta == "Feito" {
			msg = fmt.Sprintf("info: %s: Feito", id)
		}
		if _, err := r.Run(ctx, "herdr", "agent", "prompt", destino, msg); err != nil {
			e.Motivo = "herdr falhou (" + resumoErro(err) + ")"
		} else {
			e.Entregue = true
		}
	}
	return e, nil
}

func resumoErro(err error) string {
	m := strings.TrimSpace(err.Error())
	if len(m) > 120 {
		m = m[:120]
	}
	return m
}

func coordenadorDe(pasta string) string {
	b, err := os.ReadFile(filepath.Join(pasta, "meta"))
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(l), "="); ok && k == "coordinator" {
			return v
		}
	}
	return ""
}

func kindDe(arquivo, id string) string {
	b, err := os.ReadFile(arquivo)
	if err != nil {
		return ""
	}
	var q struct {
		Items []struct {
			ID   any    `json:"id"`
			Kind string `json:"kind"`
		} `json:"items"`
	}
	if json.Unmarshal(b, &q) != nil {
		return ""
	}
	for _, i := range q.Items {
		if fmt.Sprint(i.ID) == id {
			return i.Kind
		}
	}
	return ""
}
