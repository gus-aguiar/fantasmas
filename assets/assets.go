package assets

import "embed"

// Figuras são as figuras do topo embutidas no binário: um GIF ou PNG por arquivo em figuras/.
// Para adicionar uma, ponha o arquivo lá e recompile; o nome dela é o do arquivo sem extensão.
//
//go:embed figuras
var Figuras embed.FS

// Maquina é a cópia do maquina.sh do Spark, usada quando a raiz não tem o script: memória, carga e
// sessões são do computador, não do workspace. Roda com bash -c; $1 é a trava de testes pesados.
//
//go:embed maquina.sh
var Maquina string

// Gengar é a figura padrão.
//
//go:embed figuras/gengar.gif
var Gengar []byte
