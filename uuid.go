package virtual_security

import "github.com/google/uuid"

func NewUUIDGenerator() UUIDGenerator {
	return &uuidGenerator{}
}

type UUIDGenerator interface {
	Generate() string
}

type uuidGenerator struct{}

func (u *uuidGenerator) Generate() string {
	return uuid.NewString()
}
