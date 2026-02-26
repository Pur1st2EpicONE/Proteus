package image_storage

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage/minio"
	"context"

	m "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ImageStorage interface {
	UploadImage(ctx context.Context, image *models.Image) error
	DownloadImage(ctx context.Context, objectKey string) ([]byte, error)
	DeleteImage(ctx context.Context, objectKey string) error
	DeleteBatch(ctx context.Context, objectKeys []string) error
	Close()
}

func NewImageStorage(logger logger.Logger, config config.ImageStorage, imageDb *m.Client) ImageStorage {
	return minio.NewImageStorage(logger, config, imageDb)
}

func ConnectDB(config config.ImageStorage) (*m.Client, error) {

	return m.New(config.MinIOEndpoint,
		&m.Options{Creds: credentials.NewStaticV4(
			config.MinIOAccessKey,
			config.MinIOSecretKey, ""),
			Secure: config.MinIOUseSSL,
			Region: config.MinIORegion})

}
