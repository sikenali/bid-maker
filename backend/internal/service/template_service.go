package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
)

type TemplateStore struct {
	mu        sync.RWMutex
	templates map[string]model.BidTemplate
}

var globalTemplateStore *TemplateStore
var templateOnce sync.Once

func GetTemplateStore() *TemplateStore {
	templateOnce.Do(func() {
		globalTemplateStore = &TemplateStore{
			templates: make(map[string]model.BidTemplate),
		}
		globalTemplateStore.seed()
	})
	return globalTemplateStore
}

func (s *TemplateStore) seed() {
	s.templates["tpl-procurement-goods"] = model.BidTemplate{
		ID:          "tpl-procurement-goods",
		Name:        "政府采购货物类",
		Description: "适用于货物类采购项目",
		Category:    "政府采购",
		Icon:        "RiFileTextLine",
		Outline: []model.BidSection{
			{
				ID: "tpl-sec-1", Title: "商务部分", Level: 1, Content: "",
				Children: []model.BidSection{
					{ID: "tpl-sec-1-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-1-2", Title: "法定代表人授权书", Level: 2, Content: ""},
					{ID: "tpl-sec-1-3", Title: "投标报价一览表", Level: 2, Content: ""},
					{ID: "tpl-sec-1-4", Title: "分项报价明细表", Level: 2, Content: ""},
					{ID: "tpl-sec-1-5", Title: "商务条款响应/偏离表", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-2", Title: "技术部分", Level: 1, Content: "",
				Children: []model.BidSection{
					{ID: "tpl-sec-2-1", Title: "项目理解与需求分析", Level: 2, Content: ""},
					{ID: "tpl-sec-2-2", Title: "技术方案设计", Level: 2, Content: ""},
					{ID: "tpl-sec-2-3", Title: "产品配置及技术参数", Level: 2, Content: ""},
					{ID: "tpl-sec-2-4", Title: "技术条款响应/偏离表", Level: 2, Content: ""},
					{ID: "tpl-sec-2-5", Title: "项目实施计划", Level: 2, Content: ""},
				},
			},
		},
	}

	s.templates["tpl-construction"] = model.BidTemplate{
		ID:          "tpl-construction",
		Name:        "工程施工类",
		Description: "适用于工程施工招标",
		Category:    "工程建设",
		Icon:        "RiBuildingLine",
		Outline: []model.BidSection{
			{
				ID: "tpl-sec-5", Title: "投标函及报价", Level: 1, Content: "",
				Children: []model.BidSection{
					{ID: "tpl-sec-5-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-5-2", Title: "投标报价汇总表", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-6", Title: "施工组织设计", Level: 1, Content: "",
				Children: []model.BidSection{
					{ID: "tpl-sec-6-1", Title: "工程概况", Level: 2, Content: ""},
					{ID: "tpl-sec-6-2", Title: "施工总体部署", Level: 2, Content: ""},
				},
			},
		},
	}
}

func (s *TemplateStore) List() []model.BidTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.BidTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		result = append(result, t)
	}
	return result
}

func (s *TemplateStore) Get(id string) (model.BidTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[id]
	return t, ok
}

func (s *TemplateStore) Save(t model.BidTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = fmt.Sprintf("tpl-%d", time.Now().UnixNano())
	}
	s.templates[t.ID] = t
}

func (s *TemplateStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.templates, id)
}
