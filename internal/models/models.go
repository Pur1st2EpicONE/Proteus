package models

import (
	"mime/multipart"
)

const (
	StatusPending = "pending" // pending
)

const (
	Resize    = "resize"
	Thumbnail = "thumbnail"
	Watermark = "watermark"
)

type Image struct {
	ID         string
	Size       int64
	ObjectKey  string
	File       []byte
	FileHeader *multipart.FileHeader
	Prefix     string
	Request    Request
}

type Request struct {
	Action    []string
	Watermark string
	Height    int
	Width     int
	Quality   int
}

type ImageProcessTask struct {
	ID           string   `json:"id"`
	ObjectKey    string   `json:"object_key"`
	OriginalName string   `json:"original_name"`
	MimeType     string   `json:"mime_type"`
	FileSize     int64    `json:"file_size"`
	Action       []string `json:"action"`
	Watermark    string   `json:"watermark"`
	Height       int      `json:"height"`
	Width        int      `json:"width "`
	Quality      int      `json:"quality"`
}
