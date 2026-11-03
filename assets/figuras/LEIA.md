# Figuras do topo

Cada GIF (animado ou não) ou PNG desta pasta vira uma figura na aba **settings** (`,` → figura).
O nome da figura é o nome do arquivo sem extensão, em minúsculas: `entei.png` → `entei`.

- **Aqui (`assets/figuras/`):** a figura entra no binário. Recompile com `./setup.sh` ou `go build -o bin/fantasmas .`.
- **Sem recompilar:** ponha o arquivo em `~/.config/fantasmas/figuras/` e reabra o fantasmas. Uma figura de lá vence a daqui que tiver o mesmo nome.

Dicas:
- Use pixel art com fundo transparente. No PNG, pixel com alfa abaixo de 50% vira transparente.
- O topo é pequeno (até 16×8 células), então sprites de até uns 100 px de lado ficam melhores.
- `gengar.gif` é a figura padrão; não renomeie.
