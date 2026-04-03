package rooms

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// AppendPendingEffect adiciona um efeito pendente à sala
func (r *Room) AppendPendingEffect(effect structs.Effect) error {
	dto, err := effect.ToDTO()
	if err != nil {
		return err
	}

	r.PendingEffects = append(r.PendingEffects, dto)
	return nil
}

// AppendPendingPresentationEvent adiciona um evento de apresentação pendente à sala
func (r *Room) AppendPendingPresentationEvent(event structs.PresentationEvent) error {
	dto, err := event.ToDTO()
	if err != nil {
		return err
	}

	r.PendingPresentationEvents = append(r.PendingPresentationEvents, dto)
	return nil
}

// PopLastPendingEffect remove e retorna o último efeito pendente
func (r *Room) PopLastPendingEffect() (structs.EffectDTO, bool) {
	if len(r.PendingEffects) == 0 {
		return structs.EffectDTO{}, false
	}

	lastIdx := len(r.PendingEffects) - 1
	effect := r.PendingEffects[lastIdx]
	r.PendingEffects = r.PendingEffects[:lastIdx]
	return effect, true
}

// PopNextPendingPresentationEvent remove e retorna o próximo evento de apresentação pendente (FIFO)
func (r *Room) PopNextPendingPresentationEvent() (structs.PresentationEventDTO, bool) {
	if len(r.PendingPresentationEvents) == 0 {
		return structs.PresentationEventDTO{}, false
	}

	first := r.PendingPresentationEvents[0]
	r.PendingPresentationEvents = r.PendingPresentationEvents[1:]
	return first, true
}

// HasPendingPresentationEvent verifica se há eventos de apresentação pendentes
func (r *Room) HasPendingPresentationEvent() bool {
	return len(r.PendingPresentationEvents) > 0
}

// HasPendingLogicEffect verifica se há efeitos lógicos pendentes
func (r *Room) HasPendingLogicEffect() bool {
	return len(r.PendingEffects) > 0
}
