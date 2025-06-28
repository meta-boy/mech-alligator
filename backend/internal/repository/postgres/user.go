package postgres

import (
	"context"

	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/user"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRowContext(ctx, query, username)

	var u user.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, u *user.User) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query, u.Username, u.PasswordHash).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}
