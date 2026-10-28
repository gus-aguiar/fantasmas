#!/usr/bin/env python3
"""Converte ANSI (SGR truecolor) do stdin em HTML <pre> para a olhada renderizada."""
import argparse
import html
import re
import sys

SGR = re.compile(r"\x1b\[([0-9;]*)m")
OUTRO = re.compile(r"\x1b\[[0-9;?]*[A-Za-ln-z]")


def estilo(fg, bg, neg, dim):
    partes = []
    if fg:
        partes.append("color:" + fg)
    if bg:
        partes.append("background:" + bg)
    if neg:
        partes.append("font-weight:bold")
    if dim:
        partes.append("opacity:.7")
    return ";".join(partes)


def converte(texto, fundo="#1a1326"):
    texto = OUTRO.sub("", texto)
    out, pos = [], 0
    fg = bg = None
    neg = dim = False
    aberto = False

    def reabre():
        nonlocal aberto
        if aberto:
            out.append("</span>")
            aberto = False
        e = estilo(fg, bg, neg, dim)
        if e:
            out.append('<span style="%s">' % e)
            aberto = True

    for m in SGR.finditer(texto):
        out.append(html.escape(texto[pos:m.start()]))
        pos = m.end()
        p = [int(x) if x else 0 for x in m.group(1).split(";")] if m.group(1) else [0]
        i = 0
        while i < len(p):
            c = p[i]
            if c == 0:
                fg = bg = None
                neg = dim = False
            elif c == 1:
                neg = True
            elif c == 2:
                dim = True
            elif c == 22:
                neg = dim = False
            elif c == 39:
                fg = None
            elif c == 49:
                bg = None
            elif c in (38, 48) and i + 4 < len(p) and p[i + 1] == 2:
                cor = "#%02x%02x%02x" % tuple(p[i + 2 : i + 5])
                if c == 38:
                    fg = cor
                else:
                    bg = cor
                i += 4
            i += 1
        reabre()
    out.append(html.escape(texto[pos:]))
    if aberto:
        out.append("</span>")
    corpo = "".join(out)
    return (
        "<!doctype html><html><head><meta charset='utf-8'><title>quadro</title></head>"
        "<body style='margin:0;background:%s'>"
        "<pre style=\"margin:0;padding:12px;display:inline-block;background:%s;color:#f4eeff;"
        "font-family:'JetBrains Mono','DejaVu Sans Mono',monospace;font-size:14px;line-height:1.2\">%s</pre>"
        "</body></html>" % (fundo, fundo, corpo)
    )


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--fundo", default="#1a1326")
    a = ap.parse_args()
    sys.stdout.write(converte(sys.stdin.read(), a.fundo))
