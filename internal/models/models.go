package models

import "mime/multipart"

const (
	StatusPending = "pending" // pending
)

type Image struct {
	ID         string
	Size       int64
	ObjectKey  string
	File       []byte
	FileHeader *multipart.FileHeader
	Prefix     string
}
