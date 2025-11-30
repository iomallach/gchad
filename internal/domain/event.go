package domain

type ApplicationEvent interface {
	Event()
}

type UserLeftRoom struct{}

func (ulr *UserLeftRoom) Event() {}

type UserJoinedRoom struct{}

func (ujr *UserJoinedRoom) Event() {}
