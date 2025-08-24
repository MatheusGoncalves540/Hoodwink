package endpointStructures

type CreateRoomRequest struct {
	RoomName   string `json:"roomName" validate:"required,max=30,min=3"`
	Password   string `json:"password" validate:"omitempty,max=24"`
	MaxPlayers int    `json:"maxPlayers" validate:"required,min=2,max=10"`
}

type CreateRoomResponse struct {
	RoomId string `json:"RoomId"`
	Msg    string `json:"message"`
}
