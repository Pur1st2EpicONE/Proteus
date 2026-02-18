package v1

import "mime/multipart"

type UploadImageDTO struct {
	File        multipart.File
	Header      *multipart.FileHeader
	ContentType string
	Prefix      string
}
