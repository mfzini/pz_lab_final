package images

import (
	"example/internal/domain/common"
	"mime/multipart"
)

type Model struct {
	Id           string
	Sha256       string `json:"-"`
	OriginalName string `db:"original_name"`
	MimeType     string
	Multipart    *multipart.File `db:"-" json:"-"`
	common.TimeStamps
}
