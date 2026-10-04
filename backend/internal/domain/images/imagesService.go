package images

import (
	"crypto/sha256"
	"encoding/hex"
	"example/internal/apiErrors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	*Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo}
}

func Sha256(img multipart.File) (string, error) {
	hasher := sha256.New()
	img.Seek(0, io.SeekStart)
	if _, err := io.Copy(hasher, img); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}
	img.Seek(0, io.SeekStart)
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}

func (this *Service) IsImage(img multipart.File) bool {
	img.Seek(0, io.SeekStart)
	mn := make([]byte, 512)
	io.ReadAtLeast(img, mn, 512)
	img.Seek(0, io.SeekStart)

	mimeType := http.DetectContentType(mn)
	return strings.HasPrefix(mimeType, "image")
}

func (this *Service) Save(img multipart.File) (string, error) {
	if !this.IsImage(img) {
		return "", apiErrors.New("Solo se admiten imágenes", http.StatusBadRequest)
	}

	hash, err := Sha256(img)
	if err != nil {
		return "", err
	}

	exists, err := this.Repo.exists(hash)
	if err != nil {
		return "", err
	}
	if !exists {
		this.Repo.saveToDisk(hash, img)
	}
	return hash, nil
}
