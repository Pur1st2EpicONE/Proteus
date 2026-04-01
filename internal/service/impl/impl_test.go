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
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func createFileHeader(filename string, contentType string, content []byte) *multipart.FileHeader {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", filename)
	fw.Write(content)
	w.Close()

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)

	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     int64(len(content)),
	}
}

func createTestJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)
	return buf.Bytes()
}

func TestNewService(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)
	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	require.NotNil(t, svc)
}

func TestService_UploadImage_WithAsyncRollback(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockLogger := mockLogger.NewMockLogger(controller)
	mockBroker := mockBroker.NewMockProducer(controller)
	mockMeta := mockMetaStorage.NewMockMetaStorage(controller)
	mockImage := mockImageStorage.NewMockImageStorage(controller)

	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogInfo(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().LogError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	file := createTestJPEG()
	fh := createFileHeader("test.jpg", "image/jpeg", file)
	img := &models.Image{File: file, FileHeader: fh, Request: models.Request{Action: models.Thumbnail}}

	t.Run("producer error triggers iStorageRollback with async wait", func(t *testing.T) {

		done := make(chan struct{})

		mockImage.EXPECT().DeleteImage(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, _ any) {
			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			require.WithinDuration(t, time.Now().Add(10*time.Second), deadline, time.Second)
			close(done)
		}).Return(nil)

		mockMeta.EXPECT().SaveImageMeta(ctx, gomock.Any()).Return(nil)
		mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
		mockBroker.EXPECT().Send(ctx, gomock.Any(), gomock.Any()).Return(errors.New("producer fail"))

		_, err := svc.UploadImage(ctx, img)

		require.Error(t, err)
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("DeleteImage was not called in time")
		}
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

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
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

	err := svc.ProcessImage(ctx, task)
	require.NoError(t, err)

	task.Action = models.Watermark
	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(imageBytes, nil)
	mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
	mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err = svc.ProcessImage(ctx, task)
	require.NoError(t, err)

	task.Action = models.Resize
	task.Width = 50
	task.Height = 50
	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(imageBytes, nil)
	mockImage.EXPECT().UploadImage(ctx, gomock.Any()).Return(nil)
	mockMeta.EXPECT().MarkAsReady(ctx, gomock.Any(), gomock.Any()).Return(nil)

	err = svc.ProcessImage(ctx, task)
	require.NoError(t, err)

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

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)

	ctx := context.Background()
	task := models.ImageProcessTask{ID: "id1", ObjectKey: "key1.png", Action: models.Thumbnail}

	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("", "", errors.New("db fail"))
	err := svc.ProcessImage(ctx, task)
	require.Error(t, err)

	mockMeta.EXPECT().GetImageMeta(ctx, task.ID).Return("key1.png", "pending", nil)
	mockImage.EXPECT().DownloadImage(ctx, task.ObjectKey).Return(nil, errors.New("download fail"))

	err = svc.ProcessImage(ctx, task)
	require.Error(t, err)

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

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	mockImage.EXPECT().DownloadImage(ctx, "key1.jpg").Return([]byte{1, 2}, nil)
	data, ct, err := svc.DownloadImage(ctx, "key1.jpg")

	require.NoError(t, err)
	require.Equal(t, "image/jpeg", ct)
	require.Len(t, data, 2)

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

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	mockMeta.EXPECT().GetImageMeta(ctx, "id1").Return("key", "ready", nil)
	key, status, err := svc.GetImageMeta(ctx, "id1")

	require.NoError(t, err)
	require.Equal(t, "key", key)
	require.Equal(t, "ready", status)

	mockMeta.EXPECT().GetImageMeta(ctx, "id2").Return("", models.StatusDeleted, nil)
	_, _, err = svc.GetImageMeta(ctx, "id2")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

	mockMeta.EXPECT().GetImageMeta(ctx, "id3").Return("", "", sql.ErrNoRows)
	_, _, err = svc.GetImageMeta(ctx, "id3")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

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

	svc := impl.NewService(mockLogger, config.Service{}, mockBroker, mockMeta, mockImage)
	ctx := context.Background()

	mockMeta.EXPECT().MarkAsDeleted(ctx, "id1").Return(nil)
	err := svc.MarkAsDeleted(ctx, "id1")
	require.NoError(t, err)

	mockMeta.EXPECT().MarkAsDeleted(ctx, "id2").Return(sql.ErrNoRows)
	err = svc.MarkAsDeleted(ctx, "id2")
	require.ErrorIs(t, err, errs.ErrImageNotFound)

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

	cfg := config.Service{Cleaner: true, CleanupInterval: 10 * time.Millisecond}

	svc := impl.NewService(mockLogger, cfg, mockBroker, mockMeta, mockImage)

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

	go svc.Cleaner(ctx)

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("Cleaner did not delete any images in time")
	}

}
