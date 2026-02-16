package v1

type CreateCommentV1 struct {
	ParentID *int64 `json:"parent_id,omitempty"`
	Content  string `json:"content"`
	Author   string `json:"author"`
}
