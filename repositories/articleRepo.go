package repositories

import (
	"database/sql"
	"fmt"

	"github.com/svuvi/theweek/models"
)

type ArticleRepo struct {
	db *sql.DB
}

func NewArticleRepo(db *sql.DB) *ArticleRepo {
	return &ArticleRepo{
		db: db,
	}
}

func (r *ArticleRepo) Create(slug, title, textMD, description string, coverImageID int) error {
	ciID := IntToNullInt16(coverImageID)
	_, err := r.db.Exec("INSERT INTO articles(slug, title, textMD, description, cover_image_id) VALUES (?, ?, ?, ?, ?)", slug, title, textMD, description, ciID)
	return err
}

func (r *ArticleRepo) SetCoverImage(id int, newCoverImageID int) error {
	i := IntToNullInt16(newCoverImageID)
	res, err := r.db.Exec("UPDATE articles SET cover_image_id=$1 WHERE id=$2", i, id)

	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *ArticleRepo) GetByID(id int) (*models.Article, error) {
	var a models.Article
	var coverImageID sql.NullInt16

	row := r.db.QueryRow("SELECT * FROM articles WHERE id=?", id)
	err := row.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description, &coverImageID)

	a.CoverImageID = NullInt16ToInt(coverImageID)

	return &a, err
}

func (r *ArticleRepo) GetBySlug(slug string) (*models.Article, error) {
	var a models.Article
	var coverImageID sql.NullInt16

	row := r.db.QueryRow("SELECT * FROM articles WHERE slug=?", slug)
	err := row.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description, &coverImageID)

	a.CoverImageID = NullInt16ToInt(coverImageID)

	return &a, err
}

func (r *ArticleRepo) GetAll() ([]*models.Article, error) {
	rows, err := r.db.Query("SELECT * FROM articles")
	if err != nil {
		return []*models.Article{}, err
	}
	defer rows.Close()

	var articles []*models.Article
	var i sql.NullInt16
	for rows.Next() {
		a := new(models.Article)
		if err := rows.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description, &i); err != nil {
			return articles, err
		}
		a.CoverImageID = NullInt16ToInt(i)
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return articles, err
	}
	return articles, nil
}

func (r *ArticleRepo) Update(a *models.Article) error {
	i := IntToNullInt16(a.CoverImageID)
	res, err := r.db.Exec("UPDATE articles SET slug=$1, created_at=$2, title=$3, textMD=$4, description=$5, cover_image_id=$6 WHERE id=$7",
						a.Slug, a.CreatedAt, a.Title, a.TextMD, a.Description, i, a.ID)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

func (r *ArticleRepo) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM articles WHERE id=$1", id)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 && err == nil {
		return fmt.Errorf("изменено непредвиденное количество строк: %d", affected)
	}
	return nil
}

// Преобразует int в sql.NullInt16
// Если значение равно 0, то выход будет Null
func IntToNullInt16(value int) sql.NullInt16 {
	if value == 0 {
		return sql.NullInt16{
			Int16: 0,
			Valid: false,
		}
	}
	return sql.NullInt16{
		Int16: int16(value),
		Valid: true,
	}
}

// Преобразует sql.NullInt16 в int
// Если значение равно Null, то выход будет 0
func NullInt16ToInt(nullInt sql.NullInt16) int {
	if !nullInt.Valid {
		return 0
	}
	return int(nullInt.Int16)
}
