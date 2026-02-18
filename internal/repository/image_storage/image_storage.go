package image_storage

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage/minio"
	"context"
	"time"

	m "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageStorage interface {
	Close() error
	UploadImage(ctx context.Context, image *models.Image) error
	GetPresignedURL(ctx context.Context, objectKey string, expirySeconds time.Duration) (url string, err error)
	Delete(ctx context.Context, objectKey string) error
}

func NewImageStorage(logger logger.Logger, config config.ImageStorage, imageDb *m.Client) ImageStorage {
	return minio.NewImageStorage(logger, config, imageDb)
}

func ConnectDB(config config.ImageStorage) (*m.Client, error) {

	return m.New(
		config.MinIOEndpoint,
		&m.Options{Creds: credentials.NewStaticV4(
			config.MinIOAccessKey,
			config.MinIOSecretKey, ""),
			Secure: config.MinIOUseSSL,
			Region: config.MinIORegion})

}
