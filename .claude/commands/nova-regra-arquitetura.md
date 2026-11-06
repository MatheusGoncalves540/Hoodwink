---
description: Transforma um pedido de regra em padrão aplicado no repositório inteiro (doc + código + checagem + erros comuns)
argument-hint: <regra em uma frase> [--auditar]
---
Regra: $ARGUMENTS

Você é o Arquiteto. Entregar a regra não é fazer um lugar obedecer: é registrá-la, aplicá-la em todo ponto que ela alcança e deixar uma checagem que falha sozinha quando alguém a violar. Não anuncie etapas; fale no final.

`--auditar`: só inventaria e propõe o padrão; não altera código nem docs.

## 0. Classificar
- Não é regra de arquitetura (bug, feature, refactor pontual)? Diga em 1 linha e pare; o caminho é o /planejar.
- Reescreva o pedido como regra testável: "todo X deve Y; exceção Z; violação é W". Se não der para escrever W, a regra ainda não está pronta: pergunte.
- É padrão novo, formalização do que já se pratica ou mudança de padrão existente? Leia `CLAUDE.md`, `tech-docs/arquitetura/padroes.md` e o doc do domínio tocado. Contradição com decisão registrada: pare e pergunte; não sobrescreva.
- Lacuna que muda escopo, exceção ou tratamento do legado: pergunte (máx. 5, fechadas, recomendada primeiro). Padrão óbvio vira premissa registrada, não pergunta.

## 1. Inventariar (só leitura)
- `graphify query "<regra>"` quando o grafo existir; `grep` por cada forma de fazer hoje (helper, componente, campo, literal, cópia). Levante `arquivo:linha` de TODOS os pontos, por app e módulo.
- Sem o número — N pontos em M arquivos — não prossiga. Regra sem inventário é meia regra.
- Multi-cliente (P21): a regra vale para TODO cliente. Inventarie o catálogo inteiro (todos os módulos, não só os do build atual) e rode `task clientes` para saber quem sente cada ponto. Regra que só faz sentido para um cliente não é de arquitetura: pergunte.
- Escolha o LUGAR ÚNICO pela camada: server compartilhado → `apps/server/internal/<tema>` (nunca em `modules/`); client compartilhado → `src/lib` ou `src/components` (fora de `src/modules/` quando mais de um módulo usa); primitivo de UI → `components/ui`. `modules/registry.*` é gerado: nunca editar.
- Regra que vale em Go e TS exige espelho (padrão `derive.go`/`derive.ts`), com a mesma regra e teste dos dois lados.
- Suba a escada: constante/enum/tipo/helper existente antes de tabela nova, abstração nova ou dependência. Se a checagem por máquina não for viável, diga por quê e registre como convenção no doc.

## 2. Projetar (máx. 15 linhas, antes de codar)
- Nome curto + ID estável, próximo livre da família (`Pnn`, `Unn`…): confira o maior em uso e siga o padrão irmão mais próximo em `padroes.md`.
- Alcance explícito: client, server, migrations, docs — e o que fica de fora.
- API canônica em 2 a 5 linhas: assinatura, props, campos, atributos.
- Exceção: como se declara e com motivo (`// pnn:ok <motivo>` ou allowlist no check). Sem motivo, não vale.
- Migração: ordem, o que muda de comportamento e como provar que não mudou.
- Até 15 arquivos: faça agora. Acima disso, ou mudando contrato/banco: lotes de 1 commit; passando de 2 lotes, gere plano em `./plans` (formato /planejar) e pare.

## 3. Revisar falhas (pré-mortem)
- Leia "Bugs conhecidos — não repetir" (`CLAUDE.md`) e os testes dos caminhos do inventário.
- Liste de 3 a 10 falhas CONCRETAS desta regra, não riscos genéricos: `cenário → sintoma → como detectar → prevenção`. Passe só pelos que couberem: dado antigo/migração, duplicata/duplo clique, concorrência, estado parcial/rollback, data e fuso, permissão, cache/token velho, contrato server↔client, falso negativo da checagem, caminho fora da tela (API direta, import/export, job).
- Cada falha termina em teste/checagem automatizada, verificação manual com comando/passos, ou decisão de escopo registrada. Nada fica só na prosa.
- Descarte o que não é específico desta regra: tabela de obviedade é ruído.

## 4. Implementar (nesta ordem)
1. O mecanismo canônico no lugar único; nenhuma cópia por módulo.
2. Migre todos os pontos do inventário. Corrija na função compartilhada, não em cada chamador. Legado só fica com dívida explícita: quantos, por quê, até quando.
3. Mude comportamento só se a regra exigir — e prove com teste antes/depois.
4. Lógica pura nova: função exportada + `node:test` (padrão `permission-matrix.test.ts`). Contrato/DTO segue `TESTING.md`.
5. Checagem em `scripts/checks/check-<slug>.mjs`: `arquivo:linha`, exit 1, `--report` para dívida, `findViolations` exportada e fixture positiva/negativa em `.test.mjs` (molde: `check-ui-text.*`). Registre no `task check:padroes` (`Taskfile.yml`) e no script `test` do `package.json`.
6. Docs: seção em `padroes.md` com **Regra / Motivo / Como se faz / Garantia automática**; 1 a 3 bullets no `CLAUDE.md`; `TESTING.md` só se criar obrigação nova; doc do módulo se o comportamento mudou.
7. **Erros comuns** no doc de cada módulo tocado (`tech-docs/modulos/<modulo>.md`), em seção própria no fim: `## Erros comuns e como verificar`, tabela `Regra | Cenário | Sintoma | Como detectar | Prevenção`. Essa seção é o lugar de consulta antes de mexer no módulo. Só as linhas daquele módulo e só com o teste ou passo que detecta. No `padroes.md`, 1 linha apontando para a seção. Regra de plataforma (sem módulo): a tabela fica no `padroes.md` § `<regra> — erros comuns`.

## 5. Provar
- `task check:padroes && task typecheck && task test && task lint && task build`; `task ts:test:e2e` se tocar UI ou contrato.
- `task build CLIENTE=<id>` para cada cliente de `config/clientes/` e, por último, `task modules:generate` (sem CLIENTE) antes do commit.
- Plante uma violação, rode a checagem, confirme exit 1 e reverta. Checagem que nunca falha não é garantia.
- Rode as verificações manuais viáveis da tabela de erros comuns; se alguma não der para rodar, diga por quê.
- Releia o diff contra a tabela: cada linha aponta um teste/checagem que existe e roda.
- `graphify update .` no fim. Nunca edite teste nem relaxe check existente para passar.

## Saída
**Regra:** nome (ID), 1 linha
**Lugar único:** `caminho` + API
**Migrado:** N pontos em M arquivos · exceções: N (motivo em 1 linha cada)
**Garantia:** check/teste + o comando que falha
**Erros comuns:** N falhas (X automatizadas, Y manuais) · docs: `tech-docs/modulos/<modulo>.md` § Erros comuns
**Docs:** seções alteradas

**Validação:** comandos verdes
**Dívida:** o que ficou fora e por quê
**Decidir:** perguntas fechadas de uma vez, se houver

Em `--auditar`, troque o bloco por: inventário (`arquivo:linha | o que viola | correção`) + desenho do padrão + tabela de erros comuns. Nada editado.

## Regras
- Uma regra por execução. Duas regras, duas execuções.
- O menor diff que cumpre a regra. Sem abstração, dependência ou boilerplate além dela.
- Sem estado misto e sem "faço o resto depois" silencioso.
- Não enfraqueça, ignore ou comente checagem existente. Conflito real entre regras: pare e pergunte.
- Commit por lote em `tipo(escopo): descrição`, sem push.
