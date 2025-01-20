package models

import "time"

type RecoveryCode struct {
	ID           int
	UserID       int
	RecoveryCode string
	CreatedAt    time.Time
	UsedAt       time.Time
}

type RecoveryCodeRepository interface {
	Create(userID int) (*RecoveryCode, error)
	Get(code string) (*RecoveryCode, error)
	GetAll() ([]*RecoveryCode, error)
	SetUsed(code string) error
	Delete(rCodeID int) error
}

func (rCode *RecoveryCode) IsActive() bool {
	if rCode.CreatedAt.IsZero() || rCode.UsedAt.IsZero() {
		return false
	}
	if rCode.CreatedAt == rCode.UsedAt {
		return true
	}
	return false
}
