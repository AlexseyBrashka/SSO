package auth

import (
	"errors"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")
var ErrTokenExpired = errors.New("invalid refresh token")
