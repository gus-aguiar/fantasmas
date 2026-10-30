package dados

import (
	"os"
	"path/filepath"
	"testing"
)

const herdrAgentes = `{"id":"cli:agent:list","result":{"agents":[
{"agent":"claude","agent_session":{"agent":"claude","kind":"id","source":"herdr:claude","value":"sess-1"},"agent_status":"working","name":"c-a","pane_id":"w1:p1","tab_id":"w1:t1"},
{"agent":"claude","agent_session":null,"agent_status":"idle","name":"gengar-1","pane_id":"w1:p2","tab_id":"w1:t2"},
{"agent":"x"}]}}`

func TestAgentesDe(t *testing.T) {
	m := AgentesDe([]byte(herdrAgentes))
	if len(m) != 2 {
		t.Fatal(m)
	}
	a := m["c-a"]
	if a.Status != "working" || a.Sessao != "sess-1" || a.Pane != "w1:p1" || a.Aba != "w1:t1" {
		t.Error(a)
	}
	if g := m["gengar-1"]; g.Sessao != "" || g.Status != "idle" {
		t.Error(g)
	}
	if AgentesDe([]byte("lixo")) != nil {
		t.Error("lixo deveria dar nil")
	}
}

func escreve(t *testing.T, caminho, conteudo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSessaoPorIdVenceNome(t *testing.T) {
	dir := t.TempDir()
	escreve(t, filepath.Join(dir, "1.json"), `{"sessionId":"id-velho","name":"c-x","cwd":"/a","status":"idle","updatedAt":1000}`)
	escreve(t, filepath.Join(dir, "2.json"), `{"sessionId":"id-novo","name":"c-x","cwd":"/b","status":"busy","updatedAt":9000}`)
	escreve(t, filepath.Join(dir, "3.json"), `lixo`)
	m := SessoesClaude(dir)
	if len(m) != 2 {
		t.Fatal(m)
	}
	if s := sessaoDe(m, "c-x", "id-velho"); s == nil || s.Cwd != "/a" {
		t.Error(s)
	}
	if s := sessaoDe(m, "c-x", ""); s == nil || s.Cwd != "/b" || s.Status != "busy" {
		t.Error("sem id, vence a mais recente pelo nome", s)
	}
	if sessaoDe(m, "outro", "") != nil {
		t.Error("nome sem sessão")
	}
}

func TestMetasDe(t *testing.T) {
	raiz := t.TempDir()
	escreve(t, filepath.Join(raiz, ".delegate", "s1", "meta"), "slug=s1\nname=c-um\n")
	escreve(t, filepath.Join(raiz, ".delegate", "s2", "meta"), "name=c-dois\n")
	if err := os.MkdirAll(filepath.Join(raiz, ".delegate", "s3"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := MetasDe(raiz)
	if len(m) != 2 || m["c-um"] != filepath.Join(raiz, ".delegate", "s1") {
		t.Error(m)
	}
}

func TestCaminhoTranscript(t *testing.T) {
	got := CaminhoTranscript("/cl", "/home/gustavo/Documents/PROLOG/prolog-claude", "abc")
	want := "/cl/projects/-home-gustavo-Documents-PROLOG-prolog-claude/abc.jsonl"
	if got != want {
		t.Error(got)
	}
	if g := CaminhoTranscript("/cl", "/a/b.c_d", "s"); g != "/cl/projects/-a-b-c-d/s.jsonl" {
		t.Error(g)
	}
}
