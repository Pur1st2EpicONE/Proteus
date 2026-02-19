package impl

import (
	"Proteus/internal/errs"
	"Proteus/internal/models"
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

func validateImage(img *models.Image) error {

	reader := bytes.NewReader(img.File)

	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		fmt.Println(err)
		return errs.ErrInvalidImageContent
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
