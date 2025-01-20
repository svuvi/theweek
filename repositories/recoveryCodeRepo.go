package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/models"
)

type RecoveryCodeRepo struct {
	db *sql.DB
}

func NewRecoveryCodeRepo(db *sql.DB) *RecoveryCodeRepo {
	return &RecoveryCodeRepo{
		db: db,
	}
}

func (r *RecoveryCodeRepo) Create(userID int) (*models.RecoveryCode, error) {
	code := uuid.NewString()
	_, err := r.db.Exec("INSERT INTO recovery_codes(user_id, recovery_code) VALUES (?, ?)", userID, code)
	if err != nil {
		return &models.RecoveryCode{}, err
	}

	return r.Get(code)
}

func (r *RecoveryCodeRepo) GetAll() ([]*models.RecoveryCode, error) {
	rows, err := r.db.Query("SELECT * FROM recovery_codes")
	if err != nil {
		return []*models.RecoveryCode{}, err
	}
	defer rows.Close()

	var codes []*models.RecoveryCode
	for rows.Next() {
		c := new(models.RecoveryCode)
		if err := rows.Scan(&c.ID, &c.UserID, &c.RecoveryCode, &c.CreatedAt, &c.UsedAt); err != nil {
			return codes, err
		}
		codes = append(codes, c)
	}
	if err := rows.Err(); err != nil {
		return codes, err
	}
	return codes, nil
}

func (r *RecoveryCodeRepo) Get(code string) (*models.RecoveryCode, error) {
	var rCode models.RecoveryCode

	row := r.db.QueryRow("SELECT * FROM recovery_codes WHERE recovery_code=?", code)
	err := row.Scan(&rCode.ID, &rCode.UserID, &rCode.RecoveryCode, &rCode.CreatedAt, &rCode.UsedAt)
	if err != nil {
		return &models.RecoveryCode{ID: 0}, err
	}

	return &rCode, err
}

func (r *RecoveryCodeRepo) SetUsed(code string) error {
	rCode, err := r.Get(code)
	if err != nil {
		return err
	}

	if rCode.CreatedAt != rCode.UsedAt {
		return errors.New("код уже использован")
	}

	res, err := r.db.Exec("UPDATE recovery_codes SET used_at=? WHERE id=?", time.Now(), rCode.ID)
	if err != nil {
		return err
	}

	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *RecoveryCodeRepo) Delete(rCodeID int) error {
	res, err := r.db.Exec("DELETE FROM recovery_codes WHERE id=$1", rCodeID)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}
