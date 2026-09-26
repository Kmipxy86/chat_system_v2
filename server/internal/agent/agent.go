// Package agent manages support-agent accounts: registration, login and the
// account-scoped JWT used by the support package's dashboard endpoints.
package agent

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"qrchat/internal/apperr"
)

const (
	MaxNameLen     = 80
	MinPasswordLen = 8
	MaxPasswordLen = 72 // bcrypt ignores bytes beyond 72
)

type Agent struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// CleanName trims and validates a display name.
func CleanName(s string) (string, error) {
	s = strings.Join(strings.Fields(s), " ")
	n := utf8.RuneCountInString(s)
	if n == 0 {
		return "", apperr.Invalid("name is required")
	}
	if n > MaxNameLen {
		return "", apperr.Invalid("name must be at most 80 characters")
	}
	return s, nil
}

// CleanEmail trims and lower-cases an email address and checks its shape loosely;
// the only hard requirement is exactly one '@' with something on both sides.
func CleanEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 || strings.ContainsAny(s, " \t\r\n") || strings.Count(s, "@") != 1 {
		return "", apperr.Invalid("email is invalid")
	}
	if len(s) > 254 {
		return "", apperr.Invalid("email is too long")
	}
	return s, nil
}

// ValidatePassword enforces a minimum/maximum length; bcrypt truncates beyond 72 bytes.
func ValidatePassword(pw string) error {
	if len(pw) < MinPasswordLen {
		return apperr.Invalid("password must be at least 8 characters")
	}
	if len(pw) > MaxPasswordLen {
		return apperr.Invalid("password must be at most 72 characters")
	}
	return nil
}
