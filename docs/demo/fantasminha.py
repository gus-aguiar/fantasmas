# Gera fantasminha.gif: figura própria (sem arte de terceiros) usada na gravação da demo.
from PIL import Image

ARTE = [
    "....######....",
    "..##########..",
    ".############.",
    ".##WW####WW##.",
    "###WP####WP###",
    "###WP####WP###",
    "##############",
    "##############",
    "####.####.####",
    "##############",
    "##############",
    "##############",
    "##.###..###.##",
    "#...##..##...#",
]
# Paleta indexada: o índice 0 é o fundo transparente.
PALETA = [(0, 0, 0), (157, 108, 255), (255, 255, 255), (40, 20, 70)]
COR = {"#": 1, "W": 2, "P": 3}
ESC = 6

def quadro(dy):
    img = Image.new("P", (len(ARTE[0]) * ESC, (len(ARTE) + 2) * ESC), 0)
    img.putpalette([c for rgb in PALETA for c in rgb])
    for y, linha in enumerate(ARTE):
        for x, c in enumerate(linha):
            if c in COR:
                for i in range(ESC):
                    for j in range(ESC):
                        img.putpixel((x * ESC + i, (y + dy) * ESC + j), COR[c])
    return img

quadros = [quadro(d) for d in (0, 1, 2, 1)]
quadros[0].save("fantasminha.gif", save_all=True, append_images=quadros[1:], duration=220, loop=0, disposal=2, transparency=0)
