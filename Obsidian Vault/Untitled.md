- Chegar requisição sobre usar uma carta
- informar a todos que uma carta esta sendo usada:
	- debitCoins(playerId, 4 + room.tax) //Debita as moedas antes, pois sempre gasta as moedas, mesmo se revelar quue não tinha a carta e não executar a ação.
	- SendToAllPlayersNewDisplay({
		owner: playerId,
		payload: {
			target: targetPlayerId,
			targetCard: 0 (carta da direita)
		},
		displayTime: TempoMs
	 })
	- SetOnRedisDb(room:roomId.playInTimeOut, {})
	- setTimer(TempoMs, ExecutarPlayInTimeOut(roomId))

Sempre que uma carta for morta, no final tem que perguntar se a pessoa quer usar o kamikaze