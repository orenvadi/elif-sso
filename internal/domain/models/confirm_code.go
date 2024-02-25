package models

import "time"

type ConfirmCode struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Code      string    `db:"code"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}
