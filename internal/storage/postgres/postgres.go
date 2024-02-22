package postgres

import (
	// "context"
	// "database/sql"
	// "fmt"
	// "time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	// "github.com/orenvadi/auth-grpc/internal/domain/models"
	// "github.com/orenvadi/auth-grpc/internal/storage"
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
