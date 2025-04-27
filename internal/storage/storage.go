package storage

import "errors"

var (
	ErrUserExists            = errors.New("user already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrPermExists            = errors.New("permission alredy exists")
	ErrPermNotFound          = errors.New("permission not found")
	ErrAppExists             = errors.New("app alredy exists")
	ErrAppNotFound           = errors.New("app not found")
	ErrUserPermissionsExists = errors.New("user permissions alredy exists")
	ErrCantGrantPermission   = errors.New("app cant grant this permission")
	ErrNoSuchRefreshToken    = errors.New("no such refresh token")
	ErrNoSuchUserPermission  = errors.New("no such user-permission")
	ErrNoPermissionsAtApp    = errors.New("no permissions at app")
)
