package models

type User struct {
	ID           int64
	FirstName    string
	LastName     string
	PhoneNumber  string
	Email        string
	PasswordHash []byte
}
