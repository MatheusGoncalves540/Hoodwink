package rooms

import "sort"

// IsConnected informa se o jogador está no conjunto Connected (ordenado) da sala.
func (r *Room) IsConnected(playerId string) bool {
	for _, id := range r.Connected {
		if id == playerId {
			return true
		}
	}
	return false
}

// SetConnected substitui o conjunto Connected pelos ids recebidos, em ordem.
func (r *Room) SetConnected(ids []string) {
	r.Connected = append([]string(nil), ids...)
	sort.Strings(r.Connected)
}

// AddConnected adiciona o jogador ao conjunto Connected, mantendo a ordem.
func (r *Room) AddConnected(playerId string) {
	if r.IsConnected(playerId) {
		return
	}
	r.Connected = append(r.Connected, playerId)
	sort.Strings(r.Connected)
}

// RemoveConnected remove o jogador do conjunto Connected.
func (r *Room) RemoveConnected(playerId string) {
	for i, id := range r.Connected {
		if id == playerId {
			r.Connected = append(r.Connected[:i], r.Connected[i+1:]...)
			return
		}
	}
}
