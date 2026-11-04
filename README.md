# fantasmas

Painel de terminal para acompanhar, numa janela só, o trabalho dos agentes de um workspace **Spark**. Ele mostra:

- as **missões**;
- as **sidequests** de cada missão;
- os **fantasminhas** (agentes que executam as tasks);
- as **perguntas** que esperam você;
- os **PRs** com o CI;
- os serviços locais e a máquina.

O comando se chama `gengar`.

O que mudou em cada versão está em [PATCH_NOTES.md](PATCH_NOTES.md).

## Instalar

Pré-requisitos:
- Linux ou macOS (no Windows, use o WSL);
- o workspace Spark clonado;
- `gh` logado;
- `herdr` no PATH.

```bash
git clone git@github.com:gus-aguiar/fantasmas.git <spark>/shared/tui/fantasmas
cd <spark>/shared/tui/fantasmas && ./setup.sh
```

O `setup.sh` faz quatro coisas, e pode ser rodado de novo sem problema:
1. instala o Go, se faltar;
2. compila o binário;
3. cria o comando `gengar` em `~/.local/bin`;
4. descobre onde está o seu Spark.

## Usar

```bash
gengar                          # abre o painel
FANTASMAS_FONTE=ddb gengar      # liga missões e sidequests (recomendado)
```

Toda vez que abre pelo `scripts/fantasmas.sh` do Spark, o `gengar` se atualiza antes: avança a `main` do Spark, ressincroniza as skills e traz a versão nova deste repo. Ele nunca sobrescreve trabalho local; para desligar, use `GENGAR_ATUALIZAR=off`.

As teclas que mais importam estão abaixo. A lista completa abre com `?`.

| Tecla | Faz |
|---|---|
| `0`–`7` ou `← →` | troca de gaveta: `0` missões, `1` agora, `2` fantasminhas, `3` PRs, `4` ambiente, `5` máquina, `6` worktrees, `7` configurações |
| `↑ ↓` e `enter` | escolhe e abre/responde |
| `m` | abre uma missão nova (enter 2× confirma) |
| `n` / `N` | lança uma sidequest (`N` abre também um gengar dedicado) |
| `v` | veta uma decisão que o sistema tomou sozinho (enter 2× confirma) |
| `g` | vai para a aba do agente no Herdr |
| `r` / `q` | atualiza agora / sai |

## Missões e sidequests

- **Missão** é para o que é direto. Cada missão é uma workspace do Herdr com um gengar coordenador, os cards do Jira e um objetivo. Na gaveta `0`, cada missão é uma carta.
- **Sidequest** é para o que tem vários passos encadeados, como implementar → homologar → pronto, ou um bugfix. Na carta, cada sidequest aparece como uma linha com um bloco por etapa:
  - verde: feita;
  - azul: rodando;
  - âmbar: espera você;
  - coral: falhou;
  - cinza: ainda não começou.
- **Onde a sidequest nasce:** com uma missão em foco, ela nasce dentro dessa missão. Sem foco, nasce **solta** e fica na seção "SIDEQUESTS SOLTAS" até acabar.
- **Pronto** quer dizer PR aprovado pelo bot e homologação no mesmo commit.
- **Autonomia:** tarefas simples passam pela "pista rápida" e aprovam o plano sozinhas. Tudo o que o sistema decide sozinho aparece em "Decisões que tomei por você", com a opção de vetar. **O merge é sempre seu:** a TUI nunca libera o merge.

### Disparar uma sidequest

1. Abra o `gengar` com `FANTASMAS_FONTE=ddb`.
2. Escolha onde ela nasce: aperte `enter` numa missão para focar nela, ou não foque nada para ela nascer solta.
3. Aperte `n`, escolha o fluxo (`implementar`, `implementar-homologar` ou `bugfix`), preencha o card e os repos e confirme.
4. A TUI só mostra o que *seria* lançado. Para lançar de verdade, rode na raiz do Spark:
   `DDB_TICK=on python3 scripts/delegate/ddb.py tick --executar`

## Contribuir

- Leia o [AGENTS.md](AGENTS.md): ele tem a arquitetura, as convenções e as armadilhas.
- Testes: `GOTOOLCHAIN=local go test ./...`.
- **Toda mudança que alguém percebe usando entra no [PATCH_NOTES.md](PATCH_NOTES.md), no mesmo commit.**
