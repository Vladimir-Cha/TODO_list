package api

import "time"

type Task struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Title     string    `json:"title"`
	Comment   string    `json:"comment"`
	IsDone    bool      `json:"is_done"`
	Repeat    string    `json:"repeat"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
