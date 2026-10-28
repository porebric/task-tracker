package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct {
	cost  int
	dummy []byte
}

func NewPasswordHasher(cost int) PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte("placeholder"), cost)
	if err != nil {
		dummy = nil
	}
	return PasswordHasher{cost: cost, dummy: dummy}
}

func (h PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

func (PasswordHasher) Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Burn spends the time of a real comparison, so a login for an unknown email takes as long
// as one with a wrong password and the response time does not reveal registered emails.
func (h PasswordHasher) Burn(password string) {
	_ = bcrypt.CompareHashAndPassword(h.dummy, []byte(password))
}
