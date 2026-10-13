package model

import "time"

// Session 代表一次标书编制会话
type Session struct {
	SessionID      string            `json:"session_id"`
	Title          string            `json:"title"`
	BidNo          string            `json:"bid_no"`
	ParsingOutput  *BidParsingOutput `json:"parsing_output,omitempty"`
	Outline        *BidOutline       `json:"outline,omitempty"`
	Markdown       string            `json:"markdown"`
	CurrentSection string            `json:"current_section,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}
