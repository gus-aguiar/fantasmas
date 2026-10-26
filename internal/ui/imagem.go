package ui

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"strings"
	"sync"
	"time"

	"fantasmas/internal/kitty"
)

const idImagem = 4100

// FatorAmpliacao amplia cada quadro por vizinho mais próximo antes do PNG, para o terminal não suavizar.
const FatorAmpliacao = 4

// Modos de envio da imagem.
const (
	ModoAnimacao = "animacao"
	ModoQuadros  = "quadros"
	ModoT        = "T"
)

const intervaloT = 100 * time.Millisecond

// Grafico desenha o Gengar no topo pelo protocolo de gráficos do kitty.
type Grafico struct {
	w         io.Writer
	porQuadro bool
	modo      string
	agora     func() time.Time
	ultimoT   time.Time
	png       [][]byte
	delays    []int
	envio     sync.Once
	// Log, quando não nil, recebe um resumo de cada sequência enviada.
	Log func(string)
}

type areaImg struct {
	x, y, c, r int
	ok         bool
}

// NovoGrafico carrega os quadros do GIF do topo no modo pedido (ModoAnimacao, ModoQuadros ou ModoT).
func NovoGrafico(w io.Writer, modo string) (*Grafico, error) {
	g := gifTopo
	if g == nil {
		g = gifEmbutido
	}
	ps, ds, err := QuadrosPNG(g)
	if err != nil {
		return nil, err
	}
	ms := make([]int, len(ds))
	for i, d := range ds {
		ms[i] = int(d.Milliseconds())
	}
	return novoGrafico(w, modo, ps, ms), nil
}

func novoGrafico(w io.Writer, modo string, ps [][]byte, delays []int) *Grafico {
	return &Grafico{w: w, modo: modo, porQuadro: modo == ModoQuadros, png: ps, delays: delays, agora: time.Now}
}

// Modo devolve o modo de envio.
func (g *Grafico) Modo() string { return g.modo }

// porTempo: no modo T cada quadro é reenviado, limitado a ~10 fps.
func (g *Grafico) porTempo() bool { return g.modo == ModoT }

func (g *Grafico) escrever(s string) {
	if s != "" {
		if g.Log != nil {
			g.Log(resumo(s))
		}
		_, _ = io.WriteString(g.w, s)
	}
}

func (g *Grafico) enviar() {
	g.envio.Do(func() {
		if g.porTempo() {
			return
		}
		if g.porQuadro {
			g.escrever(kitty.Quadros(idImagem, g.png))
		} else {
			g.escrever(kitty.Animacao(idImagem, g.png, g.delays))
		}
	})
}

func (g *Grafico) id(quadro int) int {
	if !g.porQuadro {
		return idImagem
	}
	n := len(g.png)
	return idImagem + ((quadro%n)+n)%n
}

// devoPular: no modo T, pula o quadro se o último envio foi há menos de intervaloT (forca ignora).
func (g *Grafico) devoPular(forca bool) bool {
	return g.porTempo() && !forca && g.agora().Sub(g.ultimoT) < intervaloT
}

func (g *Grafico) colocar(a areaImg, quadro, antes int) {
	if g.porTempo() {
		n := len(g.png)
		g.ultimoT = g.agora()
		g.escrever(kitty.TransmitirColocar(idImagem, 1, a.x, a.y, a.c, a.r, g.png[((quadro%n)+n)%n]))
		return
	}
	g.enviar()
	s := kitty.Colocar(g.id(quadro), 1, a.x, a.y, a.c, a.r)
	if g.porQuadro && antes >= 0 && g.id(antes) != g.id(quadro) {
		s += kitty.Apagar(g.id(antes))
	}
	g.escrever(s)
}

func (g *Grafico) esconder(quadro int) { g.escrever(kitty.Apagar(g.id(quadro))) }

// Encerrar tira todas as imagens do app do terminal.
func (g *Grafico) Encerrar() {
	var b strings.Builder
	n := 1
	if g.porQuadro {
		n = len(g.png)
	}
	for i := 0; i < n; i++ {
		b.WriteString(kitty.Liberar(idImagem + i))
	}
	g.escrever(b.String())
}

// tamanhoImagem: 16x8 no largo, proporcionalmente menor no médio e no estreito.
func tamanhoImagem(largura int) (int, int) {
	switch Faixa(largura) {
	case "largo":
		return 16, 8
	case "medio":
		return 14, 7
	}
	return 12, 6
}

// QuadrosPNG compõe os quadros do GIF, recorta como o sprite em células e codifica em PNG.
func QuadrosPNG(gifBytes []byte) ([][]byte, []time.Duration, error) {
	g, err := gif.DecodeAll(bytes.NewReader(gifBytes))
	if err != nil {
		return nil, nil, err
	}
	if len(g.Image) == 0 {
		return nil, nil, errors.New("gif sem quadros")
	}
	fs := compor(g)
	js, w, h := janelas(fs)
	if w == 0 || h == 0 {
		return nil, nil, errors.New("gif todo transparente")
	}
	var ps [][]byte
	var ds []time.Duration
	for i, f := range fs {
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.Draw(dst, dst.Bounds(), f, js[i].Min, draw.Src)
		var b bytes.Buffer
		if err := png.Encode(&b, ampliar(dst, FatorAmpliacao)); err != nil {
			return nil, nil, err
		}
		ps = append(ps, b.Bytes())
		d := 0
		if i < len(g.Delay) {
			d = g.Delay[i]
		}
		ds = append(ds, delayDe(d))
	}
	return ps, ds, nil
}

// ampliar repete cada pixel em blocos f x f (vizinho mais próximo).
func ampliar(src *image.NRGBA, f int) *image.NRGBA {
	r := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, r.Dx()*f, r.Dy()*f))
	for y := 0; y < dst.Rect.Dy(); y++ {
		for x := 0; x < dst.Rect.Dx(); x++ {
			o, d := src.PixOffset(r.Min.X+x/f, r.Min.Y+y/f), dst.PixOffset(x, y)
			copy(dst.Pix[d:d+4], src.Pix[o:o+4])
		}
	}
	return dst
}

// sincImagem coloca, move ou apaga a imagem conforme o que o topo mostra agora.
func (m Model) sincImagem(forca bool) Model {
	if m.img == nil || m.largura <= 0 || m.altura <= 0 {
		return m
	}
	a := m.montar(m.largura, m.altura).img
	switch {
	case !a.ok:
		if m.imgPos.ok {
			m.img.esconder(m.imgQuadro)
		}
		m.imgPos = areaImg{}
	case forca || a != m.imgPos || (m.img.porQuadro || m.img.porTempo()) && m.imgQuadro != m.quadro:
		if a == m.imgPos && m.img.devoPular(forca) {
			return m
		}
		ant := m.imgQuadro
		if !m.imgPos.ok {
			ant = -1
		}
		m.img.colocar(a, m.quadro, ant)
		m.imgPos, m.imgQuadro = a, m.quadro
	}
	return m
}
