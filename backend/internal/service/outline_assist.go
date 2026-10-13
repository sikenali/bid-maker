package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/google/uuid"
)

// 章节标题分词:去编号、按常见分隔切分,供挂载匹配用
var outlineNumPrefix = regexp.MustCompile(`^[第\d一二三四五六七八九十章节\d.、\s]+`)

func sectionKeywords(title string) []string {
	norm := outlineNumPrefix.ReplaceAllString(title, "")
	norm = strings.TrimSpace(norm)
	if norm == "" {
		return nil
	}
	// 长标题整体+2-4字滑动窗口,短标题整体
	kws := []string{norm}
	runes := []rune(norm)
	if len(runes) > 4 {
		for i := 0; i+2 <= len(runes); i += 2 {
			end := i + 4
			if end > len(runes) {
				end = len(runes)
			}
			w := string(runes[i:end])
			if len([]rune(w)) >= 2 {
				kws = append(kws, w)
			}
		}
	}
	return kws
}

func scoreMount(haystack string, kws []string) int {
	score := 0
	for _, k := range kws {
		if k != "" && strings.Contains(haystack, k) {
			score += len([]rune(k))
		}
	}
	return score
}

// affinity 章节标题命中任一亲和词即得 4 分
func affinity(title string, words []string) int {
	for _, w := range words {
		if strings.Contains(title, w) {
			return 4
		}
	}
	return 0
}

// MountRequirements 把解析层输出挂载到大纲章节.
// 规则:技术指标/商务要求/评分点/废标条款按关键词与章节标题+来源节匹配度挂载,
// 每章至多挂 8 条,保证"对着要求写内容"有据可依.
func MountRequirements(parsing *model.BidParsingOutput, sections []*model.Section) {
	if parsing == nil {
		return
	}
	type cand struct {
		ref   string
		typ   string
		score int
		text  string
	}
	for _, sec := range sections {
		kws := sectionKeywords(sec.Title)
		if len(kws) == 0 {
			continue
		}
		// 类型亲和:同类要求天然归属同类章节,先给基础分,再按关键词加权
		techAffinity := affinity(sec.Title, []string{"技术", "方案", "规格", "参数", "配置"})
		bizAffinity := affinity(sec.Title, []string{"商务", "服务", "售后", "培训", "实施", "交付", "质量", "计划"})
		scoreAffinity := affinity(sec.Title, []string{"评分", "方案", "业绩", "报价", "商务", "技术", "资格"})
		dqAffinity := affinity(sec.Title, []string{"投标", "商务", "资格", "保证金"})
		var cands []cand
		for _, t := range parsing.TechnicalIndicators {
			s := techAffinity + scoreMount(t.Name+" "+t.Requirement+" "+t.ClauseRef, kws)
			if t.ClauseRef != "" && strings.Contains(t.ClauseRef, kws[0]) {
				s += 6
			}
			if s > 0 {
				cands = append(cands, cand{ref: t.IndicatorID, typ: "technical_indicator", score: s})
			}
		}
		for _, b := range parsing.CommercialReqs {
			s := bizAffinity + scoreMount(b.Title+" "+b.Description+" "+b.ClauseRef, kws)
			if s > 0 {
				cands = append(cands, cand{ref: b.ReqID, typ: "commercial_req", score: s})
			}
		}
		for _, s := range parsing.ScoringPoints {
			sc := scoreAffinity + scoreMount(s.ScoringItem+" "+s.Description+" "+s.ClauseRef, kws)
			if sc > 0 {
				cands = append(cands, cand{ref: s.ScoringID, typ: "scoring_point", score: sc})
			}
		}
		for _, d := range parsing.DisqualifyClauses {
			ds := dqAffinity + scoreMount(d.ClauseText+" "+d.ClauseRef, kws)
			if ds > 3 { // 废标条款只挂强相关,防噪音
				cands = append(cands, cand{ref: d.ClauseID, typ: "disqualify_clause", score: ds})
			}
		}
		// 按分值降序取前 8
		for i := 0; i < len(cands); i++ {
			for j := i + 1; j < len(cands); j++ {
				if cands[j].score > cands[i].score {
					cands[i], cands[j] = cands[j], cands[i]
				}
			}
		}
		if len(cands) > 8 {
			cands = cands[:8]
		}
		for _, c := range cands {
			sec.Requirements = append(sec.Requirements, &model.ReqRef{Ref: c.ref, Type: c.typ})
		}
		if sec.ResponseStrategy == "" {
			sec.ResponseStrategy = model.StrategySatisfy
		}
		if sec.Status == "" {
			sec.Status = model.ChapterStatusPending
		}
	}
}

var outlineJSONRe = regexp.MustCompile(`(?s)\[.*\]`)

// llmSections 请求模型输出纯 JSON 数组 [{title, level, required}],做清洗归一化
func llmSections(ctx context.Context, gw *LLMGateway, provider, modelName, system, user string) ([]*model.Section, error) {
	if gw == nil {
		return nil, fmt.Errorf("LLM 网关未配置")
	}
	reply, err := gw.Chat(ctx, provider, []map[string]string{
		{"role": "system", "content": system},
		{"role": "user", "content": user},
	}, modelName)
	if err != nil {
		return nil, err
	}
	raw := outlineJSONRe.FindString(reply)
	if raw == "" {
		raw = reply
	}
	var items []struct {
		Title    string `json:"title"`
		Level    int    `json:"level"`
		Required bool   `json:"required"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("大纲 JSON 解析失败: %w", err)
	}
	sections := make([]*model.Section, 0, len(items))
	var lastL1 *model.Section
	for i, it := range items {
		title := strings.TrimSpace(it.Title)
		if title == "" {
			continue
		}
		level := it.Level
		if level < 1 || level > 3 {
			level = 1
		}
		sec := &model.Section{
			SectionID: fmt.Sprintf("sec-%02d", i+1),
			Title:     title,
			Level:     level,
			Order:     i + 1,
			Required:  it.Required || level == 1,
			Source:    model.SourceManual,
			Status:    model.ChapterStatusPending,
		}
		if level > 1 && lastL1 != nil {
			pid := lastL1.SectionID
			sec.ParentID = &pid
			lastL1.Children = append(lastL1.Children, sec.SectionID)
		}
		if level == 1 {
			lastL1 = sec
		}
		sections = append(sections, sec)
	}
	if len(sections) == 0 {
		return nil, fmt.Errorf("模型未返回可用章节")
	}
	return sections, nil
}

// BuildCustomOutline 通道B:完全按用户描述生成大纲
func BuildCustomOutline(ctx context.Context, gw *LLMGateway, provider, modelName, userInput string, hint string) ([]*model.Section, error) {
	system := `你是标书大纲设计专家.只输出纯 JSON 数组,不要输出其他文字.每个元素 {"title": "章节标题(不带编号)", "level": 1|2|3, "required": true|false}.一级章节 4-8 个,关键章节配 2-4 个二级子节.`
	user := fmt.Sprintf("项目背景:\n%s\n\n用户大纲要求:\n%s", hint, userInput)
	return llmSections(ctx, gw, provider, modelName, system, user)
}

// BuildHybridOutline 通道hybrid:自动骨架 + 用户修改指令合并
func BuildHybridOutline(ctx context.Context, gw *LLMGateway, provider, modelName string, base []*model.Section, instruction string) ([]*model.Section, error) {
	var sb strings.Builder
	for _, s := range base {
		fmt.Fprintf(&sb, "- [%s] %s%s\n", s.SectionID, strings.Repeat("  ", s.Level-1), s.Title)
	}
	system := `你是标书大纲设计专家.基于给定自动提取的大纲骨架,按用户修改指令(增删改章节、调顺序)输出最终大纲.只输出纯 JSON 数组,不要输出其他文字.元素 {"title": "章节标题(不带编号)", "level": 1|2|3, "required": true|false}.保留的原章节尽量保持原标题文字.`
	user := fmt.Sprintf("自动提取骨架:\n%s\n\n用户修改指令:\n%s", sb.String(), instruction)
	sections, err := llmSections(ctx, gw, provider, modelName, system, user)
	if err != nil {
		return nil, err
	}
	for _, s := range sections {
		s.Source = model.SourceEdited
	}
	return sections, nil
}

// ValidateOutlineUpdate 按纲约束校验大纲更新:
// locked 后禁止删除章节、禁止删除 required 章节的内容结构变化;
// 标题非空、section_id 唯一;新增子章节自动标 edited.
func ValidateOutlineUpdate(old *model.BidOutline, nu *model.BidOutline) error {
	ids := map[string]bool{}
	if old != nil && old.Locked {
		oldIDs := map[string]*model.Section{}
		for _, s := range old.Sections {
			oldIDs[s.SectionID] = s
		}
		newIDs := map[string]bool{}
		for _, s := range nu.Sections {
			newIDs[s.SectionID] = true
		}
		for id, s := range oldIDs {
			if !newIDs[id] {
				if s.Required {
					return fmt.Errorf("已锁定:禁止删除必需章节 %s", s.Title)
				}
				return fmt.Errorf("已锁定:禁止删除章节 %s(仅允许增补子章节)", s.Title)
			}
		}
	}
	for _, s := range nu.Sections {
		if s.SectionID == "" {
			s.SectionID = "sec-" + uuid.NewString()[:8]
		}
		if s.Source == "" {
			s.Source = model.SourceEdited
		}
		if ids[s.SectionID] {
			return fmt.Errorf("section_id 重复: %s", s.SectionID)
		}
		ids[s.SectionID] = true
		if strings.TrimSpace(s.Title) == "" {
			return fmt.Errorf("章节标题不能为空")
		}
	}
	return nil
}
