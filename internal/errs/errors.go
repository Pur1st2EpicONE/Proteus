package errs

import "errors"

var (
	ErrInvalidJSON              = errors.New("invalid JSON format")                                           // invalid JSON format
	ErrInternal                 = errors.New("internal server error")                                         // internal server error
	ErrNoFile                   = errors.New("no file provided or invalid form field")                        // no file provided or invalid form field
	ErrFileTooLarge             = errors.New("file too large (max 10MB)")                                     // file too large (max 10MB)
	ErrReadFile                 = errors.New("failed to read file")                                           // failed to read file
	ErrRequestBodyTooLarge      = errors.New("request body too large")                                        // request body too large
	ErrInvalidImageContent      = errors.New("content is corrupted or image format not supported")            // content is corrupted or image format not supported
	ErrUnsupportedImageFormat   = errors.New("unsupported image format (only jpeg, png, webp, gif allowed)")  // unsupported image format (only jpeg, png, webp, gif allowed)
	ErrInvalidImageDimensions   = errors.New("invalid image dimensions (zero or negative size)")              // invalid image dimensions (zero or negative size)
	ErrImageTooLargeDimensions  = errors.New("image dimensions too large (max 12000x12000 pixels)")           // image dimensions too large (max 12000x12000 pixels)
	ErrEmptyFile                = errors.New("empty file")                                                    // empty file
	ErrNoActionsProvided        = errors.New("at least one action must be specified")                         // at least one action must be specified
	ErrUnsupportedAction        = errors.New("unsupported action type")                                       // unsupported action type
	ErrWatermarkTextRequired    = errors.New("watermark text is required when watermark action is requested") // watermark text is required when watermark action is requested
	ErrResizeDimensionsRequired = errors.New("resize requires width or height greater than zero")             // resize requires width or height greater than zero
	ErrNegativeResizeDimensions = errors.New("width and height must be positive values")                      // width and height must be positive values
	ErrInvalidQualityRange      = errors.New("quality must be between 1 and 100")                             // quality must be between 1 and 100
)
