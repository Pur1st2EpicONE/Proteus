package impl

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"bytes"
	"image"
)

func validateImage(img *models.Image) error {

	reader := bytes.NewReader(img.File)

	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return errs.ErrInvalidImageContent
	}

	allowedFormats := map[string]bool{"jpeg": true, "png": true, "webp": true, "gif": true}

	if !allowedFormats[format] {
		return errs.ErrUnsupportedImageFormat
	}

	if err := validateDimensions(config); err != nil {
		return err
	}

	return nil

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
