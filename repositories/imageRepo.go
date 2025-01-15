package repositories

import (
	"database/sql"
	"fmt"

	"github.com/svuvi/theweek/models"
)

type ImageRepo struct {
	db *sql.DB
}

func NewImageRepo(db *sql.DB) *ImageRepo {
	return &ImageRepo{
		db: db,
	}
}

func (r *ImageRepo) Create(filename string, uploadedBy int, content []byte) (int, error) {
	res, err := r.db.Exec("INSET INTO images(filename, uploaded_by, content) VALUES (?, ?, ?)", filename, uploadedBy, content)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("похоже, что эта база данных не поддерживает функцию LastInsertId:\n%s", err.Error())
	}
	return int(id), nil
}

func (r *ImageRepo) Get(id int) (*models.Image, error) {
	var i models.Image

	row := r.db.QueryRow("SELECT * FROM images WHERE id=?", id)
	err := row.Scan(&i.ID, &i.Filename, &i.UploadedBy, &i.UploadedAt, &i.Content)

	return &i, err
}

func (r *ImageRepo) ChangeFilename(id int, newFilename string) error {
	res, err := r.db.Exec("UPDATE images SET file_name=$1 WHERE id=$2", newFilename, id)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *ImageRepo) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM images WHERE id=$1", id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}
