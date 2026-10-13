package model

import ()

// BidParsingOutput 对应技术方案的解析层输出Schema
type BidParsingOutput struct {
	SourceMeta          *SourceMeta           `json:"source_meta"`
	Template            *ParsingTemplate      `json:"template"`
	TechnicalIndicators []*TechnicalIndicator `json:"technical_indicators"`
	CommercialReqs      []*CommercialReq      `json:"commercial_reqs"`
	ScoringPoints       []*ScoringPoint       `json:"scoring_points"`
	DisqualifyClauses   []*DisqualifyClause   `json:"disqualify_clauses"`
}

type SourceMeta struct {
	Title      string `json:"title"`
	BidNo      string `json:"bid_no"`
	Issuer     string `json:"issuer"`
	ParseTime  string `json:"parse_time"`
	FileHash   string `json:"file_hash"`
	TotalPages int    `json:"total_pages"`
}

type ParsingTemplate struct {
	HasPresetStructure bool              `json:"has_preset_structure"`
	PresetSections     []*SectionNode    `json:"preset_sections"`
	FormatRules        *FormatRules      `json:"format_rules"`
}

type SectionNode struct {
	SectionID string  `json:"section_id"`
	Title     string  `json:"title"`
	Level     int     `json:"level"`
	Order     int     `json:"order"`
	Required  bool    `json:"required"`
	ParentID  *string `json:"parent_id"`
}

type FormatRules struct {
	Font      string   `json:"font"`
	FontSize  string   `json:"font_size"`
	HasCover  bool     `json:"has_cover"`
	HasToc    bool     `json:"has_toc"`
	SealPages []string `json:"seal_pages"`
	Other     map[string]any `json:"other"`
}

type TechnicalIndicator struct {
	IndicatorID  string  `json:"indicator_id"`
	Name         string  `json:"name"`
	Requirement  string  `json:"requirement"`
	ClauseRef    string  `json:"clause_ref"`
	IsMandatory  bool    `json:"is_mandatory"`
	Unit         *string `json:"unit"`
	MinValue     *float64 `json:"min_value"`
	MaxValue     *float64 `json:"max_value"`
	ResponseType string  `json:"response_type"`
}

type CommercialReq struct {
	ReqID       string   `json:"req_id"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ClauseRef   string   `json:"clause_ref"`
	Deliverable []string `json:"deliverable"`
}

type ScoringPoint struct {
	ScoringID    string       `json:"scoring_id"`
	ScoringItem  string       `json:"scoring_item"`
	FullScore    float64      `json:"full_score"`
	Description  string       `json:"description"`
	ClauseRef    string       `json:"clause_ref"`
	GradingTiers []*GradingTier `json:"grading_tiers"`
	RequiredDocs []string     `json:"required_docs"`
}

type GradingTier struct {
	Tier     string `json:"tier"`
	Range    string `json:"range"`
	Criteria string `json:"criteria"`
}

type DisqualifyClause struct {
	ClauseID     string   `json:"clause_id"`
	ClauseText   string   `json:"clause_text"`
	ClauseRef    string   `json:"clause_ref"`
	Severity     string   `json:"severity"`
	CheckType    string   `json:"check_type"`
	RequiredDocs []string `json:"required_docs"`
	Deadline     *string  `json:"deadline"`
}
