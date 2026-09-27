package images

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

type Repo struct {
	db *pgx.Conn
}

func NewRepo(db *pgx.Conn) *Repo {
	return &Repo{db}
}

func (this *Repo) findByHash(ctx context.Context, hash string) *Model {
	query := "select * from images where sha256 = $1"
	r, err := this.db.Query(ctx, query, hash)
	if err != nil {
		return nil
	}
	img, err := pgx.CollectExactlyOneRow(r, pgx.RowToAddrOfStructByNameLax[Model])
	return img
}

func saveToDisk(img *Model) error {
	dstPath := filepath.Join(os.Getenv("PATH_UPLOADS"), fmt.Sprintf("%s.%s", img.Id, filepath.Ext(img.OriginalName)))
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, *img.Multipart); err != nil {
		return err
	}
	return nil
}

func (this *Repo) create(ctx context.Context, tx pgx.Tx, img *Model) error {
	query := "insert into images (id, sha256, mimetype, original_name, created_at) values ($1, $2, $3, $4, $5)"
	_, err := tx.Exec(ctx, query, img.Id, img.Sha256, img.MimeType, img.OriginalName, img.CreatedAt)
	if err != nil {
		return err
	}
	if err := saveToDisk(img); err != nil {
		return err
	}
	return nil
}
