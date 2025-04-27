package models

import "github.com/google/uuid"

type User struct {
	UUID        uuid.UUID
	Email       string
	PassHash    []byte
	Permissions map[string]bool
}
