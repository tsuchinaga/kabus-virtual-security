package virtual_security

import "github.com/google/uuid"

func newUUIDGenerator() iUUIDGenerator {
	return &uuidGenerator{}
}

type iUUIDGenerator interface {
	generate() string
}

type uuidGenerator struct{}

func (u *uuidGenerator) generate() string {
	return uuid.NewString()
}
