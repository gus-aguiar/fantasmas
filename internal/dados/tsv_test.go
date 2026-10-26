package dados

import (
	"reflect"
	"testing"
)

func TestTSV(t *testing.T) {
	got := TSV("a\tb\tc\td\te\n\n x\n")
	want := []Linha{{"a", "b", "c", "d", "e"}, {" x", "?", "", "", ""}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if len(TSV("")) != 0 {
		t.Fatal("esperava vazio")
	}
}
