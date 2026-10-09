package models

import "time"

// Todo maps to one row of the todos table (see migrations/001_create_todos.sql).
// The same struct is passed to HTML templates, so {{.Title}} reads the Title field.
type Todo struct {
	ID        int64
	Title     string
	Completed bool
	CreatedAt time.Time
}
