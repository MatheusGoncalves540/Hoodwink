package rules

type GameRules struct {
	Cards map[string]CardRules `yaml:"Cards"`
}

type CardRules struct {
	Price             *int  `yaml:"price,omitempty"`
	FixPrice          *int  `yaml:"fixPrice,omitempty"`
	TaxMinimum        *int  `yaml:"taxMinimum,omitempty"`
	TaxMax            *int  `yaml:"taxMax,omitempty"`
	AmountReceived    *int  `yaml:"amountReceived,omitempty"`
	AmountWithdrawn   *int  `yaml:"amountWithdrawn,omitempty"`
	InvestPrice       *int  `yaml:"investPrice,omitempty"`
	InvestedMaxAmount *int  `yaml:"investedMaxAmount,omitempty"`
	CanKillSelf       *bool `yaml:"canKillSelf,omitempty"`
}
