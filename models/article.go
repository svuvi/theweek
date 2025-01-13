package models

import "time"

type Article struct {
	ID        int
	Slug      string
	CreatedAt time.Time
	Title     string
	TextMD    string
}

type ArticleRepository interface {
	Create(slug, title, textMD string) error
	GetByID(id int) (*Article, error)
	GetBySlug(slug string) (*Article, error)
	GetAll() ([]*Article, error)
}
