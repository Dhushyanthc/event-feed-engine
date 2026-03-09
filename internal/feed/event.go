package feed

import "time"

type PostCreatedEvent struct {
	PostID  int64
	UserID  int64
	CreatedAt time.Time
}