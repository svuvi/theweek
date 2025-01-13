package repositories

import (
	"database/sql"
	"fmt"

	"github.com/svuvi/theweek/models"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Create(username, hashedPassowrd string) (*models.User, error) {
	res, err := r.db.Exec("INSERT INTO users(username, hashed_password) VALUES (?, ?)", username, hashedPassowrd)
	if err != nil {
		return &models.User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return &models.User{}, fmt.Errorf("похоже, что эта база данных не поддерживает функцию LastInsertId:\n%s", err.Error())
	}
	return r.GetByID(int(id))
}

func (r *UserRepo) GetByID(id int) (*models.User, error) {
	var user models.User

	row := r.db.QueryRow("SELECT * FROM users WHERE id=?", id)
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassowrd, &user.RegisteredAt, &user.IsAdmin)
	if err != nil {
		return &models.User{}, err
	}

	return &user, err
}

func (r *UserRepo) GetByUsername(username string) (*models.User, error) {
	var user models.User

	row := r.db.QueryRow("SELECT * FROM users WHERE username=?", username)
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassowrd, &user.RegisteredAt, &user.IsAdmin)
	if err != nil {
		return &models.User{}, err
	}

	return &user, err
}

func (r *UserRepo) ChangeUsername(id int, newUsername string) error {
	res, err := r.db.Exec("UPDATE users SET username=$1 WHERE id=$2", newUsername, id)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *UserRepo) ChangePassword(id int, newHashedPassword string) error {
	res, err := r.db.Exec("UPDATE users SET hashed_password=$1 WHERE id=$2", newHashedPassword, id)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *UserRepo) SetAdmin(id int, isAdmin bool) error {
	res, err := r.db.Exec("UPDATE users SET is_admin=$1 WHERE id=$2", isAdmin, id)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *UserRepo) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}
