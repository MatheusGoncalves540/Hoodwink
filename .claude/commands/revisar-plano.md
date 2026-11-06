---
description: Revisa um plano de ./plans (lint grátis + 1 revisor independente) e aplica as correções
argument-hint: <caminho do plano> [--profundo]
---
Revise o plano: $ARGUMENTS
Não anuncie etapas. Fale comigo só no final.

## 1. Checagem mecânica (custo zero)
Rode `python3 .claude/scripts/plan_lint.py <plano>`.
- Corrija você mesmo todo ERRO e todo AVISO de formato, caminho, dependência ou ordem. Rode de novo até zerar os ERROs.
- AVISO que exige julgamento (verificar vago, aceite sem comando, sem Docs, sem referência de UI) fica para o revisor.

## 2. Fatos que o revisor vai precisar (você coleta, ele não procura)
Junte em no máximo 25 linhas:
- saída final do lint;
- scripts do package.json (ou equivalente: Makefile, pyproject);
- árvore de 2 níveis das pastas citadas em "arquivos";
- trecho do CLAUDE.md com convenções (até 5 linhas);
- saída de `task clientes` (matriz módulo × cliente) e a regra P21 em 1 linha: core = todos os clientes sentem; opcional = só quem tem; reaproveitar antes de criar; nunca usar módulo que o cliente não tem nem `if cliente`.

## 3. Revisor independente
Despache UM subagente (modelo intermediário, ex.: sonnet). Ele não viu a conversa em que o plano nasceu, e é por isso que enxerga furos que o autor não vê. Envie só o texto do plano + os fatos do passo 2 + este pedido, sem resumir o plano:

> Você revisa um plano de implementação que outro agente vai executar sozinho, sem poder perguntar nada. Leia no máximo 5 arquivos do repositório, só para confirmar suspeitas específicas. Encontre problemas que fariam a execução falhar, desperdiçar trabalho ou entregar algo errado:
> - Furos de raciocínio: premissa não verificada, caso de borda ausente (vazio, duplicado, concorrência, falha no meio, dados antigos/migração), dinheiro/arredondamento, permissão, reversão.
> - Otimização: tarefa desnecessária, tarefa que deveria ser dividida ou juntada, ordem que bloqueia paralelismo, retrabalho entre tarefas.
> - Praticidade: "verificar" que não prova o objetivo da tarefa, aceite final que passa mesmo com a feature quebrada, pré-requisito que o executor não tem como obter.
> - Intuitividade: descrição ambígua que dois desenvolvedores implementariam diferente; fluxo de uso confuso para o usuário final.
> - Documentação: o que alguém precisa saber para usar ou manter o resultado e que nenhuma tarefa [Docs] cobre.
> - Clientes (matriz e regra nos fatos): falta a linha `Clientes:` nas Premissas; mudança em módulo padrão pensada para um cliente só; uso de módulo que o cliente-alvo não tem; módulo novo ou código copiado onde ligar um módulo existente resolveria; aceite sem `task build CLIENTE=<id>` para cada cliente afetado. Se a escolha depende do dono, use `DECISÃO:`.
> Regras: no máximo 10 achados, do mais grave ao menos grave. Não reescreva o plano. Não elogie. Se não achar nada sério, diga "sem achados graves".
> Formato, uma linha por achado:
> `SEV | tarefa ou seção | problema | correção exata (texto a inserir ou trocar)`
> SEV: BLOQ (execução falha ou entrega errada), ALTO (retrabalho provável), MÉDIO, BAIXO.
> Se a correção depende de decisão do dono do projeto, escreva `DECISÃO:` seguido de uma pergunta fechada com opções.

Com `--profundo`, ou se o plano tiver mais de 15 tarefas ou mexer com dinheiro, autenticação ou dados apagados: despache em paralelo um segundo revisor só com o tópico "furos de raciocínio" e outro só com segurança (validação de entrada, permissões, segredos). Mesmo formato e mesmos limites.

## 4. Aplicar
- Aplique no plano toda correção BLOQ, ALTO e MÉDIO que não seja DECISÃO. Ignore BAIXO, exceto se for uma troca de 1 linha.
- Descarte achado que contradiz uma resposta já registrada em "Premissas e decisões".
- Rode o lint de novo. Não pode sobrar ERRO.
- Adicione ao fim de "Premissas e decisões": `Revisado em AAAA-MM-DD: N correções aplicadas.`

## Saída
**Lint:** erros/avisos antes → depois
**Aplicado:** 1 linha por correção (tarefa: o que mudou)
**Decidir:** perguntas DECISÃO, todas de uma vez, com a opção recomendada primeiro (use a ferramenta de perguntas se houver). Ao receber as respostas, aplique-as e rode o lint mais uma vez.
