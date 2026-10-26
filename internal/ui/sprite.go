package ui

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"fantasmas/assets"

	"github.com/charmbracelet/lipgloss"
)

// MeioBloco troca a célula cheia por ▀ (fg = pixel de cima, bg = de baixo).
const MeioBloco = false

type Sprite struct {
	Quadros         [][]string
	Delays          []time.Duration
	Colunas, Linhas int
}

func (s *Sprite) Quadro(i int) []string {
	n := len(s.Quadros)
	return s.Quadros[((i%n)+n)%n]
}

func (s *Sprite) Delay(i int) time.Duration {
	n := len(s.Delays)
	return s.Delays[((i%n)+n)%n]
}

type px struct {
	c  color.RGBA
	ok bool
}

func compor(g *gif.GIF) []*image.RGBA {
	w, h := g.Config.Width, g.Config.Height
	if w == 0 || h == 0 {
		b := g.Image[0].Bounds()
		w, h = b.Max.X, b.Max.Y
	}
	tela := image.NewRGBA(image.Rect(0, 0, w, h))
	out := make([]*image.RGBA, len(g.Image))
	for i, q := range g.Image {
		disp := byte(0)
		if i < len(g.Disposal) {
			disp = g.Disposal[i]
		}
		var antes *image.RGBA
		if disp == gif.DisposalPrevious {
			antes = image.NewRGBA(tela.Bounds())
			copy(antes.Pix, tela.Pix)
		}
		draw.Draw(tela, q.Bounds(), q, q.Bounds().Min, draw.Over)
		f := image.NewRGBA(tela.Bounds())
		copy(f.Pix, tela.Pix)
		out[i] = f
		switch disp {
		case gif.DisposalBackground:
			draw.Draw(tela, q.Bounds(), image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			tela = antes
		}
	}
	return out
}

func caixaOpaca(f *image.RGBA) image.Rectangle {
	var r image.Rectangle
	b := f.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if f.RGBAAt(x, y).A >= 128 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

// janelas centra cada quadro pela própria caixa, numa janela do tamanho da maior caixa (corpo ocupa mais células).
func janelas(fs []*image.RGBA) ([]image.Rectangle, int, int) {
	cs := make([]image.Rectangle, len(fs))
	W, H := 0, 0
	for i, f := range fs {
		cs[i] = caixaOpaca(f)
		W, H = max(W, cs[i].Dx()), max(H, cs[i].Dy())
	}
	out := make([]image.Rectangle, len(fs))
	for i, c := range cs {
		cx, cy := (c.Min.X+c.Max.X)/2, (c.Min.Y+c.Max.Y)/2
		out[i] = image.Rect(cx-W/2, cy-H/2, cx-W/2+W, cy-H/2+H)
	}
	return out, W, H
}

func dist(a, b color.RGBA) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// paleta conta a frequência de cada cor opaca em todos os quadros.
func paleta(fs []*image.RGBA) (map[color.RGBA]int, int) {
	n := map[color.RGBA]int{}
	tot := 0
	for _, f := range fs {
		for i := 0; i+3 < len(f.Pix); i += 4 {
			if f.Pix[i+3] >= 128 {
				n[color.RGBA{f.Pix[i], f.Pix[i+1], f.Pix[i+2], 255}]++
				tot++
			}
		}
	}
	return n, tot
}

type amostrador struct {
	corpo color.RGBA
	raro  map[color.RGBA]bool
}

// novoAmostrador marca como destaque as cores raras e distantes do corpo (olhos, dentes).
func novoAmostrador(fs []*image.RGBA) amostrador {
	n, tot := paleta(fs)
	a := amostrador{raro: map[color.RGBA]bool{}}
	top := 0
	for c, k := range n {
		if k > top {
			a.corpo, top = c, k
		}
	}
	for c, k := range n {
		if float64(k) < 0.05*float64(tot) && dist(c, a.corpo) > 90 {
			a.raro[c] = true
		}
	}
	return a
}

// fracaoDestaque: fração mínima da célula em cor de destaque (olho, dente) para ela ganhar essa cor.
const fracaoDestaque = 0.10

// amostra: destaque com ≥10% da área vence; senão maioria transparente vira fundo, e o resto é a cor dominante.
func (s amostrador) amostra(f *image.RGBA, x0, y0, x1, y1 float64) px {
	peso := map[color.RGBA]float64{}
	var opaco, tot float64
	for y := int(math.Floor(y0)); y < int(math.Ceil(y1)); y++ {
		for x := int(math.Floor(x0)); x < int(math.Ceil(x1)); x++ {
			p := (math.Min(x1, float64(x+1)) - math.Max(x0, float64(x))) * (math.Min(y1, float64(y+1)) - math.Max(y0, float64(y)))
			if p <= 0 {
				continue
			}
			tot += p
			if c := f.RGBAAt(x, y); c.A >= 128 {
				c.A = 255
				opaco += p
				peso[c] += p
			}
		}
	}
	if tot == 0 {
		return px{}
	}
	var dest color.RGBA
	dp := 0.0
	for c, p := range peso {
		if s.raro[c] && (p > dp || p == dp && dist(c, s.corpo) > dist(dest, s.corpo)) {
			dest, dp = c, p
		}
	}
	if dp >= fracaoDestaque*tot {
		return px{dest, true}
	}
	if opaco*2 < tot {
		return px{}
	}
	var melhor color.RGBA
	mp := -1.0
	for c, p := range peso {
		if p > mp || p == mp && dist(c, s.corpo) > dist(melhor, s.corpo) {
			melhor, mp = c, p
		}
	}
	return px{melhor, true}
}

func hexDe(p px, fundo lipgloss.Color) lipgloss.Color {
	if !p.ok {
		return fundo
	}
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", p.c.R, p.c.G, p.c.B))
}

// delayDe converte centésimos do GIF; 0 vira 100 ms e o mínimo é 50 ms.
func delayDe(cs int) time.Duration {
	if cs <= 0 {
		cs = 10
	}
	return time.Duration(max(cs, 5)) * 10 * time.Millisecond
}

func CarregarSprite(gifBytes []byte, maxCol, maxLin int, fundo lipgloss.Color) (*Sprite, error) {
	g, err := gif.DecodeAll(bytes.NewReader(gifBytes))
	if err != nil {
		return nil, err
	}
	if len(g.Image) == 0 {
		return nil, errors.New("gif sem quadros")
	}
	fs := compor(g)
	js, w, h := janelas(fs)
	if w == 0 || h == 0 {
		return nil, errors.New("gif todo transparente")
	}
	am := novoAmostrador(fs)
	W, H := float64(w), float64(h)
	// sx = px por coluna; cada linha do terminal cobre 2*sx px (célula ~2:1).
	sx := math.Max(1, math.Max(W/float64(maxCol), H/(2*float64(maxLin))))
	sy := 2 * sx
	cols := min(int(math.Ceil(W/sx)), maxCol)
	lins := min(int(math.Ceil(H/sy)), maxLin)
	s := &Sprite{Colunas: cols, Linhas: lins}
	for i, f := range fs {
		ox, oy := float64(js[i].Min.X), float64(js[i].Min.Y)
		linhas := make([]string, lins)
		for l := 0; l < lins; l++ {
			var b strings.Builder
			y0 := oy + float64(l)*sy
			for c := 0; c < cols; c++ {
				x0, x1 := ox+float64(c)*sx, ox+float64(c+1)*sx
				if MeioBloco {
					cima := am.amostra(f, x0, y0, x1, y0+sx)
					baixo := am.amostra(f, x0, y0+sx, x1, y0+sy)
					b.WriteString(lipgloss.NewStyle().Foreground(hexDe(cima, fundo)).Background(hexDe(baixo, fundo)).Render("▀"))
				} else {
					b.WriteString(lipgloss.NewStyle().Background(hexDe(am.amostra(f, x0, y0, x1, y0+sy), fundo)).Render(" "))
				}
			}
			linhas[l] = b.String()
		}
		s.Quadros = append(s.Quadros, linhas)
		d := 0
		if i < len(g.Delay) {
			d = g.Delay[i]
		}
		s.Delays = append(s.Delays, delayDe(d))
	}
	return s, nil
}

// tamanhoSprite: 16x7 no largo; menor no médio e no estreito para o cartão do topo caber.
func tamanhoSprite(largura int) (int, int) {
	switch Faixa(largura) {
	case "largo":
		return 16, 7
	case "medio":
		return 14, 6
	}
	return 12, 5
}

var (
	spriteMu    sync.Mutex
	spriteMem   = map[[2]int]*Sprite{}
	gifEmbutido = assets.Gengar
	gifTopo     []byte
	spriteFn    = spriteTopo
)

func limparSprites() {
	spriteMu.Lock()
	spriteMem = map[[2]int]*Sprite{}
	spriteMu.Unlock()
}

// UsarSprite troca o GIF do topo pelo arquivo; se não ler ou não decodificar, volta ao embutido e devolve o erro.
func UsarSprite(caminho string) error {
	defer limparSprites()
	b, err := os.ReadFile(caminho)
	if err == nil {
		_, err = CarregarSprite(b, 16, 7, T.Mantle)
	}
	if err != nil {
		gifTopo = nil
		return fmt.Errorf("sprite %s: %w", caminho, err)
	}
	gifTopo = b
	return nil
}

// spriteTopo devolve o Gengar no tamanho da faixa, ou nil se nenhum GIF decodificar (cai no mascote nativo).
func spriteTopo(largura int) *Sprite {
	c, l := tamanhoSprite(largura)
	k := [2]int{c, l}
	spriteMu.Lock()
	defer spriteMu.Unlock()
	if s, ok := spriteMem[k]; ok {
		return s
	}
	var s *Sprite
	if gifTopo != nil {
		s, _ = CarregarSprite(gifTopo, c, l, T.Mantle)
	}
	if s == nil {
		s, _ = CarregarSprite(gifEmbutido, c, l, T.Mantle)
	}
	spriteMem[k] = s
	return s
}
