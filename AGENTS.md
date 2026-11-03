# AGENTS.md: fantasmas

Guia para agentes que trabalham neste repositório.

## O que é

`fantasmas` é uma TUI em Go (Bubble Tea + Lip Gloss) que acompanha os agentes de um **workspace Spark**:
- os "gengars" (coordenadores) e os "fantasminhas" (agentes delegados pelo `/delegate`);
- as perguntas que eles fazem;
- missões, pipelines, PRs, CI, serviços locais e a máquina.

O binário se chama `fantasmas`, e o comando instalado no PATH se chama `gengar`.

O app **não tem dados próprios**. A cada coleta (padrão: 30 s) ele roda ferramentas externas e lê arquivos do workspace (veja "De onde vêm os dados"). Se uma fonte falta, a gaveta correspondente fica vazia e a gaveta `mais` lista o que falta (`dados.FaltaDe`).

## Setup

```bash
./setup.sh
```

O script é idempotente. Ele faz quatro coisas:
1. **Go:** garante Go >= a versão do `go.mod`. Usa o Go do PATH; se não houver, instala pelo `mise`; sem mise, baixa o tarball oficial para `~/.local/share/go`, sem sudo.
2. **Build:** compila `bin/fantasmas` com `GOTOOLCHAIN=local`.
3. **Comando:** cria `~/.local/bin/gengar` como **symlink** para `bin/fantasmas`.
4. **PATH:** põe `~/.local/bin` no PATH via `~/.zshrc` ou `~/.bashrc`, se faltar.

`GENGAR_BIN_DIR` e `GENGAR_GO_DIR` mudam os diretórios.

**Plataformas:** Linux e macOS. Windows nativo não funciona (veja "Armadilhas"); no Windows, use WSL.

**Não troque o symlink por cópia.** O botão atualizar (`U 2×`) usa `dados.AcharRepo`, que resolve o symlink do executável até `<repo>/bin/`. Depois roda `git pull --ff-only`, recompila e reabre o processo com `syscall.Exec`.

## Uso

```bash
gengar                       # TUI, coleta a cada 30 s, raiz = diretório atual
gengar --raiz ~/spark        # aponta para o workspace Spark
gengar --intervalo 1m
```

A raiz precisa ser a do workspace Spark, não este repositório. É lá que ficam `.delegate/`, `shared/` e `scripts/`.
- **Ordem:** `--raiz` manda; sem ela vale `FANTASMAS_RAIZ` (definível na aba settings, aceita `~/`); sem as duas, a pasta atual (`resolverRaiz` em `main.go`).
- Com a raiz errada, as gavetas que dependem dos scripts do Spark ficam vazias. A máquina, por exemplo, mostra "sem leitura da máquina".

### Modo quadro único

Este é o modo para agentes e testes visuais. Ele imprime um quadro e sai, sem TTY:

```bash
bin/fantasmas --uma-vez --demo --gaveta pipes --largura 140 --altura 42 --cor
bin/fantasmas --uma-vez --demo-verde --gaveta prs
bin/fantasmas --uma-vez --demo --gaveta missoes --missao <slug>
bin/fantasmas --uma-vez --demo --gaveta pipes --pipe 13
```

- `--demo` e `--demo-verde` usam um estado fixo (`ui.EstadoDemo`) e não precisam de workspace.
- Sem `--demo`, o comando faz uma coleta real.
- Gavetas válidas: `missoes agora crew pipes prs ci ambiente maquina mais settings`. Um nome inválido sai com código 2.
- Para ver um tema sem mexer no arquivo do usuário: `FANTASMAS_AJUSTES=/tmp/a.json FANTASMAS_TEMA=orca bin/fantasmas --uma-vez --cor --demo --gaveta settings`.

### Teclas principais

A lista completa fica no `?` dentro do app, e a fonte é `ajudaLinhas` em `internal/ui/model.go`.

- `← →`/`h l`/`1-9`: troca de gaveta; `,` abre settings. `↑ ↓`/`j k`: escolhe um item. `enter`: abre ou responde. `esc`: volta.
- `/`: filtra. `0`: volta a todas as missões. `r`: coleta agora. `q`: sai.
- `o`: abre o link. `g`: foca a aba no herdr. `p`: abre o painel do delegate (`http://127.0.0.1:7480/`).
- Ações destrutivas exigem a tecla duas vezes (estado "armado"):
  - `x 2×`: arquiva a pergunta;
  - `R 2×`: retoma um fantasminha caído;
  - `U 2×`: atualiza o app;
  - `A`/`S`/`T 2×`: avança, pula ou faz retry de uma etapa do pipe.
- Gaveta pipes: `n` lança um run, `e` edita o YAML e `N` abre o gengar dedicado do run.
- Outras: `D` pede uma rodada do DayDreaming e `M` reconecta os MCPs.

## De onde vêm os dados

A coleta é `dados.Coletor.Coletar` (`internal/dados/coleta.go`), e tudo nela roda em paralelo. Os caminhos são relativos a `--raiz`.

| Fonte | Uso |
|---|---|
| `shared/mods/control-panel/scripts/delegate.py` (`tasks`, `esperando`, `opcoes-todas`) | tasks e perguntas do delegate (modo padrão) |
| `scripts/delegate/ddb.py` (`import`, `state --json`, `tick --dry-run`) | substitui o delegate.py quando `FANTASMAS_FONTE=ddb`; liga as missões e os pipes |
| `shared/mods/control-panel/scripts/{local.sh,coleta.sh worktrees,maquina.sh}` | serviços, worktrees e máquina. Sem o `maquina.sh` na raiz, roda a cópia embutida `assets/maquina.sh` (memória, carga, sessões); a trava de pesados só entra com `.delegate/` na raiz. Mantenha a cópia em dia com a do Spark |
| `shared/skills/delegate-painel/painel.py` | entrega das respostas (sem ele, a resposta só é gravada) |
| `.delegate/` (incluindo `_mural.jsonl`) | tasks, mural e daydreamer |
| `herdr agent list` | agentes vivos e focar aba |
| `gh api graphql` | PRs e CI do usuário logado no `gh` |
| `~/.claude` (ou `FANTASMAS_CLAUDE_DIR`) | transcripts para a "linha viva" e o contexto de cada agente |

As saídas em TSV passam por `dados.TSV`, e a do ddb passa por `dados.DeDdb`. As duas viram o mesmo `[]Linha`.

### Aba settings e variáveis de ambiente

A gaveta **settings** (`,`) lista as variáveis `FANTASMAS_*` do catálogo `config.Opcoes` (`internal/config/config.go`) pelo nome real, cada uma com o valor que vale agora (o padrão, quando nada foi definido), e as salva em `~/.config/fantasmas/ajustes.json`. `FANTASMAS_IMAGEM`, `FANTASMAS_IMAGEM_MODO` e `FANTASMAS_SPRITE` ficam fora da aba, só pelo shell.
- o arquivo tem permissão 0600, numa pasta 0700, e fica fora do repo; `FANTASMAS_AJUSTES` muda o caminho;
- `main` aplica o arquivo ao ambiente **antes de qualquer `os.Getenv`**, então o resto do código continua só lendo variáveis;
- **o shell manda:** uma variável já definida no shell não é sobrescrita nem editável pela aba.

Regras para quem mexer:
- **Opção nova:** acrescente em `config.Opcoes`, com grupo, descrição, `Valores` (escolhas fixas; vazio = texto livre), `Reabrir` se ela só for lida na abertura e `Padrao` com o valor que vale sem nada definido, como deve aparecer na tela.
- **Texto livre definido nunca vai para a tela:** a lista, a barra de edição (eco `•`), os toasts e as mensagens de erro mostram só a máscara. O padrão da opção aparece, porque é do app e não do usuário. Não escreva o valor definido de uma opção `Livre()` em lugar nenhum visível.

| Variável | Efeito |
|---|---|
| `FANTASMAS_TEMA=roxo\|orca\|terminal` | cores do app (`ui.UsarTema`); padrão roxo; troca na hora |
| `FANTASMAS_FIGURA=<nome>` | figura do topo, do catálogo de figuras (ver abaixo); padrão gengar; troca na hora |
| `FANTASMAS_RAIZ` | raiz do workspace Spark quando não há `--raiz`; aceita `~/`; vale ao reabrir |
| `FANTASMAS_FONTE=ddb` | usa o `ddb.py`, liga as gavetas missões e pipes e abre em `missoes` |
| `FANTASMAS_DDB_SCRIPT` | caminho alternativo do `ddb.py` |
| `FANTASMAS_EDITOR` | editor para arquivos e pastas (links continuam no `xdg-open`/`open`) |
| `FANTASMAS_SPRITE` / `--sprite` | caminho de GIF ou PNG; manda sobre `FANTASMAS_FIGURA` |
| `FANTASMAS_IMAGEM=kitty\|celulas` | força o desenho da figura. Sem ela, o modo é detectado pelo terminal (WezTerm, kitty, herdr, Orca via `ORCA_IMAGE_PROTOCOL` ou consulta ao protocolo kitty) |
| `FANTASMAS_IMAGEM_MODO=animacao\|quadros\|T` | modo do protocolo kitty (padrão: `T` no herdr e no Orca, `animacao` fora). O Orca não aceita `a=f`/`a=a`, então a animação nativa fica parada lá |
| `FANTASMAS_MURAL=off` | esconde o mural |
| `FANTASMAS_CTX_JANELA` | tamanho da janela de contexto usada no % (padrão: 1 000 000) |
| `FANTASMAS_CLAUDE_DIR`, `FANTASMAS_CACHE_DIR` | diretório do Claude e de cache |
| `FANTASMAS_DEBUG=<arquivo>` | log de detecção do terminal e do envio de imagens |
| `FANTASMAS_AJUSTES` | caminho do arquivo da aba settings (só pelo shell) |
| `HERDR_ENV=1` + `HERDR_WORKSPACE_ID` | necessários para criar a aba do daydreamer |

### Figuras do topo

Cada GIF ou PNG de duas pastas vira uma figura, com o nome do arquivo sem extensão (`ui.NomesFiguras`):
- `assets/figuras/`: embutida no binário, exige recompilar (hoje: `gengar`, `lapras`, `entei`);
- `~/.config/fantasmas/figuras/`: sem recompilar; vence a embutida de mesmo nome.

O PNG estático vira um GIF de um quadro (`ui.paraGIF`). Trocar a figura com o app aberto remonta a imagem kitty (`Grafico.Recarregado`). Veja `assets/figuras/LEIA.md`.

## Estrutura

```
main.go                 flags, ligação das Acoes, escolha do modo de imagem, re-exec após atualizar
setup.sh                instalação (Go, build, symlink gengar, PATH)
assets/                 figuras/ embutidas (go:embed; gengar.gif é a padrão) e o brief do daydreamer
internal/config/        ajustes da aba settings: catálogo de opções, arquivo 0600, aplicação ao ambiente
internal/dados/         coleta e ações sobre o mundo externo (sem UI)
  runner.go             Runner: interface de execução de comandos; testes usam runners falsos
  coleta.go             Coletor e FaltaDe
  fonte_ddb.go pipes.go modo ddb: estado, tick, runs e pipelines
  responder.go          responder e arquivar perguntas (painel, ddb ou arquivo)
  atualizar.go          AcharRepo e Atualizar (pull, build, troca atômica de bin/fantasmas)
  botoes.go             daydreamer, reconectar MCPs
  ...                   prs, resumoci, vivos, caidos, missoes, mural, linhaviva, atividade
internal/ui/            Model do Bubble Tea, gavetas, detalhe, tema, sprite
  model.go              Model, Update, teclas, ajuda, Render
  moldura.go            lista de Gavetas, topo e rodapé
  ajustes.go            gaveta settings
  tema.go               temas (roxo, orca, terminal) e UsarTema
  figuras.go            catálogo de figuras, UsarFigura, PNG → GIF
  gavetas.go missoes.go pipes.go detalhe.go ...   uma tela por arquivo
  demo.go               EstadoDemo / EstadoDemoVerde
internal/kitty/         protocolo de gráficos do kitty (detecção e envio)
ferramentas/            quadro.sh + ansi2html.py: renderizam uma gaveta da demo em HTML (shots/)
```

**Arquitetura:**
- `dados` não conhece `ui`.
- A `ui` recebe uma função de coleta e um `ui.Acoes`, que é o conjunto de callbacks com efeitos colaterais montado em `main.go`. Isso mantém o `Model` testável.
- Comandos externos passam sempre por `dados.Runner`, para os testes poderem trocá-lo por fakes.

## Desenvolvimento

```bash
GOTOOLCHAIN=local go build -o bin/fantasmas .
GOTOOLCHAIN=local go test ./...     # ~20 s; a ui é a mais lenta
go vet ./...
```

- Use `GOTOOLCHAIN=local`, igual ao setup e ao atualizar, para não baixar outro toolchain.
- Os testes não precisam de rede, workspace, `gh` nem `herdr`. Eles usam runners falsos e `internal/dados/testdata/`.
- Depois de mexer na UI, confira visualmente com `bin/fantasmas --uma-vez --demo --gaveta <g>`. Há também testes de renderização em `internal/ui/*_test.go`.
- `bin/` e `shots/` estão no `.gitignore`.

## Convenções

- **Tudo em português:** identificadores, comentários, mensagens de UI e de erro. Mantenha isso.
- Comentário de função no formato `// Nome faz X`, curto, explicando o porquê quando não é óbvio.
- Commits no estilo Conventional Commits em português, com escopo: `feat(fantasmas): …`, `fix(fantasmas): …`, `feat(ui): …`.
- Ações que mudam estado externo pedem confirmação dupla na UI (padrão `armado*` em `model.go`). Siga o mesmo padrão em ações novas.
- Prefira degradar: se uma fonte falta, a gaveta fica vazia e aparece uma linha no `FaltaDe`, em vez de erro fatal.

## Armadilhas

- **Windows não compila:**
  - `internal/dados/botoes.go` usa `syscall.SysProcAttr{Setsid}`, que só existe em Unix;
  - `main.go` usa `syscall.Exec` para reabrir após atualizar.
- **Caminhos que supõem o repo dentro do Spark.** O app costuma viver em `<spark>/shared/tui/fantasmas`, e alguns pontos dependem disso:
  - o brief do daydreamer é lido de `<raiz>/shared/tui/fantasmas/assets/daydreamer-brief.md`;
  - `ferramentas/quadro.sh` supõe `spark = ../../..` e usa `scripts/orca-orchestration/heavy-run.sh`.
- `TestPrints` (`internal/ui/detalhe_test.go`) só roda com `FANTASMAS_PRINTS=1` e grava num caminho absoluto de outra máquina.
- O botão atualizar recusa rodar com mudanças locais no repo e só faz `pull --ff-only`.
