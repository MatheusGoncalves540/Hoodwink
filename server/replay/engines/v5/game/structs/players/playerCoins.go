package players

// AddCoins adiciona moedas ao jogador respeitando o limite máximo.
// Quando o limite é ultrapassado, o excedente é desconsiderado (não fica com o jogador)
// e o retorno indica que uma carta deve ser perdida.
func (p *Player) AddCoins(amount int, maxCoins int) bool {
	if amount <= 0 {
		return false
	}

	breakLimit := p.Coins+amount > maxCoins

	p.Coins += amount
	if p.Coins > maxCoins {
		p.Coins = maxCoins
	}

	return breakLimit
}

// RemoveCoins remove moedas do jogador
func (p *Player) RemoveCoins(amount int) {
	if amount > p.Coins {
		p.Coins = 0
		return
	}
	p.Coins -= amount
}
