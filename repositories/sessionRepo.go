package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/svuvi/theweek/models"
	"golang.org/x/crypto/sha3"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{
		db: db,
	}
}

func (r *SessionRepo) Create(userID int, sessionKey string) (*models.Session, error) {
	res, err := r.db.Exec("INSERT INTO sessions(user_id, session_key_hash) VALUES (?, ?)", userID, sha3Hash(sessionKey))
	if err != nil {
		return &models.Session{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return &models.Session{}, fmt.Errorf("похоже, что эта база данных не поддерживает функцию LastInsertId:\n%s", err.Error())
	}
	return r.GetSessionByID(int(id))
}

func (r *SessionRepo) GetUserSessions(userID int) ([]*models.Session, error) {
	rows, err := r.db.Query("SELECT * FROM sessions WHERE user_id=?", userID)
	if err != nil {
		return []*models.Session{}, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		s := new(models.Session)
		if err := rows.Scan(&s.ID, &s.UserID, &s.SessionKeyHash, &s.CreatedAt, &s.LastUse, &s.IsActive); err != nil {
			return sessions, err
		}
		sessions = append(sessions, s)
	}
	if err = rows.Err(); err != nil {
		return sessions, err
	}
	return sessions, nil
}

func (r *SessionRepo) GetSessionByID(sessionID int) (*models.Session, error) {
	var s models.Session

	row := r.db.QueryRow("SELECT * FROM sessions WHERE id=?", sessionID)
	err := row.Scan(&s.ID, &s.UserID, &s.SessionKeyHash, &s.CreatedAt, &s.LastUse, &s.IsActive)

	return &s, err
}

func (r *SessionRepo) GetSessionByKey(sessionKey string) (*models.Session, error) {
	var s models.Session

	row := r.db.QueryRow("SELECT * FROM sessions WHERE session_key_hash=?", sha3Hash(sessionKey))
	err := row.Scan(&s.ID, &s.UserID, &s.SessionKeyHash, &s.CreatedAt, &s.LastUse, &s.IsActive)

	return &s, err
}

func (r *SessionRepo) UpdateLastUsedByKey(sessionKey string) error {
	res, err := r.db.Exec("UPDATE sessions SET last_use=$1 WHERE session_key_hash=$2", time.Now(), sha3Hash(sessionKey))

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *SessionRepo) UpdateLastUsedByID(sessionID int) error {
	res, err := r.db.Exec("UPDATE sessions SET last_use=$1 WHERE id=$2", time.Now(), sessionID)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *SessionRepo) SetInactive(sessionKey string) error {
	res, err := r.db.Exec("UPDATE sessions SET is_active=0 WHERE session_key_hash=?", sha3Hash(sessionKey))

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func sha3Hash(input string) string {
	hash := sha3.NewShake256()
	_, _ = hash.Write([]byte(input))
	sha3 := hash.Sum(nil)
	return string(sha3)
}
