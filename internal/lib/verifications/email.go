package verifications

import (
	"errors"
	"net/mail"
)

func VerifyEmail(email string) (bool, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return false, errors.New("invalid email")
	}
	return true, nil
}
