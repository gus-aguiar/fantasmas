# fantasmas Strategy

## Vision

<!-- vision: author-approved 2026-11-26 -->

`fantasmas` existe para que uma pessoa coordene uma frota de agentes de código sem abrir um terminal por agente e sem perder o controle do que chega à main.
Ele serve quem roda vários agentes ao mesmo tempo num workspace de orquestração.
Ele transforma o estado espalhado em tasks, transcripts, PRs, CI e máquina numa tela só, ordenada pelo que precisa de atenção humana.
Ele é dono de exatamente uma coisa: a resposta para "o que precisa de mim agora, e o que está andando sem mim".

### Atenção primeiro, volume depois

A primeira tela é sempre a fila de atenção: esperando você, depois travados, depois trabalhando, depois parados.
As decisões chegam uma por vez, com a recomendação do agente e a regra que motivou a dúvida, para a pessoa decidir rápido sem perder o contexto.
Contador zerado não aparece, e o que é antigo fica recolhido a um enter.
Uma informação entra na tela quando muda o que a pessoa faz a seguir; o resto fica no detalhe ou na aba do agente.
A pessoa entende o que cada agente faz pela linha viva, sem entrar na aba dele.

### O humano é o gateway do merge

Todo merge passa pela pessoa: a TUI nunca libera nem executa um merge, nem por opção, nem por regra, nem com veto depois.
Um critério mecânico de pronto (PR aprovado, homologação, CI verde) prepara o merge e nunca o substitui.
O que o sistema decide sozinho aparece como decisão tomada, com veto à mão.
A autonomia cresce por regra explícita e visível no workspace, como a pista rápida, e cobre planos e etapas, nunca o merge.
Toda ação com efeito fora da tela pede a tecla duas vezes, inclusive aprovar a recomendação de um agente: a segunda tecla é a pausa para ler.

### Ler, não ser a fonte

O fantasmas não guarda dados de trabalho próprios: a cada coleta ele lê as ferramentas e os arquivos do workspace.
Histórico, lições e resumos vêm de registros que o workspace grava; a TUI agrega e mostra.
Quem agenda e executa é o workspace; a TUI dispara comandos que existem fora dela, com a tecla duas vezes e mostrando antes o que vai acontecer.
Lançar agentes de verdade pela TUI é um desses disparos, não uma fila própria.
Avisar fora do terminal é trabalho de quem entrega mensagens, não da TUI.
Quando uma fonte falta, a gaveta fica vazia e a tela diz o que falta, sem quebrar o resto.

### Pequeno, calmo e seguro de expor

Cor só para estado e foco; a estrutura vem de seções curtas, colunas e cinco abas.
Uma aba ou comando entra quando alguém usa, e sai quando deixa de ser usado.
Responder é escolher uma opção ou escrever uma frase curta; conversa longa é na aba do agente.
Valor de texto livre definido pelo usuário nunca aparece na tela, e o shell manda sobre o arquivo de ajustes.
Código, demo e testes não carregam nome de empresa, repo interno, dado de trabalho nem arte de terceiros.
Rotinas de uma pessoa ou organização entram como configuração, e quem já usa mantém o mesmo comportamento pelos ajustes.
A personalidade fica (fantasminhas, figura animada, o comando `gengar`), com arte própria no binário padrão e a do usuário fora dele.
O `--demo` roda sem nenhuma dependência e é como alguém de fora conhece o produto.

### Feito para o terminal de quem usa

Roda em Linux e macOS, e no Windows pelo WSL.
A figura usa o protocolo de gráficos do kitty quando o terminal aceita e cai para caracteres quando não.
Lê o agente que o workspace usa; outro agente entra por um adaptador de leitura quando o workspace passar a usá-lo, sem mudar as telas.
Instalar e atualizar é idempotente, sem sudo, e nunca sobrescreve trabalho local.
Toda mudança que alguém percebe usando entra no PATCH_NOTES no mesmo commit, escrita para quem usa.

### Scope

O fantasmas não é um orquestrador: não decide o que roda nem em que ordem.
Não é multiplexador de terminal, sistema de CI, bot de merge, serviço de notificação nem painel web.
Não é cliente de chat com os agentes.
Não guarda banco próprio de histórico.

Uma mudança se alinha quando encurta o caminho até a pessoa saber o que precisa dela, mantém o merge e o irreversível com ela, lê de uma fonte que o workspace já grava e cabe nas cinco abas sem ruído.
Uma mudança deve ser resistida quando faz qualquer merge acontecer sem a pessoa, tira a segunda tecla de uma ação com efeito, cria uma fonte de verdade na TUI, mostra segredo, dado de trabalho ou arte de terceiros, ou adiciona tela que ninguém usa.
