# Demo gravada

Gravação de ~26 s do `fantasmas --demo`: abas 1 → 2 → 3 → 4 → 5 → 1 e a paleta (`ctrl+p`).

| Arquivo | O que é |
|---|---|
| `fantasmas.gif` | 1280×740, 24 fps, para o README |
| `fantasmas.webm` | mesmo vídeo, VP9, para `<video>` na vitrine |
| `fantasmas.cast` | asciinema v3, 120×36, para `asciinema-player` |
| `fantasminha.gif` | figura própria do topo (gerada por `fantasminha.py`), no lugar das figuras de terceiros |

No `--demo` a interface chama o coordenador de "coordenador" e a figura padrão de "fantasma"; fora dele, nada muda.

Regravar, da raiz do repo (precisa de `vhs`, `ttyd`, `ffmpeg` e `asciinema` 3):

```sh
go build -o bin/fantasmas .
vhs docs/demo/fantasmas.tape
asciinema rec --headless --window-size 120x36 -c "python3 docs/demo/gravar-cast.py" docs/demo/fantasmas.cast
```
