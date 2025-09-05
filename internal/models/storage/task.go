package storage

import "time"

type Task struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	IsDone      bool      `db:"is_done"`
	CreatedAt   time.Time `db:"created_at"`
	Date        string    `db:"date"`
	Repeat      string    `db:"repeat"`
}
