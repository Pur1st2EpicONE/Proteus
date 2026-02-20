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

var allowed = map[string]struct{}{models.Thumbnail: {}, models.Resize: {}, models.Watermark: {}}

func validateRequest(request models.Request) error {

	if len(request.Action) == 0 {
		return errs.ErrNoActionsProvided
	}

	actionSet := make(map[string]struct{})

	for _, action := range request.Action {
		if _, ok := allowed[action]; !ok {
			return errs.ErrUnsupportedAction
		}
		actionSet[action] = struct{}{}
	}

	if _, ok := actionSet[models.Watermark]; ok {
		if strings.TrimSpace(request.Watermark) == "" {
			return errs.ErrWatermarkTextRequired
		}
	}

	if _, ok := actionSet[models.Resize]; ok {
		if request.Width <= 0 && request.Height <= 0 {
			return errs.ErrResizeDimensionsRequired
		}
		if request.Width < 0 || request.Height < 0 {
			return errs.ErrNegativeResizeDimensions
		}
	}

	if request.Quality != 0 {
		if request.Quality < 1 || request.Quality > 100 {
			return errs.ErrInvalidQualityRange
		}
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
