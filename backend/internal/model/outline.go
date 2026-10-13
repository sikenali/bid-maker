package model

import "time"

// BidOutline 对应技术方案的规划层大纲Schema
type BidOutline struct {
	OutlineID  string    `json:"outline_id"`
	SourceMode string    `json:"source_mode"`
	Locked     bool      `json:"locked"`
	Sections   []*Section `json:"sections"`
}

type Section struct {
	SectionID        string    `json:"section_id"`
	Title            string    `json:"title"`
	Level            int       `json:"level"`
	Order            int       `json:"order"`
	Required         bool      `json:"required"`
	ParentID         *string   `json:"parent_id"`
	Source           string    `json:"source"`
	Requirements     []*ReqRef `json:"requirements"`
	ResponseStrategy string    `json:"response_strategy"`
	ResponseBasis    string    `json:"response_basis"`
	Children         []string  `json:"children"`
	Content          string    `json:"content,omitempty"`
	Status           string    `json:"status,omitempty"`
	GeneratedAt      *time.Time `json:"generated_at,omitempty"`
}

type ReqRef struct {
	Ref  string `json:"ref"`
	Type string `json:"type"`
}

const (
	SourceAuto      = "auto"
	SourceManual    = "manual"
	SourceEdited    = "edited"
	
	StrategySatisfy = "satisfy"
	StrategyDeviate = "deviate"
	StrategyExceed  = "exceed"
	
	ChapterStatusPending    = "pending"
	ChapterStatusGenerating = "generating"
	ChapterStatusGenerated  = "generated"
	ChapterStatusFailed     = "failed"
)
