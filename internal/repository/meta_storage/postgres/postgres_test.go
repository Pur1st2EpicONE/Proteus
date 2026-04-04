package postgres_test

import (
	"Proteus/internal/config"
	"Proteus/internal/logger"
	"Proteus/internal/models"
	"Proteus/internal/repository/meta_storage/postgres"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	wbf "github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

var testStorage *postgres.MetaStorage

func TestMain(m *testing.M) {

	if err := wbf.New().LoadEnvFiles("../../../../.env"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := config.MetaStorage{
		Host:     "postgres-test",
		Port:     "5432",
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   "proteus_test",
		SSLMode:  "disable",
		QueryRetryStrategy: config.RetryStrategy{
			Attempts: 3,
			Delay:    100 * time.Millisecond,
			Backoff:  1.5,
		},
		PendingTimeout: 24 * time.Hour,
	}

	logger, _ := logger.NewLogger(config.Logger{Debug: true})

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode), nil, &dbpg.Options{})
	if err != nil {
		logger.LogFatal("meta_storage_test — failed to connect to test DB", err, "layer", "repository.meta_storage_test")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		logger.LogFatal("meta_storage_test — failed to set goose dialect", err, "layer", "repository.meta_storage_test")
	}

	if err := goose.Up(db.Master, "../../../../migrations"); err != nil {
		logger.LogFatal("meta_storage_test — failed to run migrations", err, "layer", "repository.meta_storage_test")
	}

	testStorage = postgres.NewMetaStorage(logger, cfg, db)

	exitCode := m.Run()
	testStorage.Close()
	os.Exit(exitCode)

}

func setupTest(t *testing.T) {

	ctx := context.Background()
	_, err := testStorage.DB().ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 100 * time.Millisecond, Backoff: 1.5}, `
	
	TRUNCATE TABLE images 
	RESTART IDENTITY`)

	if err != nil {
		t.Fatalf("failed to truncate images: %v", err)
	}

}

func TestSaveImageMeta_Errors(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	image := &models.Image{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		ObjectKey: "images/test.jpg",
		Status:    models.StatusPending,
	}

	err := testStorage.SaveImageMeta(ctx, image)
	if err != nil {
		t.Fatalf("SaveImageMeta failed: %v", err)
	}

	err = testStorage.SaveImageMeta(ctx, image)
	if err == nil {
		t.Fatalf("expected error for duplicate UUID, got nil")
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
		t.Logf("expected unique violation captured: %v", err)
	} else {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestGetImageMeta(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	uuid := "550e8400-e29b-41d4-a716-446655440001"
	image := &models.Image{
		ID:        uuid,
		ObjectKey: "images/test.jpg",
		Status:    models.StatusReady,
	}

	err := testStorage.SaveImageMeta(ctx, image)
	if err != nil {
		t.Fatalf("SaveImageMeta failed: %v", err)
	}

	key, status, err := testStorage.GetImageMeta(ctx, uuid)
	if err != nil {
		t.Fatalf("GetImageMeta failed: %v", err)
	}

	if key != "images/test.jpg" || status != models.StatusReady {
		t.Fatalf("unexpected data: key=%s status=%s", key, status)
	}

	_, _, err = testStorage.GetImageMeta(ctx, "non-existent-uuid")
	if err == nil {
		t.Fatalf("expected error for non-existent UUID, got nil")
	}

}

func TestMarkAsReady(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	uuid := "550e8400-e29b-41d4-a716-446655440002"
	image := &models.Image{ID: uuid, ObjectKey: "", Status: models.StatusPending}

	err := testStorage.SaveImageMeta(ctx, image)
	if err != nil {
		t.Fatalf("SaveImageMeta failed: %v", err)
	}

	err = testStorage.MarkAsReady(ctx, "final/images/test.jpg", uuid)
	if err != nil {
		t.Fatalf("MarkAsReady failed: %v", err)
	}

	key, status, err := testStorage.GetImageMeta(ctx, uuid)
	if err != nil {
		t.Fatalf("GetImageMeta after MarkAsReady failed: %v", err)
	}
	if key != "final/images/test.jpg" || status != models.StatusReady {
		t.Fatalf("MarkAsReady did not update correctly")
	}

	err = testStorage.MarkAsReady(ctx, "dummy.jpg", "non-existent")
	if err == nil {
		t.Fatalf("expected error for non-existent image, got nil")
	}

}

func TestMarkAsDeleted(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	uuid := "550e8400-e29b-41d4-a716-446655440003"
	image := &models.Image{ID: uuid, Status: models.StatusReady}

	err := testStorage.SaveImageMeta(ctx, image)
	if err != nil {
		t.Fatalf("SaveImageMeta failed: %v", err)
	}

	err = testStorage.MarkAsDeleted(ctx, uuid)
	if err != nil {
		t.Fatalf("MarkAsDeleted failed: %v", err)
	}

	_, status, err := testStorage.GetImageMeta(ctx, uuid)
	if err != nil {
		t.Fatalf("GetImageMeta after MarkAsDeleted failed: %v", err)
	}
	if status != models.StatusDeleted {
		t.Fatalf("expected status deleted, got %s", status)
	}

	err = testStorage.MarkAsDeleted(ctx, "non-existent")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

}

func TestGetDeleted(t *testing.T) {

	setupTest(t)

	ctx := context.Background()

	delID := "550e8400-e29b-41d4-a716-446655440004"
	_ = testStorage.SaveImageMeta(ctx, &models.Image{
		ID:        delID,
		ObjectKey: "deleted.jpg",
		Status:    models.StatusDeleted,
	})

	oldID := "550e8400-e29b-41d4-a716-446655440005"
	_ = testStorage.SaveImageMeta(ctx, &models.Image{
		ID:        oldID,
		ObjectKey: "old.jpg",
		Status:    models.StatusPending,
	})

	_, _ = testStorage.DB().ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 100 * time.Millisecond, Backoff: 1.5}, `
		UPDATE images 
		SET updated_at = $1 
		WHERE uuid = $2`,
		time.Now().UTC().Add(-48*time.Hour), oldID)

	images, err := testStorage.GetDeleted(ctx)
	if err != nil {
		t.Fatalf("GetDeleted failed: %v", err)
	}

	if len(images) != 2 {
		t.Fatalf("expected 2 images (deleted + old pending), got %d", len(images))
	}

	foundDel := false
	foundOld := false

	for _, img := range images {
		if img.ID == delID {
			foundDel = true
		}
		if img.ID == oldID {
			foundOld = true
		}
	}

	if !foundDel || !foundOld {
		t.Fatalf("missing expected images in GetDeleted result: %+v", images)
	}

}

func TestDeleteBatch(t *testing.T) {

	setupTest(t)

	ctx := context.Background()

	ids := []string{
		"550e8400-e29b-41d4-a716-446655440006",
		"550e8400-e29b-41d4-a716-446655440007",
		"550e8400-e29b-41d4-a716-446655440008",
	}

	for _, id := range ids {
		_ = testStorage.SaveImageMeta(ctx, &models.Image{
			ID:        id,
			ObjectKey: "batch/" + id + ".jpg",
			Status:    models.StatusReady,
		})
	}

	err := testStorage.DeleteBatch(ctx, ids[:2])
	if err != nil {
		t.Fatalf("DeleteBatch failed: %v", err)
	}

	var count int
	err = testStorage.DB().Master.QueryRowContext(ctx, `

	SELECT COUNT(*) 
	FROM images 
	WHERE uuid = ANY($1)`,

		pq.Array(ids)).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 remaining image, got %d", count)
	}

	err = testStorage.DeleteBatch(ctx, []string{})
	if err != nil {
		t.Fatalf("DeleteBatch with empty slice failed: %v", err)
	}

}

func TestClose(t *testing.T) {
	log, _ := logger.NewLogger(config.Logger{Debug: true})
	db, _ := dbpg.New(fmt.Sprintf("host=postgres-test port=5432 user=%s password=%s dbname=proteus_test sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")), nil, &dbpg.Options{})
	st := postgres.NewMetaStorage(log, config.MetaStorage{}, db)
	st.Close()
}
