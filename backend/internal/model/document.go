package model

import "time"

type BidSection struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Level    int         `json:"level"`
	ParentID string      `json:"parent_id"`
	Content  string      `json:"content"`
	Children []BidSection `json:"children"`
}

type BidDocument struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Outline   []BidSection  `json:"outline"`
	Markdown  string        `json:"markdown"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
