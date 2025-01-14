package repositories

import (
	"database/sql"

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

func (r *ArticleRepo) Create(slug, title, textMD, description string) error {
	_, err := r.db.Exec("INSERT INTO articles(slug, title, textMD, description) VALUES (?, ?, ?, ?)", slug, title, textMD, description)
	return err
}

func (r *ArticleRepo) GetByID(id int) (*models.Article, error) {
	var a models.Article

	row := r.db.QueryRow("SELECT * FROM articles WHERE id=?", id)
	err := row.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description)

	return &a, err
}

func (r *ArticleRepo) GetBySlug(slug string) (*models.Article, error) {
	var a models.Article

	row := r.db.QueryRow("SELECT * FROM articles WHERE slug=?", slug)
	err := row.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description)

	return &a, err
}

func (r *ArticleRepo) GetAll() ([]*models.Article, error) {
	rows, err := r.db.Query("SELECT * FROM articles")
	if err != nil {
		return []*models.Article{}, err
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		a := new(models.Article)
		if err := rows.Scan(&a.ID, &a.Slug, &a.CreatedAt, &a.Title, &a.TextMD, &a.Description); err != nil {
			return articles, err
		}
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return articles, err
	}
	return articles, nil
}
