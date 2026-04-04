package minio_test

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/image_storage"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	wbf "github.com/wb-go/wbf/config"
)

var testStorage image_storage.ImageStorage
var testClient *minio.Client
var testBucket string

func TestMain(m *testing.M) {

	if err := wbf.New().LoadEnvFiles("../../../../.env"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := config.ImageStorage{
		MinIOEndpoint:  "minio-test:9000",
		MinIOAccessKey: os.Getenv("MINIO_ROOT_USER"),
		MinIOSecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
		MinIOUseSSL:    false,
		MinIORegion:    "us-east-1",
		MinIOBucket:    "proteus-test-bucket",
	}

	logger, _ := logger.NewLogger(config.Logger{Debug: true})

	var err error
	testClient, err = image_storage.ConnectDB(cfg)
	if err != nil {
		logger.LogFatal("image_storage_test — failed to connect to test MinIO", err, "layer", "repository.image_storage_test")
	}

	testBucket = cfg.MinIOBucket
	testStorage = image_storage.NewImageStorage(logger, cfg, testClient)

	exitCode := m.Run()
	testStorage.Close()
	os.Exit(exitCode)

}

func resetBucket(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	exists, err := testClient.BucketExists(ctx, testBucket)
	if err != nil {
		t.Fatalf("BucketExists failed: %v", err)
	}

	if exists {
		err = testClient.RemoveBucketWithOptions(ctx, testBucket, minio.RemoveBucketOptions{ForceDelete: true})
		if err != nil {
			t.Fatalf("RemoveBucketWithOptions failed: %v", err)
		}
	}

	err = testClient.MakeBucket(ctx, testBucket, minio.MakeBucketOptions{})
	if err != nil {
		t.Fatalf("MakeBucket failed: %v", err)
	}

}

func TestUploadImage(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	image := &models.Image{
		ObjectKey:   "test-upload.jpg",
		File:        []byte("fake image content for testing"),
		Size:        28,
		ContentType: "image/jpeg",
	}

	err := testStorage.UploadImage(ctx, image)
	if err != nil {
		t.Fatalf("UploadImage failed: %v", err)
	}

	_, err = testClient.StatObject(ctx, testBucket, image.ObjectKey, minio.StatObjectOptions{})
	if err != nil {
		t.Fatalf("uploaded object not found: %v", err)
	}

}

func TestUploadImage_Error(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	image := &models.Image{ObjectKey: "", File: []byte("test"), Size: 4, ContentType: "image/jpeg"}

	err := testStorage.UploadImage(ctx, image)
	if err == nil {
		t.Fatalf("expected error for empty ObjectKey, got nil")
	}

}

func TestDownloadImage(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	objectKey := "test-download.jpg"
	originalData := []byte("aboba")

	_ = testStorage.UploadImage(ctx, &models.Image{
		ObjectKey:   objectKey,
		File:        originalData,
		Size:        int64(len(originalData)),
		ContentType: "image/jpeg",
	})

	data, err := testStorage.DownloadImage(ctx, objectKey)
	if err != nil {
		t.Fatalf("DownloadImage failed: %v", err)
	}
	if string(data) != string(originalData) {
		t.Fatalf("downloaded data mismatch")
	}

}

func TestDownloadImage_NonExistent(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	_, err := testStorage.DownloadImage(ctx, "non-existent-object.jpg")
	if err == nil {
		t.Fatalf("expected error for non-existent object, got nil")
	}

}

func TestDeleteImage(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	objectKey := "test-delete.jpg"
	_ = testStorage.UploadImage(ctx, &models.Image{
		ObjectKey:   objectKey,
		File:        []byte("to be deleted"),
		Size:        14,
		ContentType: "image/jpeg",
	})

	err := testStorage.DeleteImage(ctx, objectKey)
	if err != nil {
		t.Fatalf("DeleteImage failed: %v", err)
	}

}

func TestDeleteImage_NonExistent(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	err := testStorage.DeleteImage(ctx, "non-existent-delete.jpg")
	if err != nil {
		t.Fatalf("DeleteImage on non-existent returned unexpected error: %v", err)
	}

}

func TestDeleteBatch(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	keys := []string{"batch1.jpg", "batch2.jpg", "batch3.jpg"}

	for _, k := range keys {
		_ = testStorage.UploadImage(ctx, &models.Image{
			ObjectKey:   k,
			File:        []byte("batch test"),
			Size:        10,
			ContentType: "image/jpeg",
		})
	}

	err := testStorage.DeleteBatch(ctx, keys)
	if err != nil {
		t.Fatalf("DeleteBatch failed: %v", err)
	}

}

func TestDeleteBatch_Empty(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	err := testStorage.DeleteBatch(ctx, []string{})
	if err != nil {
		t.Fatalf("DeleteBatch with empty slice failed: %v", err)
	}

}

func TestDeleteBatch_Error(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	keys := []string{"ok1.jpg", "ok2.jpg"}

	for _, k := range keys {
		_ = testStorage.UploadImage(ctx, &models.Image{
			ObjectKey:   k,
			File:        []byte("test"),
			Size:        4,
			ContentType: "image/jpeg",
		})
	}

	badKeys := append(keys, "very/long/key/with/invalid/chars/%!@%.jpg", "non-existent.jpg")
	err := testStorage.DeleteBatch(ctx, badKeys)
	if err == nil {
		t.Logf("DeleteBatch returned no error")
	} else {
		t.Logf("DeleteBatch returned expected error: %v", err)
	}

}

func TestUploadDownloadDeleteRoundtrip(t *testing.T) {

	resetBucket(t)
	ctx := context.Background()

	objectKey := "roundtrip.jpg"
	data := []byte("roundtrip test data")

	image := &models.Image{
		ObjectKey:   objectKey,
		File:        data,
		Size:        int64(len(data)),
		ContentType: "image/jpeg",
	}

	if err := testStorage.UploadImage(ctx, image); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	downloaded, err := testStorage.DownloadImage(ctx, objectKey)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if string(downloaded) != string(data) {
		t.Fatalf("data mismatch after roundtrip")
	}

	if err := testStorage.DeleteImage(ctx, objectKey); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

}

func TestClose(t *testing.T) {

	log, _ := logger.NewLogger(config.Logger{Debug: true})
	client, _ := image_storage.ConnectDB(config.ImageStorage{
		MinIOEndpoint:  "minio-test:9000",
		MinIOAccessKey: os.Getenv("MINIO_ROOT_USER"),
		MinIOSecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
		MinIOUseSSL:    false,
		MinIORegion:    "us-east-1",
		MinIOBucket:    "proteus-test-bucket",
	})

	st := image_storage.NewImageStorage(log, config.ImageStorage{}, client)
	st.Close()

}
