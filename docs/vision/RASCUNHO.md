# Vision do fantasmas: rascunho para revisão

Rascunho do vision e 10 hipotéticos para o Gustavo responder.
Depois das respostas, o texto aprovado vai para `STRATEGY.md` (seção Vision), e esta pasta sai do repositório: ela é só a superfície de revisão, não um documento permanente.
O board interativo (`board.html`) tem o mesmo conteúdo; aberto direto no navegador, ele funciona, mas não envia as respostas a ninguém, então anote-as aqui ou no Discord.

## Como responder

Em cada hipotético, preencha a linha **Resposta** com uma de: **In vision** (o fantasmas deve fazer isso), **Off mission** (não deve) ou **Conditional** (só com uma condição), e o motivo.
O motivo é o que reescreve o texto: ele vale mais que o veredito.

## Rascunho atual

### Vision

`fantasmas` existe para que uma pessoa coordene uma frota de agentes de código sem abrir um terminal por agente e sem perder o controle do que chega à main.
Ele serve quem roda vários agentes ao mesmo tempo num workspace de orquestração.
Ele transforma o estado espalhado em tasks, transcripts, PRs, CI e máquina numa tela só, ordenada pelo que precisa de atenção humana.
Ele é dono de exatamente uma coisa: a resposta para "o que precisa de mim agora, e o que está andando sem mim".

#### Atenção primeiro, volume depois

A primeira tela é sempre a fila de atenção: esperando você, depois travados, depois trabalhando, depois parados.
Contador zerado não aparece, e o que é antigo fica recolhido a um enter.
Uma informação entra na tela quando muda o que a pessoa faz a seguir; o resto fica no detalhe ou na aba do agente.
A pessoa entende o que cada agente faz pela linha viva, sem entrar na aba dele.

#### O humano decide o que é irreversível

O merge é sempre humano: a TUI nunca libera nem executa um merge.
O que o sistema decide sozinho aparece como decisão tomada, com veto à mão.
Toda ação com efeito fora da tela pede a tecla duas vezes.
A autonomia cresce por regra explícita e visível, como a pista rápida, nunca por omissão.

#### Ler, não ser a fonte

O fantasmas não guarda dados de trabalho próprios: a cada coleta ele lê as ferramentas e os arquivos do workspace.
Quem agenda, executa e lança é o workspace; a TUI mostra, aponta e dispara comandos que existem fora dela.
Quando uma fonte falta, a gaveta fica vazia e a tela diz o que falta, sem quebrar o resto.

#### Pequeno, calmo e seguro de expor

Cor só para estado e foco; a estrutura vem de seções curtas, colunas e cinco abas.
Uma aba ou comando entra quando alguém usa, e sai quando deixa de ser usado.
Valor de texto livre definido pelo usuário nunca aparece na tela, e o shell manda sobre o arquivo de ajustes.
Código, demo e testes não carregam nome de empresa, repo interno nem dado de trabalho; o que é de uma organização vem de variáveis `FANTASMAS_*`.
O `--demo` roda sem nenhuma dependência e é como alguém de fora conhece o produto.

#### Feito para o terminal de quem usa

Roda em Linux e macOS, e no Windows pelo WSL.
A figura usa o protocolo de gráficos do kitty quando o terminal aceita e cai para caracteres quando não.
Instalar e atualizar é idempotente, sem sudo, e nunca sobrescreve trabalho local.
Toda mudança que alguém percebe usando entra no PATCH_NOTES no mesmo commit, escrita para quem usa.

#### Scope

O fantasmas não é um orquestrador: não decide o que roda nem em que ordem.
Não é multiplexador de terminal, sistema de CI, bot de merge nem painel web.
Não é cliente de chat com os agentes: conversar é na aba do agente.
Rotinas de uma pessoa ou organização entram como configuração, não como regra do produto.

Uma mudança se alinha quando encurta o caminho até a pessoa saber o que precisa dela, mantém o merge e o irreversível com o humano, lê de uma fonte que já existe e cabe nas cinco abas sem ruído.
Uma mudança deve ser resistida quando faz a TUI decidir ou executar sozinha algo irreversível, cria uma fonte de verdade própria, mostra segredo ou dado de trabalho, ou adiciona tela que ninguém usa.

## Hipotéticos

### H-1 · Auto-merge quando o pronto é inequívoco

**Proposta:** Uma variável opt-in (FANTASMAS_AUTOMERGE) faz a TUI mergear sozinha a sidequest que chegou a pronto: PR aprovado pelo bot, homologação no mesmo commit e CI verde. O merge aparece em 'Decisões que tomei por você', com veto que vira revert.

**Regra testada:** "O merge é sempre humano: a TUI nunca libera nem executa um merge."

**A favor:** o pronto já é um critério mecânico; com dezenas de PRs verdes, o humano vira carimbo e o gargalo é ele, e o veto continua lá.

**Contra:** merge é o único ponto que garante que alguém leu o que vai para a main; um revert depois do deploy não desfaz o efeito, e opt-in tende a virar padrão.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-2 · Aprovar a recomendação com uma tecla só

**Proposta:** Na fila de perguntas, cada decisão vem com a recomendação do agente. Uma tecla (a) aprova a recomendação de primeira, sem o 2×, porque responder é o que a pessoa mais faz e o 2× vira atrito na décima pergunta seguida.

**Regra testada:** "Toda ação com efeito fora da tela pede a tecla duas vezes."

**A favor:** responder não é destrutivo, o agente ainda executa depois, e a fila só funciona se for rápida.

**Contra:** aprovar uma recomendação pode ser 'pode apagar a branch' ou 'pode subir'; uma tecla só leva a aprovar no automático, e o 2× é justamente a pausa que obriga a ler.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-3 · Lançar a sidequest de verdade pela TUI

**Proposta:** Hoje o n só mostra o que seria lançado e a pessoa roda o tick --executar na raiz. A proposta é o n (com enter 2×) executar o tick de verdade, para lançar agentes sem sair da tela.

**Regra testada:** "Quem agenda, executa e lança é o workspace; a TUI mostra, aponta e dispara comandos que existem fora dela."

**A favor:** a TUI já dispara /dia abrir e a Faxina, que também têm efeito; o tick é só mais um comando do workspace, e o passo manual é atrito puro.

**Contra:** lançar agentes custa dinheiro e máquina; manter o lançamento fora da TUI deixa a TUI ser um painel seguro de abrir em qualquer lugar.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-4 · Avisar fora do terminal

**Proposta:** Quando algo entra em 'esperando você' e a TUI não está em foco, o fantasmas manda notificação de desktop ou uma mensagem num canal (Discord, ntfy). A pessoa pode sair de perto e volta só quando precisa.

**Regra testada:** "Ele é dono de exatamente uma coisa: a resposta para 'o que precisa de mim agora'" e "Não é ... painel web."

**A favor:** supervisão por exceção é exatamente isso; uma tela que precisa ser olhada não é exceção.

**Contra:** notificação é outra superfície com outro ciclo de vida (credencial, canal, spam), e o workspace pode avisar sozinho; a TUI ficaria responsável por entrega, não por leitura.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-5 · Guardar histórico próprio para resumo e lições

**Proposta:** Para os painéis 'o que foi feito hoje' e 'lições' (quantas vezes cada erro se repetiu), o fantasmas grava cada coleta num SQLite em FANTASMAS_CACHE_DIR e calcula as contagens dali.

**Regra testada:** "O fantasmas não guarda dados de trabalho próprios: a cada coleta ele lê as ferramentas e os arquivos do workspace."

**A favor:** o workspace não guarda séries no tempo; sem histórico, 'quantas vezes' não tem resposta, e um cache local é invisível para o usuário.

**Contra:** duas fontes de verdade divergem (apagou no workspace, ficou na TUI); o certo seria o workspace registrar as lições e a TUI ler.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-6 · Responder o agente com texto longo dentro da TUI

**Proposta:** Hoje, para conversar, a pessoa usa g e vai à aba do agente. A proposta é um editor de várias linhas na TUI para responder qualquer pergunta com contexto, anexar arquivo e reabrir a conversa, sem trocar de aba.

**Regra testada:** "Não é cliente de chat com os agentes: conversar é na aba do agente."

**A favor:** trocar de aba quebra o fluxo de fila, e responder já existe em versão curta; texto longo é só o mesmo gesto completo.

**Contra:** cada recurso de chat (histórico, anexos, edição) puxa outro, e a TUI vira um cliente de agente pior que o próprio terminal do agente.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-7 · A rotina do Gustavo sai do produto

**Proposta:** Abrir o dia, Planejar o dia e Fechar o dia (com Notion e o workspace GENGAR) saem do catálogo fixo e viram comandos configuráveis, como o FANTASMAS_CMD_EXTRA. O binário padrão vem com a paleta vazia, só com Faxina e swap.

**Regra testada:** "Rotinas de uma pessoa ou organização entram como configuração, não como regra do produto."

**A favor:** para alguém de fora, /plano-dia no Notion é ruído e expõe o jeito de trabalhar de uma pessoa; o mecanismo já existe.

**Contra:** o produto nasceu dessa rotina e é usado todo dia assim; transformar em config piora a experiência do único usuário real para agradar um visitante hipotético.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-8 · Gengar, Pokémon e o nome do comando

**Proposta:** Trocar as figuras embutidas por arte própria, renomear o comando de gengar para fantasmas e os 'gengars' coordenadores para 'coordenadores' em todo o produto, não só no demo.

**Regra testada:** "Código, demo e testes não carregam nome de empresa, repo interno nem dado de trabalho" e a identidade do produto.

**A favor:** o repo é público, sprite de Pokémon é IP de terceiros e 'gengar' não diz nada a quem chega.

**Contra:** a personalidade (fantasminhas, gengar animado) é parte do porquê o painel é agradável de deixar aberto o dia inteiro; o nome já está no PATH, nos scripts e no hábito.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-9 · Pista rápida para tarefas médias

**Proposta:** Hoje só tarefa simples aprova o próprio plano. A proposta é a pista rápida valer também para tarefas médias com testes verdes, aparecendo como decisão tomada e vetável, para a fila de 'esperando você' ficar só com o que é ambíguo.

**Regra testada:** "A autonomia cresce por regra explícita e visível, como a pista rápida, nunca por omissão."

**A favor:** é crescimento por regra explícita, exatamente o que o texto permite, e a fila menor é o objetivo.

**Contra:** cada degrau parece pequeno e o veto só funciona se alguém ler as decisões; com mais autonomia, a lista de decisões vira a nova fila que ninguém lê.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 

### H-10 · Outros agentes além do Claude

**Proposta:** Ler também transcripts e estado do Codex, do Gemini CLI e do opencode, para a linha viva e o contexto funcionarem com qualquer agente que o workspace use.

**Regra testada:** "Ele serve quem roda vários agentes ao mesmo tempo num workspace de orquestração."

**A favor:** 'frota de agentes' não deveria depender de um fornecedor, e o workspace já usa mais de um.

**Contra:** cada formato de transcript é um adaptador a manter e testar; ser excelente para um agente e um workspace é o que deixa o painel pequeno e confiável.

**Resposta:** _In vision / Off mission / Conditional_ + motivo: 
