package api

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	httpapi "github.com/porebric/healthstep-libs/api"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100

	maxEmailLen    = 255
	maxNameLen     = 255
	maxTitleLen    = 255
	maxTextLen     = 10000
	minPasswordLen = 8
	maxPasswordLen = 72

	msgRequired        = "required"
	msgPositiveInteger = "must be a positive integer"
	msgInvalidStatus   = "must be one of: todo, in_progress, done"
)

func msgTooLong(maxLen int) string {
	return fmt.Sprintf("must be at most %d characters", maxLen)
}

func requireText(field, value string, maxLen int) error {
	if strings.TrimSpace(value) == "" {
		return httpapi.Invalid(field, msgRequired)
	}
	if utf8.RuneCountInString(value) > maxLen {
		return httpapi.Invalid(field, msgTooLong(maxLen))
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return httpapi.Invalid("email", msgRequired)
	}
	if len(email) > maxEmailLen {
		return httpapi.Invalid("email", msgTooLong(maxEmailLen))
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return httpapi.Invalid("email", "must be a valid email address")
	}
	return nil
}
