package players

// HasActiveInvestment verifica se o jogador tem algum investimento ativo
func (p *Player) HasActiveInvestment() bool {
	return len(p.Investments) > 0
}

// ClearInvestments limpa todos os investimentos do jogador (chamar quando o jogador morrer)
func (p *Player) ClearInvestments() {
	p.Investments = []int{}
	p.PendingInvestmentCoin = false
}

// AddInvestment adiciona investimentos ao jogador
func (p *Player) AddInvestment(gracePeriod int) {
	p.Investments = append(p.Investments, gracePeriod)
}

// TrimInvestments descarta investimentos excedentes, sem devolver as moedas.
// Usado quando as taxas diminuem e o limite máximo de investimentos cai.
// Retorna a quantidade de investimentos descartados.
func (p *Player) TrimInvestments(max int) int {
	if max < 0 {
		max = 0
	}
	if len(p.Investments) <= max {
		return 0
	}

	discarded := len(p.Investments) - max
	p.Investments = p.Investments[:max]
	p.PendingInvestmentCoin = false
	return discarded
}
