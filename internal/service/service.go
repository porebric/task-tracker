package service

import (
	"strings"
	"time"
)

func utcNow() time.Time {
	return time.Now().UTC()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
