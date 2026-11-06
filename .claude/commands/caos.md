---
description: Testa o sistema como um usuário caótico e registra as falhas
argument-hint: <alvo: tela, fluxo ou URL> [--rodadas N]
---
Alvo: $ARGUMENTS (padrão: fluxos principais da aplicação, 30 rodadas)

Aja como um usuário impaciente que não lê nada e cujo objetivo é quebrar o alvo. Registre tudo como um testador.

## Ataques
Alterne entre eles. Não repita o mesmo ataque mais de 3 vezes.
- Entrada: campo vazio, só espaços, 10 mil caracteres, emoji, acentos, `'"<script>`, `; DROP`, números negativos, zero, 1e309, vírgula vs ponto decimal, datas impossíveis (31/02), CPF/CNPJ inválidos, texto colado com quebra de linha.
- Ritmo: clique duplo ou múltiplo em salvar/confirmar, enviar antes de a tela carregar, Enter repetido, voltar no meio do fluxo, fechar e reabrir, duas abas ou instâncias no mesmo registro.
- Fluxo: pular etapas, abrir uma URL ou tela diretamente, cancelar no meio, derrubar a rede ou o servidor no meio, repetir uma operação já concluída.
- Estado: sessão expirada, dado alterado em outra aba e depois salvo nesta.

## Como executar
- Prefira automação a clicar na mão. Escreva um script (Playwright ou similar) que rode as rodadas e capture erros de console, respostas 4xx/5xx e exceções. Gaste tokens escrevendo o script, não observando tela por tela.
- Salve o script em `tests/caos/` para que ele possa rodar de novo sem você.
- Pare ao completar as rodadas ou ao encontrar 5 falhas graves.
- Use só a base de teste. Nunca apague dados reais.
- Cliente (P21): o sistema muda por cliente. Suba com `task dev CLIENTE=<id>` (padrão: boi-do-forte), ataque só módulos que esse cliente tem (`task clientes`) e diga o cliente no relatório. Ao terminar, rode `task modules:generate` (sem CLIENTE) para não deixar registries de outro cliente no disco.

## Saída
Salve em `plans/relatorios_anteriores/caos-<alvo>-<AAAA-MM-DD>.md` (crie a pasta se faltar; nome ocupado leva sufixo `-2`). Compacte antes de salvar: máx. 40 linhas; na tabela só falha reproduzível (nada de linha "ok"), o que passou em 1 linha agrupada, sem log cru nem passo intermediário. Limitação só se mudar a leitura do resultado. Rodada limpa: 5 linhas (alvo, script, rodadas, resultado).

Responda só com a tabela:

| # | Sev | Onde | Ação | Esperado | Obtido | Reproduzir |

Severidade:
- crítica: perde dado, trava ou calcula dinheiro errado
- alta
- média
- baixa: só cosmético

Escreva as falhas críticas e altas no fim do arquivo como tarefas `- [ ]` no formato dos planos, prontas para o /executar.
