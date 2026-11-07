package rules

import "fmt"

// Validate checks the integrity of the GameRules.
func (r *GameRules) Validate() error {
	// TODO validar regras das cartas, ex: taxMin <= taxMax, ou canKillSelf só ser do kamikaze, e bool, etc
	if r.Cards == nil {
		return fmt.Errorf("ruleset has no cards defined")
	}

	return nil
}
