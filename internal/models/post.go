package models

import "time"

type Post struct {
	Id int64 
	UserId int64 
	Content string 
	MediaURL string
	CreatedAt time.Time
}