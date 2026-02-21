package v1

type UploadImageDTO struct {
	Action    string `form:"action"`
	Watermark string `form:"watermark"`
	Height    int    `form:"height"`
	Width     int    `form:"width"`
	Quality   int    `form:"quality"`
}
