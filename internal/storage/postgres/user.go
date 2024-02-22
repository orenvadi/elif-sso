package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/storage"
)

func (s *Storage) SaveUser(ctx context.Context, firstName, lastName, phoneNumber, email string, passwordHash []byte) (uid int64, err error) {
	const op = "storage.postgres.SaveUser"
	var id int64
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users(first_name, last_name, phone_number, created_at, updated_at, email, pass_hash)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, firstName, lastName, phoneNumber, time.Now(), time.Now(), email, passwordHash).Scan(&id)
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
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UserAllData(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT id, first_name, last_name, phone_number, created_at, updated_at, email, pass_hash, is_admin FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w, user = %v", op, storage.ErrUserNotFound, user)
		}
		return models.User{}, fmt.Errorf("%s: %w, user = %v", op, err, user)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	var app models.App
	err := s.db.GetContext(ctx, &app, "SELECT id, name, secret FROM apps WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user models.User) error {
	const op = "storage.postgres.UpdateUser"

	log.Println(user)
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET first_name = $1, last_name = $2, phone_number = $3, email = $4 
		WHERE id = $5
	`, user.FirstName, user.LastName, user.PhoneNumber, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, userID int64) error {
	const op = "storage.postgres.DeleteUser"

	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	var isAdmin bool
	err := s.db.GetContext(ctx, &isAdmin, "SELECT is_admin FROM users WHERE id = $1", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}
