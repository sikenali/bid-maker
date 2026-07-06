package model

import "time"

type Section struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Level    int       `json:"level"`
	ParentID string    `json:"parent_id"`
	Content  string    `json:"content"`
	Children []Section `json:"children"`
}

type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Outline   []Section `json:"outline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
