package errs

import "errors"

var (
	ErrInvalidJSON      = errors.New("invalid JSON format")              // invalid JSON format
	ErrEmptyContent     = errors.New("comment content can not be empty") // comment content can not be empty
	ErrEmptyAuthor      = errors.New("comment author can not be empty")  // comment author can not be empty
	ErrCommentNotFound  = errors.New("comment not found")                // comment not found
	ErrParentNotFound   = errors.New("parent comment not found")         // parent comment not found
	ErrInternal         = errors.New("internal server error")            // internal server error
	ErrInvalidParentID  = errors.New("invalid parent id")                // invalid parent id
	ErrEmptyCommentID   = errors.New("comment id can not be empty")      // comment id can not be empty
	ErrInvalidCommentID = errors.New("comment id is invalid")            // comment id is invalid
	ErrInvalidPage      = errors.New("invalid page number")              // invalid page number
	ErrInvalidLimit     = errors.New("invalid limit")                    // invalid limit
	ErrInvalidSort      = errors.New("invalid sort value")               // invalid sort value
)
