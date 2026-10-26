package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"strings"
	"testing"
	"time"

	"fantasmas/assets"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func comImagem(m Model, porQuadro bool) (Model, *bytes.Buffer) {
	modo := ModoAnimacao
	if porQuadro {
		modo = ModoQuadros
	}
	return comImagemModo(m, modo)
}

func comImagemModo(m Model, modo string) (Model, *bytes.Buffer) {
	var b bytes.Buffer
	g := novoGrafico(&b, modo, [][]byte{[]byte("q0"), []byte("q1"), []byte("q2")}, []int{100, 100, 100})
	return m.ComGrafico(g), &b
}

func TestImagemColocaReservaEApaga(t *testing.T) {
	m, b := comImagem(modelo(fixture(), &registro{}), false)
	m, _ = atualiza(t, m, tea.WindowSizeMsg{Width: 96, Height: 40})
	s := b.String()
	if strings.Count(s, "a=t,") != 1 || !strings.Contains(s, "a=f,") || !strings.Contains(s, "s=3") {
		t.Fatalf("transmissão animada errada: %q", s)
	}
	if !strings.Contains(s, "c=16,r=8,C=1") {
		t.Fatalf("sem colocação 16x8: %q", s)
	}
	a := m.montar(96, 40).img
	if !a.ok || a.c != 16 || a.r != 8 {
		t.Fatalf("área: %+v", a)
	}
	if !strings.Contains(s, fmt.Sprintf("\x1b[%d;%dH", a.y+1, a.x+1)) {
		t.Fatalf("posição não bate com a área %+v: %q", a, s)
	}
	ls := strings.Split(render(m, 96, 40), "\n")
	for y := a.y; y < a.y+a.r; y++ {
		r := []rune(ls[y])
		if got := string(r[a.x : a.x+a.c]); strings.TrimSpace(got) != "" {
			t.Fatalf("área da imagem com texto na linha %d: %q", y, got)
		}
	}
	b.Reset()
	m, _ = atualiza(t, m, tickQuadroMsg{})
	if b.Len() != 0 {
		t.Fatalf("animação nativa não deveria recolocar a cada quadro: %q", b.String())
	}
	m = aperta(t, m, "?")
	if !strings.Contains(b.String(), "a=d,d=i") {
		t.Fatalf("ajuda não apagou a imagem: %q", b.String())
	}
	b.Reset()
	m = aperta(t, m, "x")
	if !strings.Contains(b.String(), "a=p") || strings.Contains(b.String(), "a=t,") {
		t.Fatalf("voltar deveria recolocar sem retransmitir: %q", b.String())
	}
	b.Reset()
	m, _ = atualiza(t, m, tea.WindowSizeMsg{Width: 60, Height: 30})
	if !strings.Contains(b.String(), "c=12,r=6") {
		t.Fatalf("estreito deveria ser 12x6: %q", b.String())
	}
	b.Reset()
	m, _ = atualiza(t, m, tea.WindowSizeMsg{Width: 40, Height: 30})
	if !strings.Contains(b.String(), "a=d,d=i") || m.montar(40, 30).img.ok {
		t.Fatalf("mínimo não mostra imagem: %q", b.String())
	}
	b.Reset()
	m.img.Encerrar()
	if !strings.Contains(b.String(), "a=d,d=I") {
		t.Fatalf("encerrar não liberou: %q", b.String())
	}
}

func TestImagemPorQuadro(t *testing.T) {
	m, b := comImagem(modelo(fixture(), &registro{}), true)
	m, _ = atualiza(t, m, tea.WindowSizeMsg{Width: 96, Height: 40})
	id := idImagem
	if !strings.Contains(b.String(), fmt.Sprintf("a=t,f=100,i=%d,", id+2)) || strings.Contains(b.String(), "a=f,") {
		t.Fatalf("modo quadros transmite um id por quadro: %q", b.String())
	}
	b.Reset()
	m, _ = atualiza(t, m, tickQuadroMsg{})
	s := b.String()
	if !strings.Contains(s, fmt.Sprintf("a=p,i=%d,", id+1)) || !strings.Contains(s, fmt.Sprintf("a=d,d=i,i=%d,", id)) {
		t.Fatalf("quadro seguinte: %q", s)
	}
	if strings.Index(s, "a=p") > strings.Index(s, "a=d") {
		t.Fatal("coloca o novo antes de apagar o velho, para não piscar")
	}
}

// A área do Gengar (célula ou imagem) nunca invade o corpo, em toda gaveta e largura.
func TestSpriteNaoInvadeCorpo(t *testing.T) {
	for _, img := range []bool{false, true} {
		for g := range Gavetas {
			for _, w := range []int{46, 56, 69, 70, 89, 90, 96, 140} {
				for _, h := range []int{12, 20, 30, 44} {
					m := modelo(fixture(), &registro{})
					if img {
						m, _ = comImagem(m, false)
					}
					m.largura, m.altura = w, h
					m = m.irGaveta(g)
					q := m.montar(w, h)
					if q.img.ok && q.img.y+q.img.r > q.topo {
						t.Fatalf("img=%v gaveta %d %dx%d: área %+v passa do topo (%d linhas)", img, g, w, h, q.img, q.topo)
					}
					if !img && Faixa(w) != "minimo" {
						if sp := spriteFn(w); sp != nil && q.topo < min(sp.Linhas, h-1) {
							t.Fatalf("gaveta %d %dx%d: topo com %d linhas < sprite %d", g, w, h, q.topo, sp.Linhas)
						}
					}
					if Gavetas[g].ID == "crew" && h >= 20 {
						ls := strings.Split(ansi.Strip(m.Render(w, h)), "\n")
						for y := 0; y < q.topo; y++ {
							if strings.Contains(ls[y], "ativos") {
								t.Fatalf("%dx%d: corpo no topo, linha %d", w, h, y)
							}
						}
					}
				}
			}
		}
	}
}

func TestQuadrosPNGDoGif(t *testing.T) {
	ps, ds, err := QuadrosPNG(assets.Gengar)
	if err != nil || len(ps) == 0 || len(ps) != len(ds) {
		t.Fatalf("png=%d delays=%d err=%v", len(ps), len(ds), err)
	}
	if !bytes.HasPrefix(ps[0], []byte("\x89PNG")) {
		t.Fatal("quadro não é PNG")
	}
}

// ---- minors da re-review ----

func TestCliqueForaDaOpcaoDesarma(t *testing.T) {
	c := novaCena(t)
	m := abrirJuntar(t, c)
	m = passo(t, m, "enter")
	if m.armado == 0 {
		t.Fatal("não armou")
	}
	x, y := onde(t, m, "↻", 0)
	m, _ = clique(t, m, x, y)
	if m.armado != 0 {
		t.Fatal("clique fora das opções deveria desarmar")
	}
	m = passo(t, m, "enter")
	if len(c.resps) != 0 {
		t.Fatalf("respondeu depois de desarmado: %v", c.resps)
	}
}

func TestEnterComFocoNoTextoNaoResponde(t *testing.T) {
	c := novaCena(t)
	m := abrir(t, c, "plano")
	m = passo(t, m, "tab")
	m = passo(t, m, "enter")
	if len(c.resps) != 0 || m.rascunho != nil {
		t.Fatalf("enter com foco no texto respondeu: %v", c.resps)
	}
}

func TestCliqueComFiltroAbertoNaoDeixaDigitando(t *testing.T) {
	c := novaCena(t)
	m := c.modelo(c.estado("plano"))
	m = aperta(t, m, "/")
	x, y := onde(t, m, "decidir c-pl-17724", 0)
	m, _ = clique(t, m, x, y)
	if m.aberto == nil {
		t.Fatal("clique não abriu o detalhe")
	}
	m = passo(t, m, "esc")
	if m.aberto != nil || m.filtro != nil {
		t.Fatal("ao voltar o campo do filtro não pode estar em digitação")
	}
}

func soTopo(m Model, w int) []string {
	ls, _ := m.linhasTopo(w)
	return ls
}

func TestQuadrosPNGResolucaoOriginalRGBA(t *testing.T) {
	g, err := gif.DecodeAll(bytes.NewReader(gifEmbutido))
	if err != nil {
		t.Fatal(err)
	}
	_, w, h := janelas(compor(g))
	ps, _, err := QuadrosPNG(gifEmbutido)
	if err != nil {
		t.Fatal(err)
	}
	if w > g.Config.Width || h > g.Config.Height || w < g.Config.Width/2 {
		t.Fatalf("recorte %dx%d fora do gif %dx%d", w, h, g.Config.Width, g.Config.Height)
	}
	transp := false
	for i, p := range ps {
		img, err := png.Decode(bytes.NewReader(p))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != w*FatorAmpliacao || img.Bounds().Dy() != h*FatorAmpliacao {
			t.Fatalf("quadro %d: %v, quer %dx%d", i, img.Bounds(), w*FatorAmpliacao, h*FatorAmpliacao)
		}
		if _, ok := img.(*image.NRGBA); !ok {
			if _, ok := img.(*image.RGBA); !ok {
				t.Fatalf("quadro %d sem alfa: %T", i, img)
			}
		}
		if _, _, _, a := img.At(0, 0).RGBA(); a == 0 {
			transp = true
		}
	}
	if !transp {
		t.Fatal("nenhum quadro com canto transparente")
	}
}

func TestAmpliarVizinhoMaisProximo(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			o := src.PixOffset(x, y)
			src.Pix[o], src.Pix[o+3] = uint8(10*x+y), uint8(80*x)
		}
	}
	dst := ampliar(src, 4)
	if dst.Bounds().Dx() != 12 || dst.Bounds().Dy() != 8 {
		t.Fatalf("dimensão %v", dst.Bounds())
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 12; x++ {
			if dst.At(x, y) != src.At(x/4, y/4) {
				t.Fatalf("pixel (%d,%d) %v != %v", x, y, dst.At(x, y), src.At(x/4, y/4))
			}
		}
	}
}

func TestImagemModoT(t *testing.T) {
	m, b := comImagemModo(modelo(fixture(), &registro{}), ModoT)
	agora := time.Unix(0, 0)
	m.img.agora = func() time.Time { return agora }
	m, _ = atualiza(t, m, tea.WindowSizeMsg{Width: 96, Height: 40})
	s := b.String()
	a := m.montar(96, 40).img
	pos := fmt.Sprintf("\x1b7\x1b[%d;%dH", a.y+1, a.x+1)
	quer := fmt.Sprintf("a=T,f=100,i=%d,p=1,c=16,r=8,q=2,C=1,m=0;", idImagem)
	if !strings.HasPrefix(s, pos) || !strings.Contains(s, quer) || !strings.HasSuffix(s, "\x1b8") {
		t.Fatalf("modo T: %q", s)
	}
	if strings.Contains(s, "a=t,") || strings.Contains(s, "a=p,") || strings.Contains(s, "a=f,") {
		t.Fatalf("modo T só usa a=T: %q", s)
	}
	b.Reset()
	agora = agora.Add(50 * time.Millisecond)
	m, _ = atualiza(t, m, tickQuadroMsg{})
	if b.Len() != 0 {
		t.Fatalf("acima de 10 fps deveria pular: %q", b.String())
	}
	agora = agora.Add(60 * time.Millisecond)
	m, _ = atualiza(t, m, tickQuadroMsg{})
	// q2 em base64 é cTI=: o quadro enviado é o do tempo real, não o seguinte ao último enviado.
	if !strings.Contains(b.String(), "a=T,") || !strings.Contains(b.String(), "cTI=") {
		t.Fatalf("depois de 100 ms manda o quadro atual: %q", b.String())
	}
	b.Reset()
	m = aperta(t, m, "?")
	if !strings.Contains(b.String(), "a=d,d=i") {
		t.Fatalf("ajuda não apagou: %q", b.String())
	}
}

func TestAjudaMostraModoImagem(t *testing.T) {
	m := modelo(fixture(), &registro{}).ComModoImagem("kitty-animação")
	if !strings.Contains(strings.Join(m.ajudaLinhas(80), "\n"), "kitty-animação") {
		t.Fatal("ajuda sem o modo")
	}
}
