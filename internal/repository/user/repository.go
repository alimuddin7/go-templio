// Package userrepo provides the Bun-backed implementation of user.Repository.
// It translates domain entities to/from DB models with Bun struct tags.
package userrepo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/alimuddin7/go-templio/internal/domain/user"
)

// userModel is the Bun ORM model. It is PRIVATE — the repository translates
// to/from the pure domain entity so the domain stays ORM-free.
type userModel struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,unique,notnull"`
	Password  string    `bun:"password,notnull"`
	Role      string    `bun:"role,notnull,default:'viewer'"`
	Active    bool      `bun:"active,notnull,default:true"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

// Repository is the Bun implementation of user.Repository.
type Repository struct {
	db *bun.DB
}

// New returns a new Repository. Inject *bun.DB from the engine package.
func New(db *bun.DB) *Repository {
	return &Repository{db: db}
}

// ── Repository interface implementation ──────────────────────────────────────

func (r *Repository) FindByID(ctx context.Context, id int64) (*user.User, error) {
	m := new(userModel)
	err := r.db.NewSelect().Model(m).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return toEntity(m), nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	m := new(userModel)
	err := r.db.NewSelect().Model(m).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}
	return toEntity(m), nil
}

func (r *Repository) List(ctx context.Context, search string, offset, limit int) ([]*user.User, int64, error) {
	var models []userModel
	q := r.db.NewSelect().Model(&models)

	if search != "" {
		q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("name ILIKE ?", "%"+search+"%").
				WhereOr("email ILIKE ?", "%"+search+"%")
		})
	}

	total, err := q.OrderExpr("id ASC").
		Offset(offset).
		Limit(limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]*user.User, len(models))
	for i, m := range models {
		m := m // pin
		out[i] = toEntity(&m)
	}
	return out, int64(total), nil
}

func (r *Repository) Create(ctx context.Context, u *user.User) error {
	m := fromEntity(u)
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().Model(m).Returning("id").Exec(ctx)
	if err != nil {
		return err
	}
	u.ID = m.ID
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *Repository) Update(ctx context.Context, u *user.User) error {
	m := fromEntity(u)
	m.UpdatedAt = time.Now()

	_, err := r.db.NewUpdate().
		Model(m).
		Column("name", "email", "role", "active", "updated_at").
		Where("id = ?", m.ID).
		Exec(ctx)
	return err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		TableExpr("users").
		Set("active = false").
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// ── entity ↔ model translation ────────────────────────────────────────────────

func toEntity(m *userModel) *user.User {
	return &user.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		Role:      user.Role(m.Role),
		Active:    m.Active,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromEntity(u *user.User) *userModel {
	return &userModel{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		Role:      string(u.Role),
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
