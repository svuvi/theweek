package models

import "time"

type Image struct {
	ID         int
	Filename   string
	UploadedBy int
	UploadedAt time.Time
	Content    []byte
}

type ImageRepository interface {
	Create(filename string, uploadedBy int, content []byte) (int, error) // Returns ID of the uploaded image
	Get(id int) (*Image, error)
	ChangeFilename(id int, newFilename string) error
	Delete(id int) error
}
