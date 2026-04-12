package user

import "errors"

// Sentinel errors for the user domain.
// Callers use errors.Is/As — never string comparison.
var (
	ErrNotFound       = errors.New("user: not found")
	ErrEmailTaken     = errors.New("user: email already registered")
	ErrInvalidEmail   = errors.New("user: invalid email address")
	ErrWeakPassword   = errors.New("user: password must be at least 8 characters")
	ErrInvalidRole    = errors.New("user: invalid role")
	ErrUnauthorized   = errors.New("user: unauthorized")
	ErrAlreadyDeleted = errors.New("user: already deleted")
)
