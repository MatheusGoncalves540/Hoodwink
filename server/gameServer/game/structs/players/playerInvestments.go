package players

// HasActiveInvestment verifica se o jogador tem algum investimento ativo
func (p *Player) HasActiveInvestment() bool {
	return len(p.Investments) > 0
}

// ClearInvestments limpa todos os investimentos do jogador (chamar quando o jogador morrer)
func (p *Player) ClearInvestments() {
	p.Investments = []int{}
}

// AddInvestment adiciona investimentos ao jogador
func (p *Player) AddInvestment(gracePeriod int) {
	p.Investments = append(p.Investments, gracePeriod)
}

// CountdownInvestment decrementa os investimentos do jogador, removendo os que chegam a zero (chamar a cada rodada para atualizar os investimentos)
func (p *Player) CountdownInvestment(amount int) {
	// Loop para percorrer os investimentos do jogador
	for i := 0; i < len(p.Investments); i++ {
		// Decrementa o investimento pelo valor do amount
		p.Investments[i] -= amount
		// Verifica se o investimento expirou
		if p.Investments[i] <= 0 {
			// Remove o investimento expirado
			p.Investments = append(p.Investments[:i], p.Investments[i+1:]...)
			// Decrementa o índice para verificar o próximo investimento corretamente
			i--
		}
	}
}
