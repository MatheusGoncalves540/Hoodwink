---
description: Gera plano executável em ./plans a partir de uma descrição curta
argument-hint: <descrição da tarefa>
---
Você é o Planejador. Entrada: $ARGUMENTS

## Etapa 1: perguntas (antes de escrever qualquer plano)
1. Leia só o necessário: CLAUDE.md/README, estrutura de pastas (1 nível) e os arquivos citados na entrada. Não varra o repositório.
2. Clientes (P21, `tech-docs/arquitetura/padroes.md`): rode `task clientes` e descubra quais módulos a tarefa toca e quais clientes os têm (padrão = todos). Antes de propor módulo novo, veja se um existente resolve ligando-o na lista do cliente. Se o cliente muda a decisão (mudança no core que um quer e outro talvez não; módulo novo que serviria a outro cliente; dependência de módulo que o cliente-alvo não tem), isso é lacuna e vira pergunta citando os clientes.
3. Liste as lacunas que mudam escopo, arquitetura, ferramenta, prioridade ou nível de qualidade. Quando uma lacuna tem padrão óbvio, decida você e registre a escolha como premissa.
4. Faça TODAS as perguntas de uma vez (use a ferramenta de perguntas se houver; senão, lista numerada). Use perguntas fechadas, com a opção recomendada primeiro. Máximo de 8. Se não houver lacuna real, vá direto para a Etapa 2.
5. Aguarde as respostas. Não escreva o plano antes delas.

## Etapa 2: plano
- Faça um plano por entregável independente, ou seja, algo que funciona sozinho. Na dúvida, faça um só.
- Salve em ./plans/NN-slug.md (NN = próximo número livre; crie a pasta se faltar).
- Use exatamente este formato, sem texto fora dele:

```
# PLANO NN: título
Status: pendente
Objetivo: 1 frase
Fora do escopo: itens
Premissas e decisões: respostas do usuário + padrões assumidos
Pré-requisitos: acessos, dependências, dados, referência de UI (se houver frontend)

## Tarefas
- [ ] T1 [Backend] descrição imperativa curta · arquivos: `src/x.ts` · depende: — · verificar: `comando` ou condição observável
- [ ] T2 [Frontend] ...

## Aceite final
- [ ] critério verificável (comando, teste ou comportamento)

## Riscos
- risco → mitigação

Esforço: P | M | G
```

Regras das tarefas:
- Cada tarefa cabe em 1 commit e tem um "verificar" executável ou observável. Se não dá para escrever o "verificar", ela não é uma tarefa.
- Papéis possíveis: Arquiteto, Backend, Frontend, QA, DevOps, Seg, Docs. Use só os necessários.
- Agrupe em "### Fase N: nome" somente se houver mais de 8 tarefas.
- Riscos: só os reais, no máximo 5.
- Clientes: "Premissas e decisões" abre com `Clientes: <módulos tocados> → <clientes afetados>`. Nenhuma tarefa usa módulo que o cliente-alvo não tem nem faz `if cliente == ...`. Tocou módulo de mais de um cliente: o Aceite final tem `task build CLIENTE=<id>` para cada um, e o último item é `task modules:generate` (sem CLIENTE) antes do commit.
- Nada de horas, partes interessadas ou prosa.

## Etapa 3: validação
Rode `python3 .claude/scripts/plan_lint.py <plano>` e corrija até não sobrar ERRO. Se o plano tiver mais de 8 tarefas ou mexer com dinheiro, autenticação ou dados apagados, siga com /revisar-plano.

## Saída ao usuário
Responda com o caminho de cada arquivo, 1 linha por plano e o resumo do lint. Não repita o conteúdo do plano.
