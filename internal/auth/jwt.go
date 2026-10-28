package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultTTL = 24 * time.Hour

// Issuer signs and verifies the HS256 access tokens of the service.
type Issuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type Token struct {
	Value     string
	ExpiresAt time.Time
}

func NewIssuer(secret string, ttl time.Duration) (*Issuer, error) {
	if secret == "" {
		return nil, errors.New("auth: jwt secret is required")
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Issuer{secret: []byte(secret), ttl: ttl, now: time.Now}, nil
}

func (i *Issuer) Issue(userID int64) (Token, error) {
	now := i.now()
	expiresAt := now.Add(i.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return Token{}, fmt.Errorf("auth: sign token: %w", err)
	}
	return Token{Value: signed, ExpiresAt: expiresAt}, nil
}

// Parse verifies the signature and the expiration and returns the user id from the subject.
func (i *Issuer) Parse(raw string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, i.key,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(i.now),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, fmt.Errorf("auth: parse token: %w", err)
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("auth: token subject is not a user id")
	}
	return id, nil
}

func (i *Issuer) key(*jwt.Token) (any, error) {
	return i.secret, nil
}
