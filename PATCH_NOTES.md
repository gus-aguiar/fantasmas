# Patch notes

O que mudou no `gengar` (TUI fantasmas) e no fluxo de missões e sidequests do Spark. A versão mais nova fica no topo.
Cada entrada diz o que muda para quem usa; os detalhes técnicos ficam no histórico do git.

## 2026-11-06 · Visual

**Muda**
- **Abas no rodapé, atalhos só no `?`:** a barra "1 Agora · 2 Missões · 3 PRs · 4 Máquina · 5 Config" volta para a última linha da tela, centralizada (com "? ajuda" em cinza à direita quando cabe); somem a linha de atalhos do topo e o rodapé de atalhos. `?` lista tudo, inclusive `ctrl+p`/`b`.
- **O topo voltou a respirar:** o gengar aparece em tamanho cheio à esquerda e, ao lado, "fantasmas ↻ atualizado HH:MM", o resumo ("1 precisa de você · 2 trabalhando", só o que é maior que zero) e o alerta da máquina quando há. Em janela com menos de 30 linhas o gengar some e o topo fica numa linha.
- **O gengar só se mexe quando algum fantasminha está trabalhando;** com tudo parado, ele fica quieto.
- **Paleta de comandos no lugar do painel lateral:** `ctrl+p` (ou `b`, como antes) abre uma caixa no centro da tela, com o resto esmaecido atrás. Digite para buscar; os comandos vêm agrupados (DIA: Abrir o dia, Fechar o dia · MULTAS: Saúde de multas · SISTEMA: DayDreaming, Reconectar MCPs, Atualizar app), com o estado à direita ("rodando há 4 min"). `enter 2×` executa, `g` vai à aba do disparo e `esc` fecha. Com a busca vazia, `b` também fecha.
- **Menos enfeite, mais leitura:** seções com cabeçalho cinza em caixa alta e contador ("ESPERANDO VOCÊ 1", "TRAVADOS 1", "TRABALHANDO 2", "PARADOS 1"); a cor aparece só para estado e foco.
- **Listas em colunas:** glifo de estado colorido (▲ pede você, ● trabalhando, ◌ parado, ✗ travado ou caído), o título na cor normal e os detalhes em cinza à direita, cortados com "…".
- **Acompanhando e antigos recolhidos numa linha só** ("▸ Acompanhando 3   ▸ Antigos 8"); `enter` (ou `a`) abre os dois.

## 2026-11-06 · Painel de botões

**Muda**
- **Os comandos saem da gaveta `5` para o painel lateral de botões**, que abre e fecha com `b` em qualquer gaveta (`esc` também fecha). Ele aparece à direita, por cima da lista, sem trocar de gaveta; em janela estreita ocupa a tela toda. Dentro: `/dia abrir`, `/dia fechar`, `/saude-multas`, DayDreaming, reconectar MCPs e atualizar app, com `↑ ↓`, `enter 2×` para disparar, "rodando há N min" e `g` para ir à aba. O rodapé mostra a dica `b botões`.
- **A gaveta `5` volta a ser configurações**, só com os ajustes (tema, figura e variáveis). `--gaveta configuracoes` abre nela, e `--gaveta comandos` segue valendo como apelido.

## 2026-11-06 · Cinco gavetas e comandos

**Muda**
- **As gavetas passam de 8 para 5**, nas teclas `1` a `5` (e `← →`): `1` agora, `2` missões, `3` PRs, `4` máquina, `5` comandos. O app abre sempre no agora, mesmo com `FANTASMAS_FONTE=ddb`. `?` mostra todas as teclas.
- **Agora** junta o que era "agora" e "fantasminhas": "precisa de você" é uma lista só, com o selo `decisão` ou `liberação`; abaixo ficam os travados e os fantasminhas trabalhando, parados e caídos recentes.
- **Missões** mostram cartas enxutas: nome, ▲ suas, o que o gengar faz agora e PR vermelho. `enter` abre a missão com cards, contagens, sidequests e decisões (veto incluso); `enter` de novo recolhe. Antes, `enter` levava à gaveta agora.
- **Máquina** junta máquina, ambiente e worktrees. Com tudo verde, fica numa linha que abre com `enter`; com alerta, já abre inteira.
- **PRs:** a linha dos rascunhos ocultos agora abre com `enter` (o `d` segue valendo).
- **Topo numa linha:** `▲ N suas · ● N trabalhando`, mais o alerta da máquina só quando existe. O gengar do topo tem no máximo 4 linhas e some em janela com menos de 30 linhas.
- `--gaveta` aceita os nomes novos (`agora missoes prs maquina comandos`); os antigos (`crew`, `ambiente`, `worktrees`, `configuracoes`) continuam valendo.

**Novo**
- **Gaveta comandos (`5`):** `/dia abrir` e `/dia fechar` abrem uma aba nova no workspace GENGAR do Herdr, e `/saude-multas` no workspace "saúde multas", já rodando o `claude` com o comando, na raiz do Spark. Cada um pede `enter 2×`, mostra "rodando há N min" e `g` vai até a aba. DayDreaming, reconectar MCPs e atualizar app também estão lá, e os ajustes (tema, figura, variáveis) ficam no fim.

**Sai**
- As dicas do topo e o rodapé de dicas das missões (a ajuda `?` cobre).
- Contadores zerados, como o "fantasminhas (0)" que contradizia a lista.
- A sublista de sessões do "reconectar MCPs": fica só o resumo da última rodada.
- Caídos de mais de um dia saem do destaque e vão para "antigos", a um `enter`.
- A tecla `0` (missões agora é `2`; `esc` tira o foco da missão).

## 2026-11-04 · Missões e sidequests

**Novo**
- **Missões substituem os pipes.**
  - A gaveta `0` é a das missões. Cada missão aparece como uma carta, com objetivo, cards, coordenador e contagens.
  - Os pipes agora se chamam **sidequests** e aparecem dentro da carta da missão, com um bloco colorido por etapa.
  - A gaveta `pipes` deixou de existir.
- **Sidequests soltas.**
  - Lançada sem missão em foco, a sidequest fica na seção "SIDEQUESTS SOLTAS" até acabar.
  - Com uma missão em foco, `n` lança dentro dela a partir de qualquer gaveta.
- **`m` abre uma missão nova** direto da TUI (enter 2× confirma).
- **"Decisões que tomei por você":** cada missão lista o que o sistema decidiu sozinho, com `v` para vetar (enter 2× confirma).
- **Pista rápida:** tarefa simples aprova o próprio plano sem esperar ninguém. Liberar o merge continua sempre com você.
- **Etapa `pronto`:** uma sidequest só termina com o PR aprovado pelo bot e a homologação feita no mesmo commit.
- **Atualização automática:** pelo `scripts/fantasmas.sh`, o `gengar` atualiza o Spark, as skills e este repo antes de abrir. `GENGAR_ATUALIZAR=off` desliga.
- **README e patch notes** neste repositório.

**Mudou**
- Teclas das gavetas: `0` missões, `1` agora, `2` fantasminhas, `3` PRs, `4` ambiente, `5` máquina, `6` worktrees, `7` configurações.
- O painel web do delegate mostra as mesmas missões, as sidequests em blocos e as decisões.

**Corrigido**
- O gengar não pisca mais no Herdr, nem o cursor e o mouse junto com ele, e continua animado. No Windows ou em terminal sem suporte a imagem, ele aparece desenhado com caracteres em vez de uma linha piscando.

## 2026-11-03 · Configurações e máquina

- **Aba configurações:** tema, figura do topo e variáveis `FANTASMAS_*` com o valor que vale agora. O valor que você define nunca aparece na tela.
- **Figuras do topo:** gengar, lapras e entei. Para usar uma figura sua, ponha o arquivo em `~/.config/fantasmas/figuras/`.
- **Gaveta máquina:** memória, swap, disco e pressão, lidos de qualquer pasta.
- **Gaveta worktrees.** PRs e CI ficam numa aba só.
- **`setup.sh`** instala o Go, compila, cria o comando `gengar` e acha o seu Spark.
- **Botões DayDreaming e "reconectar MCPs"** foram para a aba configurações.
- **Orca:** o gengar animado agora aparece no terminal do Orca.

## 2026-10-30 · Ver sem entrar

- **Linha viva por fantasminha:** o que ele está fazendo agora, o contexto usado e há quanto tempo.
- **Raias do "agora":** precisa de você, trabalhando e parados.
- **CI enxuto:** uma linha por PR, com os verdes resumidos numa linha só.
- **Mural** no topo, com os avisos entre os gengars.
- **Modo `ddb`** (`FANTASMAS_FONTE=ddb`): responder, arquivar e retomar um fantasminha caído (`R 2×`).
- **Botão atualizar** (`U 2×`): puxa, recompila e reabre o app.

## 2026-10-26 · Primeira versão

- Import inicial da TUI em Go (Bubble Tea + Lip Gloss), com a paleta roxa e o modo de quadro único (`--uma-vez`) para testes visuais.
