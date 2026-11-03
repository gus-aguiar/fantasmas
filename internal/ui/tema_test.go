package ui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// comTemas roda f em cada tema de cores fixas (o terminal usa a paleta do próprio terminal).
func comTemas(t *testing.T, f func(t *testing.T)) {
	for _, nome := range []string{"roxo", "orca"} {
		t.Run(nome, func(t *testing.T) {
			defer UsarTema("roxo")
			if err := UsarTema(nome); err != nil {
				t.Fatal(err)
			}
			f(t)
		})
	}
}

func TestContrasteLegivel(t *testing.T) { comTemas(t, checaContraste) }

func checaContraste(t *testing.T) {
	checa := func(nome string, fg lipgloss.Color, min float64, fundos map[string]lipgloss.Color) {
		for fn, bg := range fundos {
			if c := contraste(fg, bg); c < min {
				t.Errorf("%s sobre %s: %.2f < %.1f", nome, fn, c, min)
			}
		}
	}
	quatro := map[string]lipgloss.Color{"Base": T.Base, "Mantle": T.Mantle, "Surface": T.Surface, "Sel": T.Sel}
	dois := map[string]lipgloss.Color{"Base": T.Base, "Mantle": T.Mantle}
	checa("Texto", T.Texto, 7.0, quatro)
	checa("Sub", T.Sub, 4.5, quatro)
	checa("Dim", T.Dim, 3.0, dois)
	for n, c := range map[string]lipgloss.Color{"Acao": T.Acao, "Ok": T.Ok, "Espera": T.Espera,
		"Andamento": T.Andamento, "Erro": T.Erro, "Info": T.Info} {
		checa(n, c, 4.5, dois)
	}
	checa("ChipTexto", T.ChipTexto, 4.5, map[string]lipgloss.Color{"Chip": T.Chip})
	checa("ChipVerdeTexto", T.ChipVerdeTexto, 4.5, map[string]lipgloss.Color{"ChipVerde": T.ChipVerde})
	for i, c := range T.Acentos {
		checa(fmt.Sprintf("Acento%d", i), c, 4.5, quatro)
		checa(fmt.Sprintf("Mantle sobre Acento%d", i), T.Mantle, 4.5, map[string]lipgloss.Color{"acento": c})
	}
	for n, c := range map[string]lipgloss.Color{"CtxBaixo": T.CtxBaixo, "CtxMedio": T.CtxMedio, "CtxAlto": T.CtxAlto} {
		checa(n, c, 4.5, dois)
	}
}

func neutros() map[string]lipgloss.Color {
	return map[string]lipgloss.Color{
		"Base": T.Base, "Mantle": T.Mantle, "Surface": T.Surface, "Sel": T.Sel, "Borda": T.Borda,
		"Texto": T.Texto, "Sub": T.Sub, "Dim": T.Dim,
	}
}

func TestFundosENeutrosRoxos(t *testing.T) {
	for n, c := range neutros() {
		h, s := matiz(c)
		if s >= 0.08 && (h < 220 || h > 340) {
			t.Errorf("%s %s: matiz %.0f fora de [220,340]", n, c, h)
		}
	}
}

func TestEstadosDistinguiveis(t *testing.T) { comTemas(t, checaEstados) }

func checaEstados(t *testing.T) {
	est := []struct {
		n string
		c lipgloss.Color
	}{{"Ok", T.Ok}, {"Espera", T.Espera}, {"Andamento", T.Andamento}, {"Erro", T.Erro}}
	for i := range est {
		for j := i + 1; j < len(est); j++ {
			r1, g1, b1 := rgb(est[i].c)
			r2, g2, b2 := rgb(est[j].c)
			d := 255 * math.Sqrt((r1-r2)*(r1-r2)+(g1-g2)*(g1-g2)+(b1-b2)*(b1-b2))
			if d < 60 {
				t.Errorf("%s x %s: distancia %.1f < 60", est[i].n, est[j].n, d)
			}
		}
	}
}

func TestAcentosDistintosEntreSi(t *testing.T) { comTemas(t, checaAcentos) }

func checaAcentos(t *testing.T) {
	for i := range T.Acentos {
		for j := i + 1; j < len(T.Acentos); j++ {
			h1, _ := matiz(T.Acentos[i])
			h2, _ := matiz(T.Acentos[j])
			d := math.Abs(h1 - h2)
			if d > 180 {
				d = 360 - d
			}
			if d < 25 {
				t.Errorf("acentos %d e %d: matiz a %.0f°", i, j, d)
			}
		}
	}
}

func TestUsarTema(t *testing.T) {
	defer UsarTema("roxo")
	if err := UsarTema("verde"); err == nil || !strings.Contains(err.Error(), "orca") {
		t.Fatalf("tema desconhecido: %v", err)
	}
	if err := UsarTema("terminal"); err != nil || T.Base != "" || T.Mantle != "" {
		t.Fatalf("terminal deveria não pintar fundo: %v %q %q", err, T.Base, T.Mantle)
	}
	if err := UsarTema(""); err != nil || T.Base != temaRoxo.Base {
		t.Fatalf("vazio deveria ser o roxo: %v %q", err, T.Base)
	}
	if temaOrca.Acentos != temaRoxo.Acentos {
		t.Fatal("orca deveria manter os acentos do painel")
	}
}

func TestTemaTerminalSoANSI(t *testing.T) {
	re := regexp.MustCompile(`^([0-9]|1[0-5]|23[2-9]|24[0-9]|25[0-5])?$`)
	v := reflect.ValueOf(temaTerminal)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() == reflect.Array {
			for j := 0; j < f.Len(); j++ {
				if c := f.Index(j).String(); !re.MatchString(c) {
					t.Errorf("%s[%d] = %q fora da paleta ANSI e dos cinzas", v.Type().Field(i).Name, j, c)
				}
			}
			continue
		}
		if c := f.String(); !re.MatchString(c) {
			t.Errorf("%s = %q fora da paleta ANSI e dos cinzas", v.Type().Field(i).Name, c)
		}
	}
}

func TestCorSoNoTema(t *testing.T) {
	re := regexp.MustCompile(`"#[0-9a-fA-F]{6}`)
	filepath.Walk("../..", func(p string, i os.FileInfo, err error) error {
		if err != nil || i.IsDir() || !strings.HasSuffix(p, ".go") ||
			strings.HasSuffix(p, "_test.go") || filepath.Base(p) == "tema.go" {
			return nil
		}
		b, _ := os.ReadFile(p)
		if re.Match(b) {
			t.Errorf("hex fora do tema em %s", p)
		}
		return nil
	})
}
