package models

import "time"

type User struct {
	ID               int64     `db:"id"`
	FirstName        string    `db:"first_name"`
	LastName         string    `db:"last_name"`
	PhoneNumber      string    `db:"phone_number"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Email            string    `db:"email"`
	PasswordHash     []byte    `db:"pass_hash"`
	IsAdmin          bool      `db:"is_admin"`
	IsEmailConfirmed bool      `db:"is_email_confirmed"`
}
