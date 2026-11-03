package ui

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fantasmas/assets"
)

// Figuras do topo: o gengar padrão, as embutidas em assets/figuras e as do usuário em
// <config>/fantasmas/figuras. Cada arquivo GIF ou PNG é uma figura, com o nome do arquivo sem extensão;
// a do usuário vence a embutida de mesmo nome.

// FiguraPadrao é o gengar embutido.
const FiguraPadrao = "gengar"

func nomeFigura(arquivo string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(arquivo))
	if ext != ".gif" && ext != ".png" {
		return "", false
	}
	n := strings.ToLower(strings.TrimSuffix(filepath.Base(arquivo), filepath.Ext(arquivo)))
	return n, n != "" && !strings.HasPrefix(n, ".")
}

// figurasDe devolve nome → leitor, das embutidas e das do usuário (dir "" = só embutidas).
func figurasDe(dir string) map[string]func() ([]byte, error) {
	out := map[string]func() ([]byte, error){}
	es, _ := fs.ReadDir(assets.Figuras, "figuras")
	for _, e := range es {
		if n, ok := nomeFigura(e.Name()); ok && !e.IsDir() {
			p := "figuras/" + e.Name()
			out[n] = func() ([]byte, error) { return assets.Figuras.ReadFile(p) }
		}
	}
	if dir == "" {
		return out
	}
	us, _ := os.ReadDir(dir)
	for _, e := range us {
		if n, ok := nomeFigura(e.Name()); ok && !e.IsDir() {
			p := filepath.Join(dir, e.Name())
			out[n] = func() ([]byte, error) { return os.ReadFile(p) }
		}
	}
	return out
}

// NomesFiguras lista as figuras, com o gengar primeiro e as demais em ordem alfabética.
func NomesFiguras(dir string) []string {
	var ns []string
	for n := range figurasDe(dir) {
		if n != FiguraPadrao {
			ns = append(ns, n)
		}
	}
	sort.Strings(ns)
	return append([]string{FiguraPadrao}, ns...)
}

// UsarFigura põe a figura no topo; "" ou "gengar" (sem arquivo do usuário) é o embutido.
func UsarFigura(nome, dir string) error {
	if nome == "" {
		nome = FiguraPadrao
	}
	ler, ok := figurasDe(dir)[nome]
	if !ok {
		usarBytes(nil, "")
		return errors.New("figura " + nome + " não encontrada; ponha o GIF ou PNG em " + dir)
	}
	b, err := ler()
	if err != nil {
		usarBytes(nil, "")
		return err
	}
	return usarBytes(b, "figura "+nome)
}

// paraGIF aceita GIF como veio e converte imagem estática (PNG) num GIF de um quadro:
// paleta exata até 255 cores, senão a Plan9; alfa abaixo da metade vira transparente.
func paraGIF(b []byte) ([]byte, error) {
	if _, err := gif.DecodeConfig(bytes.NewReader(b)); err == nil {
		return b, nil
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, errors.New("nem GIF nem PNG")
	}
	r := img.Bounds()
	opaca := func(c color.Color) (color.NRGBA, bool) {
		n := color.NRGBAModel.Convert(c).(color.NRGBA)
		return color.NRGBA{n.R, n.G, n.B, 255}, n.A >= 128
	}
	pal := color.Palette{color.NRGBA{}}
	idx := map[color.NRGBA]uint8{}
	exata := true
	for y := r.Min.Y; y < r.Max.Y && exata; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c, ok := opaca(img.At(x, y))
			if _, visto := idx[c]; !ok || visto {
				continue
			}
			if len(pal) == 256 {
				exata = false
				break
			}
			idx[c] = uint8(len(pal))
			pal = append(pal, c)
		}
	}
	if !exata {
		pal = append(color.Palette{color.NRGBA{}}, palette.Plan9[:255]...)
	}
	p := image.NewPaletted(image.Rect(0, 0, r.Dx(), r.Dy()), pal)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c, ok := opaca(img.At(x, y))
			switch {
			case !ok:
				p.SetColorIndex(x-r.Min.X, y-r.Min.Y, 0)
			case exata:
				p.SetColorIndex(x-r.Min.X, y-r.Min.Y, idx[c])
			default:
				p.SetColorIndex(x-r.Min.X, y-r.Min.Y, uint8(1+color.Palette(palette.Plan9[:255]).Index(c)))
			}
		}
	}
	var buf bytes.Buffer
	if err := gif.Encode(&buf, p, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
