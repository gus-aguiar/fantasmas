package kitty

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"
)

func cargas(s string) []string {
	var out []string
	for _, p := range strings.Split(s, "\x1b\\") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func TestTransmitirEmBlocos(t *testing.T) {
	dados := bytes.Repeat([]byte{7}, 7000)
	s := Transmitir("a=t,f=100,i=5,q=2", dados)
	ps := cargas(s)
	b64 := base64.StdEncoding.EncodeToString(dados)
	if len(ps) != 3 {
		t.Fatalf("esperava 3 blocos, veio %d", len(ps))
	}
	if !strings.HasPrefix(ps[0], "\x1b_Ga=t,f=100,i=5,q=2,m=1;") || !strings.HasPrefix(ps[1], "\x1b_Gm=1;") || !strings.HasPrefix(ps[2], "\x1b_Gm=0;") {
		t.Fatalf("cabeçalhos errados: %q / %q / %q", ps[0][:30], ps[1][:10], ps[2][:10])
	}
	var junto string
	for _, p := range ps {
		carga := p[strings.Index(p, ";")+1:]
		if len(carga) > 4096 {
			t.Fatalf("bloco com %d bytes", len(carga))
		}
		junto += carga
	}
	if junto != b64 {
		t.Fatal("base64 remontado não bate")
	}
	if um := cargas(Transmitir("a=t,i=1", []byte("x"))); len(um) != 1 || !strings.HasPrefix(um[0], "\x1b_Ga=t,i=1,m=0;") {
		t.Fatalf("bloco único: %q", um)
	}
}

func TestAnimacao(t *testing.T) {
	s := Animacao(9, [][]byte{[]byte("a"), []byte("b"), []byte("c")}, []int{100, 200, 300})
	for _, w := range []string{"\x1b_Ga=t,f=100,i=9,q=2,m=0;", "\x1b_Ga=a,i=9,r=1,z=100,q=2\x1b\\", "\x1b_Ga=f,f=100,i=9,z=200,q=2,m=0;", "\x1b_Ga=f,f=100,i=9,z=300,q=2,m=0;", "\x1b_Ga=a,i=9,s=3,v=1,q=2\x1b\\"} {
		if !strings.Contains(s, w) {
			t.Fatalf("falta %q em %q", w, s)
		}
	}
	if strings.Index(s, "s=3") < strings.Index(s, "z=300") {
		t.Fatal("o loop tem de vir depois do último quadro")
	}
}

func TestQuadrosSeparados(t *testing.T) {
	s := Quadros(40, [][]byte{[]byte("a"), []byte("b")})
	if !strings.Contains(s, "a=t,f=100,i=40,q=2") || !strings.Contains(s, "a=t,f=100,i=41,q=2") {
		t.Fatalf("ids por quadro: %q", s)
	}
}

func TestColocarEApagar(t *testing.T) {
	if g := Colocar(9, 1, 4, 2, 16, 8); g != "\x1b7\x1b[3;5H\x1b_Ga=p,i=9,p=1,c=16,r=8,C=1,q=2\x1b\\\x1b8" {
		t.Fatalf("colocar: %q", g)
	}
	if g := Apagar(9); g != "\x1b_Ga=d,d=i,i=9,q=2\x1b\\" {
		t.Fatalf("apagar: %q", g)
	}
	if g := Liberar(9); g != "\x1b_Ga=d,d=I,i=9,q=2\x1b\\" {
		t.Fatalf("liberar: %q", g)
	}
}

func TestResposta(t *testing.T) {
	casos := []struct {
		in            string
		decidido, sim bool
	}{
		{"\x1b_Gi=31;OK\x1b\\\x1b[?62;22c", true, true},
		{"\x1b[?62;22c", true, false},
		{"\x1b_Gi=31;EINVAL:x\x1b\\\x1b[?1c", true, false},
		{"\x1b_Gi=31;O", false, false},
		{"", false, false},
	}
	for _, c := range casos {
		d, s := Resposta([]byte(c.in))
		if d != c.decidido || s != c.sim {
			t.Fatalf("%q: decidido=%v sim=%v", c.in, d, s)
		}
	}
}

func TestDetectarSemRespostaCaiNoTimeout(t *testing.T) {
	r, w, _ := os.Pipe()
	defer r.Close()
	defer w.Close()
	var saida bytes.Buffer
	ini := time.Now()
	if Detectar(&saida, r, 150*time.Millisecond) {
		t.Fatal("sem resposta não pode dar suporte")
	}
	if d := time.Since(ini); d < 140*time.Millisecond || d > 2*time.Second {
		t.Fatalf("timeout fora: %v", d)
	}
	if !strings.Contains(saida.String(), "a=q") {
		t.Fatal("não consultou")
	}
}

func TestDetectarComOK(t *testing.T) {
	r, w, _ := os.Pipe()
	defer r.Close()
	go func() {
		time.Sleep(20 * time.Millisecond)
		w.Write([]byte("\x1b_Gi=31;"))
		time.Sleep(20 * time.Millisecond)
		w.Write([]byte("OK\x1b\\\x1b[?62c"))
	}()
	if !Detectar(&bytes.Buffer{}, r, time.Second) {
		t.Fatal("OK deveria ligar a imagem")
	}
}

func TestDetectarSoDA1(t *testing.T) {
	r, w, _ := os.Pipe()
	defer r.Close()
	w.Write([]byte("\x1b[?62;22c"))
	ini := time.Now()
	if Detectar(&bytes.Buffer{}, r, time.Second) || time.Since(ini) > 500*time.Millisecond {
		t.Fatal("DA1 sem OK = sem suporte, e sem esperar o timeout")
	}
}
