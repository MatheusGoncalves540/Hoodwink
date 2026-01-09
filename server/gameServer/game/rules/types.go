package rules

import (
	"fmt"
	"reflect"
)

// GameRules representa as regras do jogo carregadas do arquivo YAML
type GameRules struct {
	Cards    map[string]CardRules `yaml:"Cards"`
	Timeouts map[string]Timeouts  `yaml:"Timeouts"`
}

// CardRules representa as regras das cartas no jogo configuradas no arquivo YAML
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

// TimeoutsTypes representa os diferentes tipos de timeout configuráveis para cada sala
type TimeoutsTypes string

const (
	TimeoutsTypeDefault     TimeoutsTypes = "DEFAULT"
	TimeoutsTypeSuddenDeath TimeoutsTypes = "SUDDEN_DEATH"
)

// Timeouts representa os diferentes tipos de timeout possiveis para cada tipo de estado de sala
// como DEFAULT e SUDDEN_DEATH, cada um com seus próprios valores
type Timeouts struct {
	DisplayMessage *int `yaml:"displayMessage"`
	WaitingAction  *int `yaml:"waitingAction"`
}

// Get retorna o valor do timeout pelo nome do campo dinamicamente
func (t *Timeouts) Get(field string) (int, error) {
	// usa reflection para obter o valor do campo dinamicamente
	val := reflect.ValueOf(t).Elem().FieldByName(field)
	// verifica se o campo existe e não é nulo
	if !val.IsValid() {
		return 0, fmt.Errorf("campo '%s' não existe", field)
	}
	if val.IsNil() {
		return 0, fmt.Errorf("campo '%s' não configurado", field)
	}
	// retorna o valor do campo como int
	return int(val.Elem().Int()), nil
}
