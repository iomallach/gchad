package application

import "github.com/iomallach/gchad/internal/domain"

type ClientHandler interface {
	Id() string
	Send() chan domain.Messager
}
