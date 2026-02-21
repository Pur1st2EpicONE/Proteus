package impl

import (
	"Proteus/internal/models"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"time"

	"github.com/disintegration/imaging"
)

func (s *Service) ProcessImage(ctx context.Context, task models.ImageProcessTask) error {

	srcImg, format, err := s.getImage(ctx, task.ObjectKey)
	if err != nil {
		return fmt.Errorf("failed to get source image: %w", err)
	}

	var file []byte

	switch task.Action {

	case models.Thumbnail:
		img := imaging.Thumbnail(srcImg, 150, 150, imaging.Lanczos)
		buf, _ := encode(img, format)
		file = buf

	case models.Watermark:
		img := addWatermark(srcImg)
		buf, _ := encode(img, format)
		file = buf
	}

	time.Sleep(10 * time.Second)

	err = s.imageStorage.UploadProcessed(ctx, task.ObjectKey[2:], file, task.MimeType)
	if err != nil {
		return err
	}

	if err := s.metaStorage.MarkAsReady(ctx, task.ID, task.ObjectKey[2:]); err != nil {
		return fmt.Errorf("failed to mark image as ready: %w", err)
	}

	fmt.Println("processing completed:", task.ID)

	return nil

}

func (s *Service) getImage(ctx context.Context, objectKey string) (image.Image, string, error) {

	imageBytes, err := s.imageStorage.DownloadImage(ctx, objectKey)
	if err != nil {
		return nil, "", err
	}

	srcImg, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", err
	}

	return srcImg, format, nil

}

func encode(img image.Image, format string) ([]byte, error) {

	var buf bytes.Buffer

	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
		return buf.Bytes(), err
	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	case "gif":
		err := gif.Encode(&buf, img, nil)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

}

func addWatermark(src image.Image) image.Image {
	return imaging.Overlay(src, imaging.New(200, 50, color.NRGBA{0, 0, 0, 100}),
		image.Pt(src.Bounds().Dx()-210, src.Bounds().Dy()-60), 0.6)
}
