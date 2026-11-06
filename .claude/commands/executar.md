---
description: Executa um plano de ./plans até o fim
argument-hint: <caminho do plano> [--paralelo]
---
Execute o plano: $ARGUMENTS
Você implementa. Não anuncie o que vai fazer, faça. Fale comigo só no final ou se ficar bloqueado.

## Início
1. Leia o plano. O estado está nos checkboxes: [x] feito, [ ] pendente, [!] bloqueado. Não tente deduzir o estado varrendo o repositório.
2. Se estiver em main/master, crie a branch `plano/NN-slug` e trabalhe nela.
3. Leia o CLAUDE.md. Se houver tarefa de UI, leia a referência de UI do plano. Se não houver referência, leia os 2 componentes existentes mais parecidos com o que vai criar.
4. Clientes (P21): leia a linha `Clientes:` das Premissas. Sem ela, rode `task clientes`, descubra quem tem os módulos tocados e registre a linha nas Premissas.

## Loop por tarefa (em ordem, respeitando o "depende")
1. Leia só os arquivos listados na tarefa e o que eles importam diretamente.
2. Se a tarefa já estiver implementada e o "verificar" passar, marque [x] com "(já existia)" e siga.
3. Implemente o mínimo que satisfaz a tarefa. Nada de escopo extra nem de refatoração que não foi pedida.
4. Rode o "verificar" e o lint/typecheck/testes do módulo afetado. A suíte inteira roda só no aceite final.
5. Se passou: rode `task modules:generate` (sem CLIENTE; os registries do git são sempre os do cliente padrão), marque [x] no plano e faça um commit `tipo(escopo): descrição [TN]` incluindo o plano.
   Se falhou: corrija (até 3 tentativas). Depois disso, marque [!] com o motivo em 1 linha e siga com as tarefas que não dependem dela.

## Subagentes
Use somente com --paralelo ou quando houver 2 ou mais tarefas independentes e grandes, sem arquivos em comum.
Cada subagente recebe: o texto literal da tarefa, os caminhos, o comando de verificação e as convenções do projeto em até 3 linhas. Ele não lê o plano inteiro.
Para tarefas mecânicas, use o modelo mais barato disponível. Revise o diff de cada subagente antes de marcar [x].

## Regras
- Containers: podman, nunca docker.
- UI: reutilize componentes, tokens de cor e espaçamentos existentes. Não crie estilo inline novo nem cores soltas.
- Decisão não coberta pelo plano: escolha a opção mais conservadora e reversível, registre em "Premissas" e siga em frente.
- Multi-cliente (P21): nunca use módulo que o cliente-alvo não tem, nunca copie módulo por cliente nem escreva `if cliente == ...`; reaproveite ligando módulo existente. Se a decisão muda conforme o cliente e o plano não cobre (core que um quer e outro talvez não, dependência ausente), marque a tarefa [!] com a pergunta citando os clientes, siga com as outras e traga a pergunta no final. Tocou módulo de mais de um cliente: no aceite, `task build CLIENTE=<id>` para cada um.
- Pare e pergunte apenas antes de algo irreversível: apagar dados, migração destrutiva, push ou deploy.
- Não faça push nem merge.

## Relatório em arquivo
Salve em `plans/relatorios_anteriores/AAAA-MM-DD-<slug-do-plano>.md` (nome ocupado leva sufixo `-2`) ao encerrar, mesmo parcial.
Compacte antes de salvar: máx. 40 linhas, sem log cru, sem repetir o plano e sem o que não muda decisão futura. Só:
- plano, branch e status;
- concluído: 1 linha por grupo, não por tarefa;
- validação: comandos rodados e resultado em 1 linha;
- bloqueios: tarefa, erro real e o que falta;
- ressalvas e premissas que mudaram o resultado;
- erros corrigidos: causa raiz em 1 linha;
- dívida deixada, com `arquivo:linha` e como verificar.
Sem erro, bloqueio, ressalva nem dívida, não crie arquivo.

## Fim
Rode o "Aceite final" completo e atualize `Status:` (concluído ou parcial). Responda só com:

**Progresso:** X/Y tarefas (Z%) · branch `...`
**Feito:** 1 linha por tarefa ou grupo de tarefas
**Bloqueado:** tarefa — motivo — o que você precisa de mim
**Como usar:** comandos ou passos para rodar e testar o que foi entregue (máx. 6 linhas)
