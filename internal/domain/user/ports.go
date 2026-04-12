package user

import "context"

// Repository defines the persistence contract for the User aggregate.
// Implementations live in /internal/repository/user — this interface is owned by the domain.
//
//go:generate mockgen -source=ports.go -destination=../../repository/user/mock_repository.go -package=userrepo
type Repository interface {
	// FindByID retrieves a user by primary key. Returns ErrNotFound if absent.
	FindByID(ctx context.Context, id int64) (*User, error)

	// FindByEmail retrieves a user by unique email. Returns ErrNotFound if absent.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// List returns a paginated slice of users with filtering support.
	List(ctx context.Context, search string, offset, limit int) ([]*User, int64, error)

	// Create persists a new user and populates ID/timestamps.
	Create(ctx context.Context, u *User) error

	// Update persists mutations on an existing user.
	Update(ctx context.Context, u *User) error

	// Delete soft-deletes a user by ID.
	Delete(ctx context.Context, id int64) error
}

// Service defines the business-logic contract consumed by transport handlers.
// Implementations live in /internal/service/user.
//
//go:generate mockgen -source=ports.go -destination=../../service/user/mock_service.go -package=usersvc
type Service interface {
	// GetByID returns a user or a domain error.
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByEmail returns a user by email — used by the auth handler.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// List returns a page of users with total count.
	List(ctx context.Context, search string, page, pageSize int) ([]*User, int64, error)

	// Create validates input, hashes the password, and creates the user.
	Create(ctx context.Context, in CreateUserInput) (*User, error)

	// Update applies partial updates to an existing user.
	Update(ctx context.Context, id int64, in UpdateUserInput) (*User, error)

	// Delete removes a user by ID.
	Delete(ctx context.Context, id int64) error
}
