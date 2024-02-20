package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/storage"
)

type Storage struct {
	db *sqlx.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, firstName, lastName, phoneNumber, email string, passwordHash []byte) (uid int64, err error) {
	const op = "storage.postgres.SaveUser"

	var id int64
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users(first_name, last_name, phone_number, email, pass_hash)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id
	`, firstName, lastName, phoneNumber, email, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT id, email, pass_hash FROM users WHERE email = $1", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	var app models.App
	err := s.db.GetContext(ctx, &app, "SELECT id, name, secret FROM apps WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	var isAdmin bool
	err := s.db.GetContext(ctx, &isAdmin, "SELECT is_admin FROM users WHERE id = $1", userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}
