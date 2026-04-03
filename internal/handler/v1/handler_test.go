package v1

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"Proteus/internal/config"
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"Proteus/internal/service/mocks"

	"github.com/stretchr/testify/require"
	"github.com/wb-go/wbf/ginext"
	"go.uber.org/mock/gomock"
)

const testUUID = "123e4567-e89b-12d3-a456-426614174000"
const invalidUUID = "lqhfjlqhwfljkhqwklf"

func createUploadRequest(t *testing.T, fileContent []byte, filename, contentType string, fields map[string]string) *http.Request {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}

	header := make(textproto.MIMEHeader)

	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	require.NoError(t, err)

	_, err = part.Write(fileContent)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request

}

func TestHandler_UploadImage(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := mocks.NewMockService(controller)

	header := NewHandler(config.Server{MaxFileSize: 2048, MaxRequestSize: 5 * 1024 * 1024}, mockService)

	router := ginext.New("")
	router.POST("/upload", header.UploadImage)

	t.Run("invalid form binding (ShouldBind error)", func(t *testing.T) {
		fields := map[string]string{"height": "not-a-number"}
		request := createUploadRequest(t, []byte("data"), "test.png", "image/png", fields)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("no image file (FormFile error)", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		require.NoError(t, writer.WriteField("action", "resize"))
		require.NoError(t, writer.Close())
		request := httptest.NewRequest(http.MethodPost, "/upload", body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrNoFile.Error())
	})

	t.Run("validateHeader: empty file (Size == 0)", func(t *testing.T) {
		request := createUploadRequest(t, []byte{}, "empty.png", "image/png", map[string]string{"action": "resize"})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrNoFile.Error())
	})

	t.Run("validateHeader: file too large (Size > MaxFileSize)", func(t *testing.T) {
		request := createUploadRequest(t, make([]byte, 3000), "big.png", "image/png", map[string]string{"action": "resize"})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrFileTooLarge.Error())
	})

	t.Run("validateHeader: unsupported format", func(t *testing.T) {
		request := createUploadRequest(t, []byte("data"), "bad.txt", "text/plain", map[string]string{"action": "resize"})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrUnsupportedImageFormat.Error())
	})

	t.Run("file too large after reading (LimitReader + len check)", func(t *testing.T) {
		request := createUploadRequest(t, make([]byte, 3000), "big.png", "image/png", map[string]string{"action": "resize"})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrFileTooLarge.Error())
	})

	t.Run("service error", func(t *testing.T) {
		fields := map[string]string{"action": "resize", "height": "800", "width": "600"}
		request := createUploadRequest(t, []byte("fake data"), "test.png", "image/png", fields)
		mockService.EXPECT().UploadImage(gomock.Any(), gomock.Any()).Return("", errors.New("internal error"))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("success upload", func(t *testing.T) {
		fields := map[string]string{"action": "resize", "height": "800", "width": "600"}
		request := createUploadRequest(t, []byte("fake data"), "test.png", "image/png", fields)
		mockService.EXPECT().UploadImage(gomock.Any(), gomock.Any()).Return("id123", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Contains(t, response.Body.String(), "id123")
	})

}

func TestHandler_GetImage(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := mocks.NewMockService(controller)
	header := NewHandler(config.Server{}, mockService)

	router := ginext.New("")
	router.GET("/image/:id", header.GetImage)

	t.Run("invalid UUID", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/image/"+invalidUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrInvalidImageID.Error())
	})

	t.Run("service GetImageMeta error", func(t *testing.T) {
		mockService.EXPECT().GetImageMeta(gomock.Any(), testUUID).Return("", "", errors.New("internal error"))
		request := httptest.NewRequest(http.MethodGet, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("image pending", func(t *testing.T) {
		mockService.EXPECT().GetImageMeta(gomock.Any(), testUUID).Return("objKey", models.StatusPending, nil)
		request := httptest.NewRequest(http.MethodGet, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusAccepted, response.Code)
	})

	t.Run("DownloadImage error", func(t *testing.T) {
		mockService.EXPECT().GetImageMeta(gomock.Any(), testUUID).Return("objKey", models.StatusReady, nil)
		mockService.EXPECT().DownloadImage(gomock.Any(), "objKey").Return(nil, "", errors.New("download failed"))
		request := httptest.NewRequest(http.MethodGet, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("success download", func(t *testing.T) {
		mockService.EXPECT().GetImageMeta(gomock.Any(), testUUID).Return("objKey", models.StatusReady, nil)
		mockService.EXPECT().DownloadImage(gomock.Any(), "objKey").Return([]byte("image-data"), "image/png", nil)
		request := httptest.NewRequest(http.MethodGet, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "image/png", response.Header().Get("Content-Type"))
		require.Equal(t, "image-data", response.Body.String())
	})

	t.Run("image not found", func(t *testing.T) {
		mockService.EXPECT().GetImageMeta(gomock.Any(), testUUID).Return("", "", errs.ErrImageNotFound)
		request := httptest.NewRequest(http.MethodGet, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrImageNotFound.Error())
	})

}

func TestHandler_MarkAsDeleted(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockService := mocks.NewMockService(controller)
	header := NewHandler(config.Server{}, mockService)

	router := ginext.New("")
	router.DELETE("/image/:id", header.MarkAsDeleted)

	t.Run("invalid UUID", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodDelete, "/image/"+invalidUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrInvalidImageID.Error())
	})

	t.Run("service error", func(t *testing.T) {
		mockService.EXPECT().MarkAsDeleted(gomock.Any(), testUUID).Return(errors.New("internal error"))
		request := httptest.NewRequest(http.MethodDelete, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockService.EXPECT().MarkAsDeleted(gomock.Any(), testUUID).Return(nil)
		request := httptest.NewRequest(http.MethodDelete, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Contains(t, response.Body.String(), "deleted")
	})

	t.Run("image not found", func(t *testing.T) {
		mockService.EXPECT().MarkAsDeleted(gomock.Any(), testUUID).Return(errs.ErrImageNotFound)
		request := httptest.NewRequest(http.MethodDelete, "/image/"+testUUID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusNotFound, response.Code)
		require.Contains(t, response.Body.String(), errs.ErrImageNotFound.Error())
	})

}
