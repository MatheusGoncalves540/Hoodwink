---
name: texto-claro
description: Revisa e escreve todo texto que o jogador do Hoodwink lê (copy.ts, Como se joga, falas do Proso, avisos, botões). Linguagem simples, sem jargão do jogo sem explicação e sem cara de texto feito por IA. Use antes de criar ou mudar qualquer texto de tela.
---

# Texto claro

Quem lê é alguém que nunca jogou e está no celular, com pressa. Se precisar ler duas vezes, o texto falhou.

## 1. Palavras do jogo

Use a palavra comum. Se a palavra do tema for necessária, explique na primeira vez que ela aparecer na tela.

| Não use (sem explicar) | Use |
|---|---|
| ator, atores | jogador, jogadores |
| papel, máscara (a carta) | personagem. "Cada carta é um personagem." |
| declarar um papel | usar um personagem / dizer que tem um personagem |
| contestar, contestação | duvidar. "Blefe!" pode aparecer como o grito, explicado |
| plateia (os outros jogadores) | os outros jogadores |
| em cena, sua deixa | na sua vez |
| camarote | lugar na mesa (no texto; o nome visual pode ficar) |
| diário da cena | histórico da partida |
| baralho morto | pilha de descarte |
| espectador | sai da partida e só assiste |
| espetáculo | partida |
| elenco (lista de personagens) | personagens |

O tema de teatro fica no visual, nos títulos curtos e na voz do Proso. Regra e instrução usam as palavras da tabela.

Termos que o jogo precisa ter, explique uma vez, numa frase:
- **Blefar**: dizer que tem um personagem sem ter a carta.
- **Carta morta**: carta virada para cima. Não vale mais.
- **Taxa**: número de −5 a +5 que soma no preço das ações.

## 2. Sinais de texto feito por IA (corte todos)

- Contraste montado: "não é X, é Y", "mais do que X, é Y".
- Listas de três por hábito ("rápido, simples e divertido").
- Travessão (—) para pausa dramática. Use ponto ou vírgula.
- Pergunta retórica para abrir seção ("Alguém vai perguntar?", "Pronto para a aventura?").
- Frase de efeito no fim do parágrafo ("A cortina espera.", "E é aí que a mágica acontece.").
- Palavras de propaganda: mergulhe, jornada, desvende, incrível, épico, experiência única, eleve, explore, universo.
- Título com dois-pontos ("Contestação: a arte de duvidar").
- Exclamação em sequência, emoji, "Vamos lá!".
- Adjetivo empilhado ("um jogo intenso, estratégico e envolvente").
- Ponto e vírgula. Quebre em duas frases.
- Voz passiva quando dá para dizer quem faz ("A carta é conferida" → "Ele mostra a carta").

## 3. Como escrever

- Frases curtas. Uma ideia por frase. Até ~20 palavras.
- "Você" para o jogador. Verbo no presente: "Você paga 6 e mata uma carta."
- Número em algarismo: "6 moedas", não "seis moedas" (exceto rótulo para leitor de tela).
- Comece pelo que a pessoa faz, depois a exceção.
- Botão diz a ação: "Criar sala", "Duvidar", não "Confirmar" ou "OK".
- Erro e aviso dizem o que fazer: "Faltam 2 moedas", não "Saldo insuficiente".

## 4. Revisão (rode antes de entregar)

1. Busque os termos da tabela e os sinais da seção 2:

   ```bash
   rg -n -i "\bator(es)?\b|declar|contest|plateia|em cena|deixa|camarote|diário|baralho morto|espectador|espetáculo|elenco|máscara|mergulh|jornada|desvend|incrível|épico|—|;" web/src/game/copy.ts
   ```

   Cada resultado: troque, ou confirme que está explicado na tela.
2. Leia em voz alta. Onde travar, reescreva.
3. Teste do amigo: alguém que nunca jogou entende o que fazer só com esse texto?
4. Conte as frases com mais de 20 palavras. Quebre.
5. Se mudar texto que um teste confere (`*.test.tsx` com `toContain`), atualize o teste no mesmo commit e diga no resumo.
