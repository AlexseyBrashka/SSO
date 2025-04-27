package models

import "github.com/google/uuid"

type AuthApp struct {
	UUID   uuid.UUID
	Name   string
	Secret string
}

func NewApp(UUID uuid.UUID, name string, secret string) *AuthApp {
	if len(secret) == 0 || name == "" || UUID == uuid.Nil {
		return nil
	}
	return &AuthApp{
		UUID:   UUID,
		Name:   name,
		Secret: secret,
	}
}

type App struct {
	UUID uuid.UUID
	Name string
}
