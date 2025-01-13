package models

import "time"

type Invite struct {
	ID              int
	Code            string
	CreatedAt       time.Time
	ClaimedAt       time.Time
	IsActive        bool
	ClaimedByUserID int
}

type InviteRepository interface {
	Create() (*Invite, error)
	GetByID(id int) (*Invite, error)
	GetByCode(code string) (*Invite, error)
	GetAll() ([]*Invite, error)
	// Claim обновляет ClaimedAt и ClaimedByUserID поля и делает приглашение неактивным
	// Возвращает ошибку если приглашение уже использовано, или есть проблема с БД
	Claim(code string, userID int) error
	Delete(code string) error
}