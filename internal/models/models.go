// Package models defines the core domain models, status/action constants
// and data transfer structures used throughout the Proteus image processing
// service (API layer, service layer, storage, Kafka messages, etc.).
package models

import (
	"mime/multipart"
)

// Image status constants used in meta-storage, API responses and
// business logic to represent the lifecycle of an image.
const (
	StatusPending = "pending" // StatusPending means the image has been accepted and is either queued for processing or currently being processed asynchronously.
	StatusDeleted = "deleted" // StatusDeleted means the image was soft-deleted by the user.
	StatusReady   = "ready"   // StatusReady means the image has been successfully processed and is available for download.
)

// Supported processing action constants
// that can be requested via the upload API.
const (
	Resize    = "resize"    // Resize instructs the service to resize the image to the provided width/height.
	Thumbnail = "thumbnail" // Thumbnail instructs the service to create a square thumbnail (uses width/height).
	Watermark = "watermark" // Watermark instructs the service to overlay text on the image.
)

type Image struct {
	ID          string                // ID is the unique UUIDv4 identifier of the image (used as primary key in meta DB).
	Size        int64                 // Size is the original file size in bytes (stored in meta DB).
	ObjectKey   string                // ObjectKey is the MinIO storage key under which the original/processed image resides.
	File        []byte                // File holds the raw image bytes (populated only during upload, not persisted).
	ContentType string                // ContentType is the detected MIME type of the image (image/jpeg, image/png, etc.).
	FileHeader  *multipart.FileHeader // FileHeader is the original multipart header received from the client (used only during upload).
	Prefix      string                // Prefix is an optional storage prefix for the object key (future extensibility).
	Status      string                // Status is the current processing status (one of StatusPending/StatusReady/StatusDeleted).
	Request     Request               // Request contains the original user-requested processing parameters.
}

type Request struct {
	Action    string // Action to apply (e.g. "resize,watermark").
	Watermark string // Watermark is the text to overlay when the watermark action is requested.
	Height    int    // Height is the target height in pixels (used by resize/thumbnail actions).
	Width     int    // Width is the target width in pixels (used by resize/thumbnail actions).
}

// ImageProcessTask is the Kafka message payload sent by the upload handler
// and consumed by the background worker. All fields are JSON-serialized.
type ImageProcessTask struct {
	ID           string `json:"id"`            // ID is the unique image identifier (same as models.Image.ID).
	ObjectKey    string `json:"object_key"`    // ObjectKey is the MinIO key of the original uploaded image.
	OriginalName string `json:"original_name"` // OriginalName is the client-provided filename.
	ContentType  string `json:"content_type"`  // ContentType is the MIME type of the original image.
	FileSize     int64  `json:"file_size"`     // FileSize is the size of the original file in bytes.
	Action       string `json:"action"`        // Action is the comma-separated list of processing actions to perform.
	Watermark    string `json:"watermark"`     // Watermark is the text to overlay (empty if not requested).
	Height       int    `json:"height"`        // Height is the target height in pixels for resize/thumbnail.
	Width        int    `json:"width "`        // Width is the target width in pixels for resize/thumbnail.
}
