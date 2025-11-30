package infrastructure

import "github.com/google/uuid"

type IdGen func() string

func UUIDGen() string {
	return uuid.NewString()
}
