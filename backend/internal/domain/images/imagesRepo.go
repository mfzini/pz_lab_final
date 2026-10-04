package images

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Repo struct {
	Path string
}

func NewRepo(path string) *Repo {
	return &Repo{path}
}

func (this *Repo) exists(hash string) (bool, error) {
	file_path := filepath.Join(this.Path, hash)
	_, err := os.Stat(file_path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return true, err
}

func (this *Repo) saveToDisk(hash string, img multipart.File) error {
	dstPath := filepath.Join(this.Path, hash)
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()
	(img).Seek(0, io.SeekStart)
	if _, err := io.Copy(file, img); err != nil {
		return err
	}
	return nil
}
