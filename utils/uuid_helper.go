package utils

import (
	"github.com/google/uuid"
)

func ParseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func IsValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

func NewUUID() uuid.UUID {
	return uuid.New()
}
