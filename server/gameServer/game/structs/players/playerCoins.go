package players

// AddCoins adiciona moedas ao jogador
func (p *Player) AddCoins(amount int, maxCoins int) bool {
	// breakLimit indica se o limite de moedas será ultrapassado
	breakLimit := false
	if p.Coins+amount > maxCoins {
		breakLimit = true
	}

	p.Coins += amount

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
