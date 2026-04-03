package impl_test

import (
	mockBroker "Proteus/internal/broker/mocks"
	"Proteus/internal/config"
	"Proteus/internal/errs"
	mockLogger "Proteus/internal/logger/mocks"
	"Proteus/internal/models"
	mockImageStorage "Proteus/internal/repository/image_storage/mocks"
	mockMetaStorage "Proteus/internal/repository/meta_storage/mocks"
	"Proteus/internal/service/impl"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/image/bmp"
)

func createFileHeader(filename string, contentType string, content []byte) *multipart.FileHeader {

	var buffer bytes.Buffer
	mWriter := multipart.NewWriter(&buffer)
	writer, _ := mWriter.CreateFormFile("file", filename)
	_, _ = writer.Write(content)
	_ = mWriter.Close()

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)

	return &multipart.FileHeader{Filename: filename, Header: header, Size: int64(len(content))}

}

func createTestJPEG() []byte {
	image := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, image, nil)
	return buf.Bytes()
}

func createTestGIF() []byte {
	image := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = gif.Encode(&buf, image, nil)
	return buf.Bytes()
}

func createTestBMP() []byte {
	image := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = bmp.Encode(&buf, image)
	return buf.Bytes()
}

func createInvalidImage() []byte {
	return []byte("the best image format — .ABOBUS")
}

func createHugeImage() []byte {
	image := image.NewRGBA(image.Rect(0, 0, 13000, 100))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, image, nil)
	return buf.Bytes()
}

func createZeroDimensionImage() []byte {
	image := image.NewRGBA(image.Rect(0, 0, 0, 0))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, image, nil)
	return buf.Bytes()
}

func TestService_UploadImage(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	validFile := createTestJPEG()
	validFH := createFileHeader("test.jpg", "image/jpeg", validFile)

	t.Run("validateFile unsupported format default", func(t *testing.T) {
		bmpFile := createTestBMP()
		bmpFH := createFileHeader("test.bmp", "image/bmp", bmpFile)
		image := &models.Image{File: bmpFile, FileHeader: bmpFH, Request: models.Request{Action: models.Thumbnail}}
		_, err := service.UploadImage(ctx, image)
		require.ErrorIs(t, err, errs.ErrUnsupportedImageFormat)
	})

	t.Run("validate fails early return", func(t *testing.T) {
		testCases := []struct {
			name        string
			image       *models.Image
			expectedErr error
		}{
			{"no action provided", &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: ""}}, errs.ErrNoActionsProvided},
			{"unsupported action", &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: "unknown"}}, errs.ErrUnsupportedAction},
			{"watermark missing text", &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Watermark}}, errs.ErrWatermarkTextRequired},
			{"resize missing dimensions", &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Resize}}, errs.ErrResizeDimensionsRequired},
			{"resize negative dimensions", &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Resize, Width: -10}}, errs.ErrNegativeResizeDimensions},
			{"invalid image content (decode fail)", &models.Image{File: createInvalidImage(), FileHeader: createFileHeader("bad.bin", "image/jpeg", createInvalidImage()), Request: models.Request{Action: models.Thumbnail}}, errs.ErrInvalidImageContent},
			{"zero dimensions", &models.Image{File: createZeroDimensionImage(), FileHeader: createFileHeader("zero.jpg", "image/jpeg", createZeroDimensionImage()), Request: models.Request{Action: models.Thumbnail}}, errs.ErrInvalidImageDimensions},
			{"image too large dimensions", &models.Image{File: createHugeImage(), FileHeader: createFileHeader("huge.jpg", "image/jpeg", createHugeImage()), Request: models.Request{Action: models.Thumbnail}}, errs.ErrImageTooLargeDimensions},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := service.UploadImage(ctx, tc.image)
				require.ErrorIs(t, err, tc.expectedErr)
			})
		}
	})

	t.Run("metaStorage.SaveImageMeta fails, async rollback, iStorageRollback called", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		image := &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Thumbnail}}
		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(errors.New("meta save failed"))
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, _ any) { wg.Done() }).Return(nil)
		_, err := service.UploadImage(ctx, image)
		require.Error(t, err)
		wg.Wait()
	})

	t.Run("imageStorage.UploadImage fails, async rollback, only logging", func(t *testing.T) {
		image := &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Thumbnail}}
		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(errors.New("image upload failed"))
		_, err := service.UploadImage(ctx, image)
		require.Error(t, err)
	})

	t.Run("producer.Send fails, iStorageRollback", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		image := &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Thumbnail}}
		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Send(ctx, gomock.Any(), gomock.Any()).Return(errors.New("producer fail"))
		mockImage.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, _ any) { wg.Done() }).Return(nil)
		_, err := service.UploadImage(ctx, image)
		require.Error(t, err)
		wg.Wait()
	})

	t.Run("iStorageRollback DeleteImage fails", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		image := &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Thumbnail}}
		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Send(ctx, gomock.Any(), gomock.Any()).Return(errors.New("producer fail"))
		mockImage.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, _ any) { wg.Done() }).Return(errors.New("delete failed"))
		_, err := service.UploadImage(ctx, image)
		require.Error(t, err)
		wg.Wait()
	})

	t.Run("full success path returns ID", func(t *testing.T) {
		image := &models.Image{File: validFile, FileHeader: validFH, Request: models.Request{Action: models.Thumbnail}}
		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Send(ctx, gomock.Any(), gomock.Any()).Return(nil)
		id, err := service.UploadImage(ctx, image)
		require.NoError(t, err)
		require.NotEmpty(t, id)
		require.Equal(t, image.ID, id)
	})

}

func TestService_ProcessImage_Actions(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	image := image.NewRGBA(image.Rect(0, 0, 300, 300))
	var buf bytes.Buffer
	_ = png.Encode(&buf, image)
	imageBytes := buf.Bytes()

	task := models.ImageProcessTask{
		ID:          "id1",
		ObjectKey:   "key1.png",
		Action:      models.Thumbnail,
		ContentType: "image/png",
		Width:       100,
		Height:      100,
		Watermark:   "WM",
	}

	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(imageBytes, nil)
	mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
	mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err := service.ProcessImage(ctx, task)
	require.NoError(t, err)

	task.Action = models.Watermark
	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(imageBytes, nil)
	mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
	mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err = service.ProcessImage(ctx, task)
	require.NoError(t, err)

	task.Action = models.Resize
	task.Width = 50
	task.Height = 50
	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(imageBytes, nil)
	mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
	mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err = service.ProcessImage(ctx, task)
	require.NoError(t, err)

	t.Run("encode gif", func(t *testing.T) {
		task := models.ImageProcessTask{
			ID:          "id_gif",
			ObjectKey:   "key.gif",
			Action:      models.Watermark,
			ContentType: "image/gif",
			Watermark:   "WM",
		}
		gifBytes := createTestGIF()
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.gif", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(gifBytes, nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)
		err := service.ProcessImage(ctx, task)
		require.NoError(t, err)
	})

}

func TestService_ProcessImage_Errors(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	validImageBytes := createTestJPEG()

	t.Run("generic meta error", func(t *testing.T) {
		task := models.ImageProcessTask{ID: "id1", ObjectKey: "key1.png", Action: models.Thumbnail}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("", "", errors.New("db fail"))
		err := service.ProcessImage(ctx, task)
		require.Error(t, err)
	})

	t.Run("GetImageMeta sql.ErrNoRows", func(t *testing.T) {
		task := models.ImageProcessTask{ID: "id_no_rows", ObjectKey: "key.png", Action: models.Thumbnail}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("", "", sql.ErrNoRows)
		err := service.ProcessImage(ctx, task)
		require.NoError(t, err)
	})

	t.Run("unsupported action default case to ErrUnsupportedAction", func(t *testing.T) {
		task := models.ImageProcessTask{ID: "id_unsupported", ObjectKey: "key.png", Action: "unknown_action"}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.png", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(validImageBytes, nil)
		err := service.ProcessImage(ctx, task)
		require.ErrorIs(t, err, errs.ErrUnsupportedAction)
	})

	t.Run("download error", func(t *testing.T) {
		task := models.ImageProcessTask{ID: "id_download", ObjectKey: "key.png", Action: models.Thumbnail}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.png", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(nil, errors.New("download fail"))
		err := service.ProcessImage(ctx, task)
		require.Error(t, err)
	})

	t.Run("image.Decode error in getImage", func(t *testing.T) {
		task := models.ImageProcessTask{ID: "id_decode", ObjectKey: "key.png", Action: models.Thumbnail}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.png", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(createInvalidImage(), nil)
		err := service.ProcessImage(ctx, task)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error decoding image")
	})

	t.Run("UploadImage error after processing", func(t *testing.T) {
		task := models.ImageProcessTask{
			ID:          "id_upload_fail",
			ObjectKey:   "key.png",
			Action:      models.Thumbnail,
			ContentType: "image/png",
		}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.png", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(validImageBytes, nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(errors.New("upload failed"))
		err := service.ProcessImage(ctx, task)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to upload image to image storage")
	})

	t.Run("MarkAsReady error after upload", func(t *testing.T) {
		task := models.ImageProcessTask{
			ID:          "id_mark_fail",
			ObjectKey:   "key.png",
			Action:      models.Thumbnail,
			ContentType: "image/png",
		}
		mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key.png", "pending", nil)
		mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(validImageBytes, nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(errors.New("mark ready failed"))
		err := service.ProcessImage(ctx, task)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to mark image as ready")
	})

}

func TestService_DownloadImage(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	t.Run("jpg", func(t *testing.T) {
		mockImage.EXPECT().DownloadImage(ctx, "key1.jpg").Return([]byte{1, 2}, nil)
		data, ct, err := service.DownloadImage(ctx, "key1.jpg")
		require.NoError(t, err)
		require.Equal(t, "image/jpeg", ct)
		require.Len(t, data, 2)
	})

	t.Run("png", func(t *testing.T) {
		mockImage.EXPECT().DownloadImage(ctx, "key1.png").Return([]byte{3, 4}, nil)
		_, ct, err := service.DownloadImage(ctx, "key1.png")
		require.NoError(t, err)
		require.Equal(t, "image/png", ct)
	})

	t.Run("gif", func(t *testing.T) {
		mockImage.EXPECT().DownloadImage(ctx, "key1.gif").Return([]byte{5, 6}, nil)
		_, ct, err := service.DownloadImage(ctx, "key1.gif")
		require.NoError(t, err)
		require.Equal(t, "image/gif", ct)
	})

	t.Run("webp", func(t *testing.T) {
		mockImage.EXPECT().DownloadImage(ctx, "key1.webp").Return([]byte{7, 8}, nil)
		_, ct, err := service.DownloadImage(ctx, "key1.webp")
		require.NoError(t, err)
		require.Equal(t, "image/webp", ct)
	})

	t.Run("download error", func(t *testing.T) {
		mockImage.EXPECT().DownloadImage(ctx, "key1.jpg").Return(nil, errors.New("download fail"))
		_, _, err := service.DownloadImage(ctx, "key1.jpg")
		require.Error(t, err)
	})

}

func TestService_GetImageMeta(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	mockMeta.EXPECT().GetImageMeta(ctx, "id1").Return("key", "ready", nil)
	key, status, err := service.GetImageMeta(ctx, "id1")

	require.NoError(t, err)
	require.Equal(t, "key", key)
	require.Equal(t, "ready", status)

	mockMeta.EXPECT().GetImageMeta(ctx, "id2").Return("", models.StatusDeleted, nil)
	_, _, err = service.GetImageMeta(ctx, "id2")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

	mockMeta.EXPECT().GetImageMeta(ctx, "id3").Return("", "", sql.ErrNoRows)
	_, _, err = service.GetImageMeta(ctx, "id3")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

	t.Run("generic meta error", func(t *testing.T) {
		dbErr := errors.New("db connection failed")
		mockLogger.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMeta.EXPECT().GetImageMeta(ctx, "id4").Return("", "", dbErr)
		_, _, err := service.GetImageMeta(ctx, "id4")
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.NotErrorIs(t, err, errs.ErrImageNotFound)
	})

}

func TestService_MarkAsDeleted(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	service := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	mockMeta.EXPECT().MarkAsDeleted(ctx, "id1").Return(nil)
	err := service.MarkAsDeleted(ctx, "id1")
	require.NoError(t, err)

	mockMeta.EXPECT().MarkAsDeleted(ctx, "id2").Return(sql.ErrNoRows)
	err = service.MarkAsDeleted(ctx, "id2")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

	t.Run("generic error", func(t *testing.T) {
		dbErr := errors.New("unexpected mark error")
		mockLogger.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockMeta.EXPECT().MarkAsDeleted(ctx, "id3").Return(dbErr)
		err := service.MarkAsDeleted(ctx, "id3")
		require.Error(t, err)
		require.ErrorIs(t, err, dbErr)
		require.NotErrorIs(t, err, errs.ErrImageNotFound)
	})

}
func TestService_Cleaner(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	config := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}

	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})

	mockMeta.EXPECT().GetDeleted(gomock.Any()).DoAndReturn(func(ctx context.Context) ([]models.Image, error) {
		return []models.Image{{ID: "1", ObjectKey: "o1"}}, nil
	}).AnyTimes()

	mockImage.EXPECT().DeleteBatch(gomock.Any(), []string{"o1"}).Do(func(ctx context.Context, keys []string) {
		close(done)
	}).Return(nil).AnyTimes()

	mockMeta.EXPECT().DeleteBatch(gomock.Any(), []string{"1"}).Return(nil).AnyTimes()

	go service.Cleaner(ctx)

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("Cleaner did not delete any images in time")
	}

}

func TestService_Cleaner_Disabled(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	config := config.Service{Cleaner: false, CleanupInterval: 10 * time.Millisecond}
	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	service.Cleaner(ctx)

}

func TestService_Cleaner_CleanGetDeletedError(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError("service — cleanup failed", gomock.Any(), "layer", "service.impl").MinTimes(1)

	config := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}

	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	mockMeta.EXPECT().GetDeleted(gomock.Any()).Return(nil, errors.New("get deleted fail")).AnyTimes()

	go service.Cleaner(ctx)
	time.Sleep(50 * time.Millisecond)

}

func TestService_Cleaner_CleanEmptyDeleted(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	config := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}
	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	mockMeta.EXPECT().GetDeleted(gomock.Any()).Return([]models.Image{}, nil).AnyTimes()

	go service.Cleaner(ctx)
	time.Sleep(50 * time.Millisecond)

}

func TestService_Cleaner_CleanImageDeleteBatchError(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError("service — cleanup failed", gomock.Any(), "layer", "service.impl").MinTimes(1)

	config := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}
	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	mockMeta.EXPECT().GetDeleted(gomock.Any()).Return([]models.Image{{ID: "1", ObjectKey: "o1"}}, nil).AnyTimes()
	mockImage.EXPECT().DeleteBatch(gomock.Any(), []string{"o1"}).Return(errors.New("image delete batch fail")).AnyTimes()

	go service.Cleaner(ctx)
	time.Sleep(50 * time.Millisecond)

}

func TestService_Cleaner_CleanMetaDeleteBatchError(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError("service — cleanup failed", gomock.Any(), "layer", "service.impl").MinTimes(1)

	config := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}

	service := impl.NewService(mockLogger, config, mockBroker, mockMeta, mockImage)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	mockMeta.EXPECT().GetDeleted(gomock.Any()).Return([]models.Image{{ID: "1", ObjectKey: "o1"}}, nil).AnyTimes()
	mockImage.EXPECT().DeleteBatch(gomock.Any(), []string{"o1"}).Return(nil).AnyTimes()
	mockMeta.EXPECT().DeleteBatch(gomock.Any(), []string{"1"}).Return(errors.New("meta delete batch fail")).AnyTimes()

	go service.Cleaner(ctx)
	time.Sleep(50 * time.Millisecond)

}
