package model

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type PublicUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func (u User) Public() PublicUser {
	return PublicUser{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Role: u.Role}
}
