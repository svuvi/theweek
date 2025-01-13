package models

import "time"

type User struct {
	ID             int
	Username       string
	HashedPassowrd string
	RegisteredAt   time.Time
	IsAdmin        bool
}

type UserRepository interface {
	Create(username, hashedPassowrd string) (*User, error)
	GetByID(id int) (*User, error)
	GetByUsername(username string) (*User, error)
	ChangeUsername(id int, newUsername string) error
	ChangePassword(id int, newHashedPassword string) error
	SetAdmin(id int, isAdmin bool) error
	Delete(id int) error
}
