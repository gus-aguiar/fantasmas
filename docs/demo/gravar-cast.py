# Grava fantasmas.cast (asciinema) com o mesmo roteiro do fantasmas.tape.
# Uso, da raiz do repo:  asciinema rec --headless --window-size 120x36 -c "python3 docs/demo/gravar-cast.py" docs/demo/fantasmas.cast
import fcntl, os, pty, select, struct, sys, tempfile, termios, time

RAIZ = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
ROTEIRO = [(4.0, b"2"), (3.5, b"3"), (3.5, b"4"), (2.5, b"5"), (3.0, b"1"), (2.5, b"\x10"), (4.0, b"\x1b"), (2.0, b"q")]

env = dict(os.environ, HOME=tempfile.mkdtemp(), FANTASMAS_SPRITE=os.path.join(RAIZ, "docs/demo/fantasminha.gif"))
sys.stdout.buffer.write(b"$ fantasmas --demo\r\n"); sys.stdout.flush()
pid, fd = pty.fork()
if pid == 0:
    os.execve(os.path.join(RAIZ, "bin/fantasmas"), ["fantasmas", "--demo"], env)

# Responde às consultas que o fantasmas faz ao terminal (cor de fundo, posição do cursor, atributos);
# sem resposta ele espera o tempo-limite. Sem resposta ao Kitty Graphics, cai no desenho em meio-bloco.
RESPOSTAS = [(b"\x1b]11;?", b"\x1b]11;rgb:1e1e/1e1e/2e2e\x1b\\"), (b"\x1b[6n", b"\x1b[1;1R"), (b"\x1b[c", b"\x1b[?62;22c")]

def responder(dados):
    for consulta, resposta in RESPOSTAS:
        for _ in range(dados.count(consulta)):
            os.write(fd, resposta)
# O pty novo nasce com 0×0; herda o tamanho do terminal da gravação.
colunas, linhas = os.get_terminal_size(sys.stdout.fileno())
fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", linhas, colunas, 0, 0))

def drenar(ate):
    while True:
        falta = ate - time.monotonic()
        if falta <= 0:
            return True
        r, _, _ = select.select([fd], [], [], falta)
        if r:
            try:
                dados = os.read(fd, 65536)
            except OSError:
                return False
            if not dados:
                return False
            responder(dados)
            sys.stdout.buffer.write(dados); sys.stdout.flush()

for espera, tecla in ROTEIRO:
    if not drenar(time.monotonic() + espera):
        break
    os.write(fd, tecla)
drenar(time.monotonic() + 1.5)
os.waitpid(pid, 0)
