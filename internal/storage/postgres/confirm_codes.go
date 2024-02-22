package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/orenvadi/auth-grpc/internal/storage"
)

func (s *Storage) SaveConfirmationCode(ctx context.Context, userID int64, code string) error {
	const op = "storage.postgres.SaveConfirmationCode"

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO email_confirmation(user_id, code, created_at)
		VALUES($1, $2, $3)
	`, userID, code, time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ConfirmationCode(ctx context.Context, userID int64) (confirmCode string, email string, err error) {
	const op = "storage.postgres.GetConfirmationCodeByEmail"

	err = s.db.QueryRowContext(ctx, `
		SELECT code, email
		FROM email_confirmation ec
		INNER JOIN users u ON ec.user_id = u.id
		WHERE ec.user_id = $1
	`, userID).Scan(&confirmCode, &email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("%s: %w", op, storage.ErrConfirmCodeNotFound)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return confirmCode, email, nil
}

func (s *Storage) DeleteConfirmationCode(ctx context.Context, user_id int64) error {
	const op = "storage.postgres.DeleteConfirmationCode"

	_, err := s.db.ExecContext(ctx, "DELETE FROM email_confirmation WHERE user_id = $1", user_id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
