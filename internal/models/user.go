package models

import "time"

type User struct {
	Id int64
	Name string
	Email string
	PasswordHash string
	CreatedAt time.Time
}