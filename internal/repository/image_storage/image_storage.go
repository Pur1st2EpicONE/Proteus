// Package image_storage provides a high-level abstraction for
// object storage operations (upload, download, delete) backed by MinIO.
// It defines the ImageStorage interface and wires the concrete MinIO
// implementation.
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

// ImageStorage defines the contract for persisting and retrieving
// raw and processed images in object storage (MinIO).
type ImageStorage interface {
	UploadImage(ctx context.Context, image *models.Image) error          // UploadImage stores the original image bytes under the computed object key.
	DownloadImage(ctx context.Context, objectKey string) ([]byte, error) // DownloadImage retrieves the full image bytes by its MinIO object key.
	DeleteImage(ctx context.Context, objectKey string) error             // DeleteImage removes a single object from the bucket. It is used by the service layer for rollback/cleanup when upload or task enqueueing fails.
	DeleteBatch(ctx context.Context, objectKeys []string) error          // DeleteBatch removes multiple objects (including their "unprocessed" variants) in one operation.
	Close()                                                              // Close releases idle connections and any other resources held by the underlying client.
}

// NewImageStorage returns a concrete MinIO-based implementation of
// the ImageStorage interface.
func NewImageStorage(logger logger.Logger, config config.ImageStorage, imageDb *m.Client) ImageStorage {
	return minio.NewImageStorage(logger, config, imageDb)
}

// ConnectDB creates and returns a configured MinIO client using
// static credentials and the settings from the config.
func ConnectDB(config config.ImageStorage) (*m.Client, error) {

	return m.New(config.MinIOEndpoint,
		&m.Options{Creds: credentials.NewStaticV4(
			config.MinIOAccessKey,
			config.MinIOSecretKey, ""),
			Secure: config.MinIOUseSSL,
			Region: config.MinIORegion})

}
