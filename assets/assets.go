package assets

import "embed"

// Figuras são as figuras do topo embutidas no binário: um GIF ou PNG por arquivo em figuras/.
// Para adicionar uma, ponha o arquivo lá e recompile; o nome dela é o do arquivo sem extensão.
//
//go:embed figuras
var Figuras embed.FS

// Gengar é a figura padrão.
//
//go:embed figuras/gengar.gif
var Gengar []byte
