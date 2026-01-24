package rules

import (
	"fmt"
	"reflect"
)

// GameRules representa as regras do jogo carregadas do arquivo YAML
type GameRules struct {
	Cards    map[string]CardRules `yaml:"Cards"`
	Generals Generals             `yaml:"General"`
	Timeouts map[string]Timeouts  `yaml:"Timeouts"`
}

// CardRules representa as regras das cartas no jogo configuradas no arquivo YAML
type CardRules struct {
	Value               *int  `yaml:"value,omitempty"`       // Valores afetados pelas taxas (nem sempre são os preços das cartas)
	FixedValue          *int  `yaml:"fixedValue,omitempty"`  // Valores fixos (não afetados por taxas)
	TaxMinimum          *int  `yaml:"taxMinimum,omitempty"`  // Valores mínimos após aplicação de taxas
	TaxMaximum          *int  `yaml:"taxMaximum,omitempty"`  // Valores máximos após aplicação de taxas
	CanKillSelf         *bool `yaml:"canKillSelf,omitempty"` // Indica se a carta pode eliminar a si mesma
	DoubledValueCard    *bool `yaml:"doubledValueCard,omitempty"`
	MaxDoubled          *int  `yaml:"maxDoubled,omitempty"`
	AffectedByTaxes     *bool `yaml:"affectedByTaxes,omitempty"`
	RoundsUntilDecrease *int  `yaml:"roundsUntilDecrease,omitempty"`
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

// Generals representa regras gerais do jogo configuradas no arquivo YAML
type Generals struct {
	InitialCoins *int `yaml:"initialCoins"`
	MaxPlayers   *int `yaml:"maxPlayers"`
	MaxCoins     *int `yaml:"maxCoins"`
}
