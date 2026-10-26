package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"strings"
)

// Dimensoes descreve cada PNG transmitido: largura x altura, modelo de cor e bytes.
func (g *Grafico) Dimensoes() []string {
	var out []string
	for _, p := range g.png {
		c, err := png.DecodeConfig(bytes.NewReader(p))
		if err != nil {
			out = append(out, "erro: "+err.Error())
			continue
		}
		out = append(out, fmt.Sprintf("%dx%d %s %dB", c.Width, c.Height, nomeModelo(c.ColorModel), len(p)))
	}
	return out
}

func nomeModelo(c color.Model) string {
	switch c {
	case color.NRGBAModel:
		return "NRGBA"
	case color.RGBAModel:
		return "RGBA"
	}
	return fmt.Sprintf("%T", c)
}

// resumo troca o payload base64 de cada APC pelo tamanho, para o log.
func resumo(s string) string {
	const ini, fim = "\x1b_G", "\x1b\\"
	var b strings.Builder
	n := 0
	for {
		i := strings.Index(s, ini)
		if i < 0 {
			break
		}
		j := strings.Index(s[i:], fim)
		if j < 0 {
			break
		}
		apc := s[i+len(ini) : i+j]
		if k := strings.IndexByte(apc, ';'); k >= 0 {
			apc = fmt.Sprintf("%s;<%dB>", apc[:k], len(apc)-k-1)
		}
		if n < 6 {
			b.WriteString(apc + " ")
		}
		n++
		s = s[i+j+len(fim):]
	}
	if n > 6 {
		fmt.Fprintf(&b, "... (%d APC no total)", n)
	}
	return strings.TrimSpace(b.String())
}
