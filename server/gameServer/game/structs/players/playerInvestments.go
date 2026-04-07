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
