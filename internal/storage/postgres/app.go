package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/orenvadi/auth-grpc/internal/domain/models"
	"github.com/orenvadi/auth-grpc/internal/storage"
)

func (s *Storage) App(ctx context.Context, id int64) (models.App, error) {
	const op = "storage.postgres.App"

	var app models.App
	err := s.db.GetContext(ctx, &app, "SELECT * FROM apps WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

// func (s *Storage) AppSecret(ctx context.Context, appID int) (string, error) {
// 	const op = "storage.postgres.AppSecret"

// 	var secret string
// 	err := s.db.GetContext(ctx, &secret, "SELECT secret FROM apps WHERE id = $1", appID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return "", fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
// 		}
// 		return "", fmt.Errorf("%s: %w", op, err)
// 	}

// 	return secret, nil
// }
