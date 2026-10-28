package domain

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
}

func (User) TableName() string { return "users" }
