package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/storage"
)

func (s *Storage) SaveConfirmationCode(ctx context.Context, userID int64, code string) error {
	const op = "storage.postgres.SaveConfirmationCode"

	// location, _ := time.LoadLocation("Asia/Bishkek")
	// now := time.Now().In(location)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO email_confirmation(user_id, code)
		VALUES($1, $2)
	`, userID, code)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ConfirmationCode(ctx context.Context, userID int64) (confCodeModel models.ConfirmCode, err error) {
	const op = "storage.postgres.ConfirmationCode"

	confCode := models.ConfirmCode{}

	err = s.db.GetContext(ctx, &confCode, `
		SELECT ec.id, code, email, ec.created_at
		FROM email_confirmation ec
		INNER JOIN users u ON ec.user_id = u.id
		WHERE ec.user_id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ConfirmCode{}, fmt.Errorf("%s: %w", op, storage.ErrConfirmCodeNotFound)
		}
		return models.ConfirmCode{}, fmt.Errorf("%s: %w", op, err)
	}

	return confCode, nil
}

func (s *Storage) DeleteConfirmationCode(ctx context.Context, user_id int64) error {
	const op = "storage.postgres.DeleteConfirmationCode"

	_, err := s.db.ExecContext(ctx, "DELETE FROM email_confirmation WHERE user_id = $1", user_id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
