package domain

import "time"

type User struct {
	UserID    string     `json:"user_id"`
	Username  string     `json:"username"`
	TeamName  string     `json:"team_name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}
