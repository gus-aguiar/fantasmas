package dados

import "time"

type Travado struct {
	Repo   string
	Number int
	Motivo string
	Title  string
	URL    string
	Draft  bool
	Branch string
	CI     string
	Falhos []string
}

type Merge struct {
	Repo   string `json:"repo"`
	PR     int    `json:"pr"`
	URL    string `json:"url"`
	Base   string `json:"base"`
	Spark  string `json:"spark"`
	Checks string `json:"checks"`
}

type Recursos struct {
	GH, Herdr, Delegate bool
	Painel, Autor       string
}

type Estado struct {
	Tasks, Esperando, Opcoes, Worktrees, Maquina []Linha
	Servicos                                     []Linha
	Travados                                     []Travado
	Recursos                                     Recursos
	At                                           time.Time
	Raiz                                         string
	Vivos                                        map[string]bool
	Agentes                                      map[string]Agente
	Linhas                                       []LinhaViva
	Vistos                                       map[string]bool
	Daydreamer                                   Daydreamer
	MCP                                          ProgressoMCP
	Mural                                        []Aviso
}
