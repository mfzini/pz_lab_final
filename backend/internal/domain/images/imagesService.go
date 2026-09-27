package images

import (
	"context"
	"example/internal/apiErrors"
	"example/internal/utils"
	"io"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	*Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo}
}

func (this *Service) Create(ctx context.Context, tx pgx.Tx, img *Model) *apiErrors.Error {
	if img.CreatedAt == nil {
		timestamp := time.Now()
		img.CreatedAt = &timestamp
	}
	mn := make([]byte, 512)
	io.ReadAtLeast(*img.Multipart, mn, 512)
	img.MimeType = http.DetectContentType(mn)
	if !strings.HasPrefix(img.MimeType, "image") {
		return apiErrors.New("Solo se admiten imágenes.", http.StatusBadRequest)
	}

	hash, err := utils.SHA256(*img.Multipart)
	if err != nil {
		return apiErrors.Generic(err)
	}
	img.Sha256 = hash

	if onDb := this.findByHash(ctx, img.Sha256); onDb != nil {
		img = onDb
		return nil
	}

	img.Id = uuid.NewV7().String()
	if err := this.Repo.create(ctx, tx, img); err != nil {
		return apiErrors.Generic(err)
	}

	return nil
}
