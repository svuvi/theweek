package models

import "time"

type Session struct {
	ID             int
	UserID         int
	SessionKeyHash string
	CreatedAt      time.Time
	LastUse        time.Time
	IsActive       bool
}

type SessionRepository interface {
	Create(userID int, sessionKey string) (*Session, error)
	GetUserSessions(userID int) ([]*Session, error)
	GetSessionByID(sessionID int) (*Session, error)
	GetSessionByKey(sessionKey string) (*Session, error)
	// Updates last_use field in session. Uses current time
	UpdateLastUsedByKey(sessionKey string) error
	UpdateLastUsedByID(sessionID int) error
	SetInactive(sessionKey string) error
}
