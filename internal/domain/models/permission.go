package models

import "github.com/google/uuid"

type Permission struct {
	UUID    uuid.UUID
	Name    string
	AppUUID uuid.UUID
}
