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

type ImageProcessTask struct {
	ID           string   `json:"id"`
	ObjectKey    string   `json:"object_key"`
	OriginalName string   `json:"original_name"`
	MimeType     string   `json:"mime_type"`
	FileSize     int64    `json:"file_size"`
	Requested    []string `json:"requested"`
}
