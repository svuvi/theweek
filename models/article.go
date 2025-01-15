package models

import (
	"time"
)

type Article struct {
	ID           int
	Slug         string
	CreatedAt    time.Time
	Title        string
	TextMD       string
	Description  string
	CoverImageID int
}

type ArticleRepository interface {
	Create(slug, title, textMD, description string, coverImageID int) error // coverImageID = 0 если отсутствует
	SetCoverImage(id int, newCoverImageID int) error                                // coverImageID = 0 если отсутствует
	GetByID(id int) (*Article, error)
	GetBySlug(slug string) (*Article, error)
	GetAll() ([]*Article, error)
	Delete(id int) error
}
