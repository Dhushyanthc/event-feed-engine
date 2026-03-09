package models

import "time"

type FeedItem struct {
	UserID int64
	PostID int64
	CreatedAt time.Time
}