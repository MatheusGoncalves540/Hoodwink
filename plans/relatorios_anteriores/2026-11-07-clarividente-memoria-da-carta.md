# Relatório — 08-clarividente-memoria-da-carta

Plano: `plans/08-clarividente-memoria-da-carta.md` · branch `plano/08-clarividente-memoria-da-carta` · status: concluído

## Concluído
- T1: `LogEntry.reveal` guarda alvo, índice, carta e `showToAll` no `REVEALED_CARD`.
- T2: `selectPublicReveals` agrupa por índice com o turno mais recente e `selectSeatSheet` acrescenta a linha derivada, ordenada com o diário do ator.
- T3: bloco `role="note"` "Carta revelada no turno N" com o papel na `SeatSheet`, mais textos em `copy.ts`.
- T4: testes do caso público, do privado e da duplicata da mesma carta.

## Validação
- `npm test -- Sheets selectors log` verde; `npm test` verde (370 testes).

## Ressalvas
- Só a revelação com `showToAllPlayers: true` entra na memória, como decidido; a privada continua só no diário de quem jogou o Clarividente.
- Nada muda no estado da carta (`SheetCard` segue "escondida"), nem no protocolo.
