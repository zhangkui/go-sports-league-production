package validator

import (
	"regexp"
	"strings"

	"github.com/goxm2/sports-league/internal/pkg/errorsx"
)

var (
	usernameRE = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,32}$`)
	emailRE    = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
	codeRE     = regexp.MustCompile(`^[A-Za-z0-9_-]{2,32}$`)
	timeRE     = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d(:[0-5]\d)?$`)
)

// Password validates minimum password strength.
func Password(p string) error {
	if len(p) < 8 {
		return errorsx.BadRequest("password must be at least 8 characters")
	}
	if len(p) > 72 {
		return errorsx.BadRequest("password too long (max 72)")
	}
	hasLetter, hasDigit := false, false
	for _, r := range p {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errorsx.BadRequest("password must contain letters and digits")
	}
	return nil
}

func Username(u string) error {
	if !usernameRE.MatchString(u) {
		return errorsx.BadRequest("username must be 3-32 chars: letters, digits, _ . -")
	}
	return nil
}

func Email(e string) error {
	if !emailRE.MatchString(e) {
		return errorsx.BadRequest("invalid email format")
	}
	return nil
}

func CodeField(name, v string) error {
	if !codeRE.MatchString(v) {
		return errorsx.BadRequest(name + " must be 2-32 chars: letters, digits, _ -")
	}
	return nil
}

func TimeStr(v string) error {
	if v == "" {
		return nil
	}
	if !timeRE.MatchString(v) {
		return errorsx.BadRequest("time must be HH:MM or HH:MM:SS")
	}
	return nil
}

// Require returns a bad-request error if s is empty.
func Require(fields map[string]string) error {
	for name, v := range fields {
		if strings.TrimSpace(v) == "" {
			return errorsx.BadRequest(name + " is required")
		}
	}
	return nil
}

// OneOf validates v against an allowed set (lowercased).
func OneOf(name, v string, allowed ...string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return errorsx.BadRequest(name + " must be one of: " + strings.Join(allowed, ", "))
}
