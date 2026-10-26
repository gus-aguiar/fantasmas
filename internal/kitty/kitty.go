// Package kitty gera as sequências do protocolo de gráficos do kitty.
package kitty

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const bloco = 4096

func apc(s string) string { return "\x1b_G" + s + "\x1b\\" }

// Transmitir manda os dados em base64, em blocos de até 4096 com m=1 e o último m=0.
func Transmitir(chaves string, dados []byte) string {
	b64 := base64.StdEncoding.EncodeToString(dados)
	var b strings.Builder
	for primeiro := true; primeiro || b64 != ""; primeiro = false {
		n := min(bloco, len(b64))
		carga := b64[:n]
		b64 = b64[n:]
		m := 0
		if b64 != "" {
			m = 1
		}
		if primeiro {
			b.WriteString(apc(fmt.Sprintf("%s,m=%d;%s", chaves, m, carga)))
		} else {
			b.WriteString(apc(fmt.Sprintf("m=%d;%s", m, carga)))
		}
	}
	return b.String()
}

// Animacao transmite os quadros PNG como uma imagem animada (a=f) e liga o loop.
func Animacao(id int, quadros [][]byte, delays []int) string {
	var b strings.Builder
	for i, q := range quadros {
		z := 100
		if i < len(delays) {
			z = delays[i]
		}
		if i == 0 {
			b.WriteString(Transmitir(fmt.Sprintf("a=t,f=100,i=%d,q=2", id), q))
			b.WriteString(apc(fmt.Sprintf("a=a,i=%d,r=1,z=%d,q=2", id, z)))
			continue
		}
		b.WriteString(Transmitir(fmt.Sprintf("a=f,f=100,i=%d,z=%d,q=2", id, z), q))
	}
	b.WriteString(apc(fmt.Sprintf("a=a,i=%d,s=3,v=1,q=2", id)))
	return b.String()
}

// Quadros transmite cada quadro como imagem própria, ids base, base+1, ...
func Quadros(base int, quadros [][]byte) string {
	var b strings.Builder
	for i, q := range quadros {
		b.WriteString(Transmitir(fmt.Sprintf("a=t,f=100,i=%d,q=2", base+i), q))
	}
	return b.String()
}

// Colocar posiciona a imagem na célula (x, y) base 0 sem mover o cursor; o mesmo p substitui a colocação anterior.
func Colocar(id, p, x, y, c, r int) string {
	return fmt.Sprintf("\x1b7\x1b[%d;%dH", y+1, x+1) + apc(fmt.Sprintf("a=p,i=%d,p=%d,c=%d,r=%d,C=1,q=2", id, p, c, r)) + "\x1b8"
}

// TransmitirColocar manda e coloca o PNG de uma vez (a=T) na célula (x, y) base 0, sem mover o cursor.
func TransmitirColocar(id, p, x, y, c, r int, png []byte) string {
	return fmt.Sprintf("\x1b7\x1b[%d;%dH", y+1, x+1) + Transmitir(fmt.Sprintf("a=T,f=100,i=%d,p=%d,c=%d,r=%d,q=2,C=1", id, p, c, r), png) + "\x1b8"
}

// Apagar tira as colocações da imagem e mantém os dados.
func Apagar(id int) string { return apc(fmt.Sprintf("a=d,d=i,i=%d,q=2", id)) }

// Liberar tira as colocações e os dados.
func Liberar(id int) string { return apc(fmt.Sprintf("a=d,d=I,i=%d,q=2", id)) }

// Consulta pergunta pelo suporte (i=31) e emenda um DA1, que todo terminal responde.
func Consulta() string { return apc("i=31,s=1,v=1,a=q,t=d,f=24;AAAA") + "\x1b[c" }

// Resposta: decidido quando chegou o OK ou o DA1; sim só com o OK.
func Resposta(b []byte) (decidido, sim bool) {
	if bytes.Contains(b, []byte("\x1b_Gi=31;OK")) {
		return true, true
	}
	if i := bytes.Index(b, []byte("\x1b[?")); i >= 0 && bytes.IndexByte(b[i:], 'c') > 0 {
		return true, false
	}
	return false, false
}

type leitorComPrazo interface {
	io.Reader
	SetReadDeadline(time.Time) error
}

// Detectar escreve a consulta e lê a resposta até o prazo.
func Detectar(w io.Writer, r leitorComPrazo, prazo time.Duration) bool {
	if _, err := io.WriteString(w, Consulta()); err != nil {
		return false
	}
	fim := time.Now().Add(prazo)
	if r.SetReadDeadline(fim) != nil {
		return false
	}
	defer r.SetReadDeadline(time.Time{})
	var acc []byte
	buf := make([]byte, 256)
	for time.Now().Before(fim) {
		n, err := r.Read(buf)
		acc = append(acc, buf[:n]...)
		// Com o OK, segue lendo até o DA1 para não vazar a resposta como tecla.
		if d, sim := Resposta(acc); d && (!sim || temDA1(acc)) {
			return sim
		}
		if err != nil {
			_, sim := Resposta(acc)
			return sim
		}
	}
	_, sim := Resposta(acc)
	return sim
}

func temDA1(b []byte) bool {
	i := bytes.Index(b, []byte("\x1b[?"))
	return i >= 0 && bytes.IndexByte(b[i:], 'c') > 0
}

// Saida serializa as escritas do Bubble Tea e das imagens no mesmo terminal.
type Saida struct {
	mu sync.Mutex
	*os.File
}

func (s *Saida) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.File.Write(b)
}
