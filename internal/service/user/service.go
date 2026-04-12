// Package usersvc implements the user.Service interface.
// It depends ONLY on the user.Repository interface — never on a concrete DB type.
package usersvc

import (
	"context"
	"fmt"
	"net/mail"

	"golang.org/x/crypto/bcrypt"

	"templio.local/cms/internal/domain/user"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	bcryptCost      = 12
)

// Service implements user.Service.
type Service struct {
	repo user.Repository
}

// New creates a Service with the given repository dependency.
func New(repo user.Repository) *Service {
	return &Service{repo: repo}
}

// ── user.Service implementation ───────────────────────────────────────────────

func (s *Service) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *Service) List(ctx context.Context, search string, page, pageSize int) ([]*user.User, int64, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, search, offset, pageSize)
}

func (s *Service) Create(ctx context.Context, in user.CreateUserInput) (*user.User, error) {
	// ── Validation ───────────────────────────────────────────────────────────
	if err := validateEmail(in.Email); err != nil {
		return nil, err
	}
	if len(in.Password) < 8 {
		return nil, user.ErrWeakPassword
	}
	if err := validateRole(in.Role); err != nil {
		return nil, err
	}

	// ── Uniqueness check ─────────────────────────────────────────────────────
	if _, err := s.repo.FindByEmail(ctx, in.Email); err == nil {
		return nil, user.ErrEmailTaken
	}

	// ── Hash password ────────────────────────────────────────────────────────
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("usersvc: hash password: %w", err)
	}

	u := &user.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: string(hash),
		Role:     in.Role,
		Active:   true,
	}
	u.Sanitize()

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("usersvc: create: %w", err)
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id int64, in user.UpdateUserInput) (*user.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Email != nil {
		if err := validateEmail(*in.Email); err != nil {
			return nil, err
		}
		u.Email = *in.Email
	}
	if in.Role != nil {
		if err := validateRole(*in.Role); err != nil {
			return nil, err
		}
		u.Role = *in.Role
	}
	if in.Active != nil {
		u.Active = *in.Active
	}

	u.Sanitize()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("usersvc: update: %w", err)
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return user.ErrInvalidEmail
	}
	return nil
}

func validateRole(r user.Role) error {
	switch r {
	case user.RoleAdmin, user.RoleEditor, user.RoleViewer:
		return nil
	default:
		return user.ErrInvalidRole
	}
}
