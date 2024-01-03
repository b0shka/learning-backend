package identity

import "github.com/google/uuid"

type Generator interface {
	GenerateUUID() uuid.UUID
}

type IDGenerator struct{}

func NewIDGenerator() Generator {
	return &IDGenerator{}
}

func (g *IDGenerator) GenerateUUID() uuid.UUID {
	return uuid.New()
}
