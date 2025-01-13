package repositories

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/models"
)

type InviteRepo struct {
	db *sql.DB
}

func NewInviteRepo(db *sql.DB) *InviteRepo {
	return &InviteRepo{
		db: db,
	}
}

func (r *InviteRepo) Create() (*models.Invite, error) {
	res, err := r.db.Exec("INSERT INTO invites(code) VALUES (?)", uuid.NewString())
	if err != nil {
		return &models.Invite{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return &models.Invite{}, fmt.Errorf("this database probably doesn't suppore LastInsertId function:\n%s", err.Error())
	}
	return r.GetByID(int(id))
}

func (r *InviteRepo) GetByID(id int) (*models.Invite, error) {
	var i models.Invite

	row := r.db.QueryRow("SELECT * FROM invites WHERE id=?", id)
	err := row.Scan(&i.ID, &i.Code, &i.CreatedAt, &i.ClaimedAt, &i.IsActive, &i.ClaimedByUserID)

	return &i, err
}

func (r *InviteRepo) GetByCode(code string) (*models.Invite, error) {
	var i models.Invite

	row := r.db.QueryRow("SELECT * FROM invites WHERE code=?", code)
	err := row.Scan(&i.ID, &i.Code, &i.CreatedAt, &i.ClaimedAt, &i.IsActive, &i.ClaimedByUserID)

	return &i, err
}

func (r *InviteRepo) GetAll() ([]*models.Invite, error) {
	rows, err := r.db.Query("SELECT * FROM invites")
	if err != nil {
		return []*models.Invite{}, err
	}
	defer rows.Close()

	var invites []*models.Invite
	for rows.Next() {
		i := new(models.Invite)
		if err := rows.Scan(&i.ID, &i.Code, &i.CreatedAt, &i.ClaimedAt, &i.IsActive, &i.ClaimedByUserID); err != nil {
			return invites, err
		}
		invites = append(invites, i)
	}
	if err = rows.Err(); err != nil {
		return invites, err
	}
	return invites, nil
}

func (r *InviteRepo) Claim(code string, userID int) error {
	res, err := r.db.Exec("UPDATE invites SET is_active=0, claimed_by_user_id=? WHERE code=?", userID, code)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("unexpected amount of rows affected: %d", affected)
	}
	return nil
}

func (r *InviteRepo) Delete(code string) error {
	res, err := r.db.Exec("DELETE FROM invites WHERE code=?", code)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("unexpected amount of rows affected: %d", affected)
	}
	return nil
}
