package service

import (
	"net/mail"
)

func IsEmailValid(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}

	return true
}
