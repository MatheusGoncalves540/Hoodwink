package endpointStructures

type CreateRoomRequest struct {
	RoomName string `json:"roomName" validate:"required,max=30,min=3"`
	Password string `json:"password" validate:"omitempty,max=24"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
	Msg    string `json:"message"`
}
