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

Terminal sem suporte a imagem (ex.: Windows Terminal): o gengar aparece desenhado com caracteres; para forçar, `FANTASMAS_IMAGEM=celulas`.

As teclas que mais importam estão abaixo. A lista completa abre com `?`.

| Tecla | Faz |
|---|---|
| `1`–`5` ou `← →` | troca de gaveta: `1` agora (abre aqui), `2` missões, `3` PRs, `4` máquina, `5` comandos |
| `↑ ↓` e `enter` | escolhe e abre/responde; numa missão, abre as sidequests e as decisões |
| `enter 2×` | na gaveta `5`, dispara `/dia abrir`, `/dia fechar`, `/saude-multas`, DayDreaming, MCPs ou atualizar |
| `?` | ajuda com todas as teclas |
| `m` | abre uma missão nova (enter 2× confirma) |
| `n` / `N` | lança uma sidequest (`N` abre também um gengar dedicado) |
| `v` | veta uma decisão que o sistema tomou sozinho (enter 2× confirma) |
| `g` | vai para a aba do agente no Herdr (na gaveta `5`, para a aba do comando disparado) |
| `r` / `q` | atualiza agora / sai |

## Comandos (gaveta 5)

`/dia abrir` e `/dia fechar` abrem uma aba nova no workspace do Herdr chamado **GENGAR**; `/saude-multas`, no workspace **saúde multas**. A aba nasce na raiz do Spark com o `claude` já rodando o comando e se chama `dia-MMDD-HHMM` ou `saude-MMDD-HHMM`. Cada disparo mostra há quanto tempo roda, e `g` leva até a aba. Os ajustes (tema, figura e variáveis) ficam no fim da mesma gaveta.

## Missões e sidequests

- **Missão** é para o que é direto. Cada missão é uma workspace do Herdr com um gengar coordenador, os cards do Jira e um objetivo. Na gaveta `2`, cada missão é uma carta enxuta; `enter` abre as sidequests e as decisões dela.
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
