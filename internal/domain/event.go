package domain

type ApplicationEvent interface {
	Event()
}

type UserLeftRoom struct {
	Name string
}

func NewUserLeftRoomEvent(name string) *UserLeftRoom {
	return &UserLeftRoom{
		Name: name,
	}
}

func (ulr *UserLeftRoom) Event() {}

type UserJoinedRoom struct {
	Name string
}

func NewUserJoinedRoomEvent(name string) *UserJoinedRoom {
	return &UserJoinedRoom{
		Name: name,
	}
}

func (ujr *UserJoinedRoom) Event() {}
