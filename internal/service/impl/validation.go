package impl

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

func validate(image *models.Image) error {
	if err := validateRequest(image.Request); err != nil {
		return err
	}
	return validateFile(image.File)
}

func validateRequest(request models.Request) error {

	if err := validateAction(request.Action); err != nil {
		return err
	}

	switch request.Action {
	case models.Thumbnail:
		return nil
	case models.Watermark:
		return validateWatermark(request.Watermark)
	case models.Resize:
		return validateResize(request.Width, request.Height)
	default:
		return errs.ErrUnsupportedAction
	}

}

var allowedActions = map[string]struct{}{models.Thumbnail: {}, models.Resize: {}, models.Watermark: {}}

func validateAction(action string) error {

	if strings.TrimSpace(action) == "" {
		return errs.ErrNoActionsProvided
	}

	if _, ok := allowedActions[action]; !ok {
		return errs.ErrUnsupportedAction
	}

	return nil

}

func validateWatermark(watermark string) error {
	if strings.TrimSpace(watermark) == "" {
		return errs.ErrWatermarkTextRequired
	}
	return nil
}

func validateResize(width int, height int) error {

	if width == 0 && height == 0 {
		return errs.ErrResizeDimensionsRequired
	}

	if width < 0 || height < 0 {
		return errs.ErrNegativeResizeDimensions
	}

	return nil

}

func validateFile(file []byte) error {

	reader := bytes.NewReader(file)

	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return errs.ErrInvalidImageContent
	}

	switch format {
	case "jpeg", "png", "gif", "webp":
	default:
		return errs.ErrUnsupportedImageFormat
	}

	return validateDimensions(config)

}

func validateDimensions(config image.Config) error {

	if config.Width < 1 || config.Height < 1 {
		return errs.ErrInvalidImageDimensions
	}

	if config.Width > 12000 || config.Height > 12000 {
		return errs.ErrImageTooLargeDimensions
	}

	return nil

}
