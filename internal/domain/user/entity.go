// Package user defines the pure domain entity for the User aggregate.
// It has NO dependency on any framework, ORM, or transport layer.
package user

import "time"

// Role represents a user's authorization level.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// User is the core aggregate root for the user domain.
// Bun struct tags live ONLY in the repository layer — this entity is framework-agnostic.
type User struct {
	ID        int64
	Name      string
	Email     string
	Password  string // bcrypt hash
	Role      Role
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateUserInput holds validated data required to create a new user.
type CreateUserInput struct {
	Name     string
	Email    string
	Password string // plain text — service layer hashes it
	Role     Role
}

// UpdateUserInput holds fields that can be mutated after creation.
// Pointer fields allow partial updates (nil = don't touch).
type UpdateUserInput struct {
	Name   *string
	Email  *string
	Role   *Role
	Active *bool
}

// Sanitize trims leading/trailing whitespace from user-supplied strings.
func (u *User) Sanitize() {
	u.Name = trimSpace(u.Name)
	u.Email = trimSpace(u.Email)
}

func trimSpace(s string) string {
	result := []rune{}
	start, end := 0, len([]rune(s))-1
	runes := []rune(s)
	for start <= end && (runes[start] == ' ' || runes[start] == '\t') {
		start++
	}
	for end >= start && (runes[end] == ' ' || runes[end] == '\t') {
		end--
	}
	return string(result) + string(runes[start:end+1])
}
