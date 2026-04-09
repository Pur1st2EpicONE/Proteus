package v1

// UploadImageDTO represents the multipart form data expected when
// uploading an image for processing.
type UploadImageDTO struct {
	Action    string `form:"action"`    // Action is a processing action (resize, watermark, etc.).
	Watermark string `form:"watermark"` // Watermark is the text to overlay when the watermark action is requested.
	Height    int    `form:"height"`    // Height is the desired height in pixels (used by resize action).
	Width     int    `form:"width"`     // Width is the desired width in pixels (used by resize action).
}
