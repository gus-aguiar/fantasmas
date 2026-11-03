package ui

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fantasmas/internal/config"
)

// pngTeste: fundo transparente com um bloco de n cores distintas no meio.
func pngTeste(t *testing.T, n int) []byte {
	t.Helper()
	lado := 24
	for lado*lado/4 < n {
		lado *= 2
	}
	img := image.NewNRGBA(image.Rect(0, 0, lado, lado))
	k := 0
	for y := lado / 4; y < 3*lado/4; y++ {
		for x := lado / 4; x < 3*lado/4; x++ {
			img.Set(x, y, color.NRGBA{uint8(k % 256), uint8(k / 256 % 256), 200, 255})
			if k+1 < n {
				k++
			}
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestParaGIFDePNG(t *testing.T) {
	for _, n := range []int{3, 600} {
		g, err := paraGIF(pngTeste(t, n))
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		dec, err := gif.DecodeAll(bytes.NewReader(g))
		if err != nil || len(dec.Image) != 1 {
			t.Fatalf("n=%d: gif inválido: %v", n, err)
		}
		q := compor(dec)[0]
		if _, _, _, a := q.At(0, 0).RGBA(); a != 0 {
			t.Fatalf("n=%d: canto deveria ser transparente", n)
		}
		meio := q.Bounds().Dx() / 2
		if _, _, b, a := q.At(meio, meio).RGBA(); a == 0 || b>>8 < 150 {
			t.Fatalf("n=%d: meio perdeu a cor", n)
		}
		if _, err := CarregarSprite(g, 16, 7, T.Mantle); err != nil {
			t.Fatalf("n=%d: sprite: %v", n, err)
		}
	}
	if _, err := paraGIF([]byte("texto")); err == nil {
		t.Fatal("lixo aceito")
	}
}

func TestCatalogoDeFiguras(t *testing.T) {
	defer usarBytes(nil, "")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Teste.png"), pngTeste(t, 3), 0o644)
	os.WriteFile(filepath.Join(dir, "lapras.png"), pngTeste(t, 3), 0o644)
	os.WriteFile(filepath.Join(dir, "leia.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, ".oculta.gif"), []byte("x"), 0o644)
	if got := strings.Join(NomesFiguras(dir), ","); got != "gengar,entei,lapras,teste" {
		t.Fatalf("figuras=%s", got)
	}
	if got := strings.Join(NomesFiguras(""), ","); got != "gengar,entei,lapras" {
		t.Fatalf("embutidas=%s", got)
	}
	for _, n := range []string{"teste", "entei", "lapras"} {
		if err := UsarFigura(n, dir); err != nil || gifTopo == nil {
			t.Fatalf("%s: %v", n, err)
		}
	}
	embutido, _ := figurasDe("")["lapras"]()
	if UsarFigura("lapras", dir); bytes.Equal(gifTopo, embutido) {
		t.Fatal("lapras do usuário deveria vencer o embutido")
	}
	if err := UsarFigura("", dir); err != nil || !bytes.Equal(gifTopo, gifEmbutido) {
		t.Fatalf("vazio deveria ser o gengar: %v", err)
	}
	if err := UsarFigura("sumiu", dir); err == nil || gifTopo != nil {
		t.Fatal("figura ausente deveria dar erro e voltar ao embutido")
	}
}

func TestSettingsTrocaFigura(t *testing.T) {
	defer usarBytes(nil, "")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "teste.png"), pngTeste(t, 3), 0o644)
	antes, _ := config.Achar(config.ChaveFigura)
	defer config.DefinirValores(config.ChaveFigura, antes.Valores)
	config.DefinirValores(config.ChaveFigura, NomesFiguras(dir))

	m, l, _ := modeloAjustes(t, nil)
	m.acoes.DirFiguras = dir
	m.sel = indiceAjuste(t, config.ChaveFigura)
	if out := render(m, 110, 50); !strings.Contains(out, "4 figuras") {
		t.Fatalf("sem a contagem de figuras:\n%s", out)
	}
	for _, quer := range []string{"entei", "lapras", "teste"} {
		m = tecla(m, "enter")
		if l.Valor(config.ChaveFigura) != quer || gifTopo == nil {
			t.Fatalf("enter deveria escolher %s: %q", quer, l.Valor(config.ChaveFigura))
		}
	}
	m = tecla(m, "enter")
	if l.Origem(config.ChaveFigura) != config.Padrao || !bytes.Equal(gifTopo, gifEmbutido) {
		t.Fatal("segundo enter deveria voltar ao gengar")
	}
	if !strings.Contains(render(m, 110, 50), "FANTASMAS_FIGURA: gengar") {
		t.Fatal("toast sem a figura")
	}
}
