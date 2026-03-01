package impl

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

const wmWidth = 150
const wmHeight = 50

func (s *Service) ProcessImage(ctx context.Context, task models.ImageProcessTask) error {

	_, _, err := s.metaStorage.GetImageMeta(ctx, task.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to get image status from meta storage: %w", err)
	}

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
		img := addWatermark(srcImg, task.Watermark)
		buf, _ := encode(img, format)
		file = buf

	case models.Resize:
		img := imaging.Resize(srcImg, task.Width, task.Height, imaging.Lanczos)
		buf, _ := encode(img, format)
		file = buf

	default:
		return errs.ErrUnsupportedAction
	}

	err = s.imageStorage.UploadImage(ctx, &models.Image{
		ObjectKey:   task.ObjectKey[2:],
		File:        file,
		Size:        int64(len(file)),
		ContentType: task.ContentType})

	if err != nil {
		return fmt.Errorf("failed to upload image to image storage: %w", err)
	}

	if err := s.metaStorage.MarkAsReady(ctx, task.ObjectKey[2:], task.ID); err != nil {
		return fmt.Errorf("failed to mark image as ready: %w", err)
	}

	return nil

}

func (s *Service) getImage(ctx context.Context, objectKey string) (image.Image, string, error) {

	imageBytes, err := s.imageStorage.DownloadImage(ctx, objectKey)
	if err != nil {
		return nil, "", fmt.Errorf("error downloading image: %w", err)
	}

	srcImg, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", fmt.Errorf("error decoding image: %w", err)
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

func addWatermark(src image.Image, watermark string) image.Image {

	res := imaging.Clone(src)
	bounds := res.Bounds()

	wm := image.NewRGBA(image.Rect(0, 0, wmWidth, wmHeight))

	for y := range wmHeight {
		for x := range wmWidth {
			wm.Set(x, y, color.NRGBA{0, 0, 0, 100})
		}
	}

	colour := color.White
	face := inconsolata.Bold8x16

	d := &font.Drawer{Dst: wm, Src: image.NewUniform(colour), Face: face}

	textWidth := d.MeasureString(watermark).Round()
	textHeight := face.Metrics().Height.Round()
	descent := face.Metrics().Descent.Round()

	x := (wmWidth - textWidth) / 2
	y := (wmHeight+textHeight)/2 - descent

	d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d.DrawString(watermark)

	overlayPoint := image.Pt(bounds.Dx()-wmWidth-10, bounds.Dy()-wmHeight-10)
	draw.Draw(res, wm.Bounds().Add(overlayPoint), wm, image.Point{}, draw.Over)

	return res

}
