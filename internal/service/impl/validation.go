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

const minImageDimension = 1     // minImageDimension is the smallest allowed width or height of an image.
const maxImageDimension = 12000 // maxImageDimension is the largest allowed width or height of an image

// validate runs all business validation rules for an uploaded image:
// request actions and file content/dimensions.
func validate(image *models.Image) error {
	if err := validateRequest(image.Request); err != nil {
		return err
	}
	return validateFile(image.File)
}

// validateRequest checks that at least one supported action is provided
// and validates action-specific parameters (watermark text, resize dimensions).
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

// validateAction ensures the action string is non-empty and is one of
// the supported actions (thumbnail, resize, watermark).
func validateAction(action string) error {

	if strings.TrimSpace(action) == "" {
		return errs.ErrNoActionsProvided
	}

	if _, ok := allowedActions[action]; !ok {
		return errs.ErrUnsupportedAction
	}

	return nil

}

// validateWatermark requires non-empty text when the watermark action is used.
func validateWatermark(watermark string) error {
	if strings.TrimSpace(watermark) == "" {
		return errs.ErrWatermarkTextRequired
	}
	return nil
}

// validateResize ensures that at least one dimension is positive and
// neither is negative.
func validateResize(width int, height int) error {

	if width == 0 && height == 0 {
		return errs.ErrResizeDimensionsRequired
	}

	if width < 0 || height < 0 {
		return errs.ErrNegativeResizeDimensions
	}

	return nil

}

// validateFile decodes the image header to verify supported format
// and valid dimensions (1..12000 px).
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

// validateDimensions checks that image width and height are at least
// minImageDimension and do not exceed maxImageDimension.
func validateDimensions(config image.Config) error {

	if config.Width < minImageDimension || config.Height < minImageDimension {
		return errs.ErrInvalidImageDimensions
	}

	if config.Width > maxImageDimension || config.Height > maxImageDimension {
		return errs.ErrImageTooLargeDimensions
	}

	return nil

}
