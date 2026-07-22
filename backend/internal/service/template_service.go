package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
)

type TemplateStore struct {
	mu        sync.RWMutex
	templates map[string]model.Template
}

var globalTemplateStore *TemplateStore
var templateOnce sync.Once

func GetTemplateStore() *TemplateStore {
	templateOnce.Do(func() {
		globalTemplateStore = &TemplateStore{
			templates: make(map[string]model.Template),
		}
		globalTemplateStore.seed()
	})
	return globalTemplateStore
}

func (s *TemplateStore) seed() {
	s.templates["tpl-procurement-goods"] = model.Template{
		ID:          "tpl-procurement-goods",
		Name:        "政府采购货物类",
		Description: "适用于货物类采购项目",
		Category:    "政府采购",
		Icon:        "RiFileTextLine",
		Outline: []model.Section{
			{
				ID: "tpl-sec-1", Title: "商务部分", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-1-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-1-2", Title: "法定代表人授权书", Level: 2, Content: ""},
					{ID: "tpl-sec-1-3", Title: "投标报价一览表", Level: 2, Content: ""},
					{ID: "tpl-sec-1-4", Title: "分项报价明细表", Level: 2, Content: ""},
					{ID: "tpl-sec-1-5", Title: "商务条款响应/偏离表", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-2", Title: "技术部分", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-2-1", Title: "项目理解与需求分析", Level: 2, Content: ""},
					{ID: "tpl-sec-2-2", Title: "技术方案设计", Level: 2, Content: ""},
					{ID: "tpl-sec-2-3", Title: "产品配置及技术参数", Level: 2, Content: ""},
					{ID: "tpl-sec-2-4", Title: "技术条款响应/偏离表", Level: 2, Content: ""},
					{ID: "tpl-sec-2-5", Title: "项目实施计划", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-3", Title: "服务部分", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-3-1", Title: "售后服务方案", Level: 2, Content: ""},
					{ID: "tpl-sec-3-2", Title: "培训方案", Level: 2, Content: ""},
					{ID: "tpl-sec-3-3", Title: "质量保证措施", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-4", Title: "资格证明文件", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-4-1", Title: "营业执照", Level: 2, Content: ""},
					{ID: "tpl-sec-4-2", Title: "财务状况报告", Level: 2, Content: ""},
					{ID: "tpl-sec-4-3", Title: "依法缴纳税收和社会保障资金记录", Level: 2, Content: ""},
					{ID: "tpl-sec-4-4", Title: "参加政府采购活动前三年内无重大违法记录声明", Level: 2, Content: ""},
				},
			},
		},
	}

	s.templates["tpl-construction"] = model.Template{
		ID:          "tpl-construction",
		Name:        "工程施工类",
		Description: "适用于工程施工招标",
		Category:    "工程建设",
		Icon:        "RiBuildingLine",
		Outline: []model.Section{
			{
				ID: "tpl-sec-5", Title: "投标函及报价", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-5-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-5-2", Title: "投标报价汇总表", Level: 2, Content: ""},
					{ID: "tpl-sec-5-3", Title: "工程量清单报价表", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-6", Title: "施工组织设计", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-6-1", Title: "工程概况", Level: 2, Content: ""},
					{ID: "tpl-sec-6-2", Title: "施工总体部署", Level: 2, Content: ""},
					{ID: "tpl-sec-6-3", Title: "主要施工方法", Level: 2, Content: ""},
					{ID: "tpl-sec-6-4", Title: "施工进度计划及保证措施", Level: 2, Content: ""},
					{ID: "tpl-sec-6-5", Title: "质量保证体系及措施", Level: 2, Content: ""},
					{ID: "tpl-sec-6-6", Title: "安全文明施工措施", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-7", Title: "项目管理机构", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-7-1", Title: "项目管理团队配置", Level: 2, Content: ""},
					{ID: "tpl-sec-7-2", Title: "主要人员简历及资格证书", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-8", Title: "资格证明文件", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-8-1", Title: "企业资质证书", Level: 2, Content: ""},
					{ID: "tpl-sec-8-2", Title: "类似项目业绩", Level: 2, Content: ""},
				},
			},
		},
	}

	s.templates["tpl-it-service"] = model.Template{
		ID:          "tpl-it-service",
		Name:        "信息化服务类",
		Description: "适用于IT服务采购",
		Category:    "IT服务",
		Icon:        "RiServerLine",
		Outline: []model.Section{
			{
				ID: "tpl-sec-9", Title: "投标响应函", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-9-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-9-2", Title: "法定代表人授权委托书", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-10", Title: "项目理解与解决方案", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-10-1", Title: "项目背景及需求理解", Level: 2, Content: ""},
					{ID: "tpl-sec-10-2", Title: "总体解决方案", Level: 2, Content: ""},
					{ID: "tpl-sec-10-3", Title: "系统架构设计", Level: 2, Content: ""},
					{ID: "tpl-sec-10-4", Title: "功能详细设计", Level: 2, Content: ""},
					{ID: "tpl-sec-10-5", Title: "技术路线与选型说明", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-11", Title: "项目实施与服务", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-11-1", Title: "项目实施方法论", Level: 2, Content: ""},
					{ID: "tpl-sec-11-2", Title: "项目进度计划", Level: 2, Content: ""},
					{ID: "tpl-sec-11-3", Title: "质量保障措施", Level: 2, Content: ""},
					{ID: "tpl-sec-11-4", Title: "培训及知识转移方案", Level: 2, Content: ""},
					{ID: "tpl-sec-11-5", Title: "售后服务及技术支持", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-12", Title: "公司实力与资质", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-12-1", Title: "公司简介", Level: 2, Content: ""},
					{ID: "tpl-sec-12-2", Title: "相关资质证书", Level: 2, Content: ""},
					{ID: "tpl-sec-12-3", Title: "典型成功案例", Level: 2, Content: ""},
				},
			},
		},
	}

	s.templates["tpl-consulting"] = model.Template{
		ID:          "tpl-consulting",
		Name:        "咨询服务类",
		Description: "适用于咨询类采购",
		Category:    "咨询服务",
		Icon:        "RiCustomerServiceLine",
		Outline: []model.Section{
			{
				ID: "tpl-sec-13", Title: "投标函", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-13-1", Title: "投标函", Level: 2, Content: ""},
					{ID: "tpl-sec-13-2", Title: "报价明细表", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-14", Title: "咨询方案", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-14-1", Title: "项目理解与需求分析", Level: 2, Content: ""},
					{ID: "tpl-sec-14-2", Title: "咨询方法论", Level: 2, Content: ""},
					{ID: "tpl-sec-14-3", Title: "工作内容与交付物", Level: 2, Content: ""},
					{ID: "tpl-sec-14-4", Title: "工作计划与时间安排", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-15", Title: "项目团队", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-15-1", Title: "项目团队组成", Level: 2, Content: ""},
					{ID: "tpl-sec-15-2", Title: "核心成员简历", Level: 2, Content: ""},
				},
			},
			{
				ID: "tpl-sec-16", Title: "公司资质与经验", Level: 1, Content: "",
				Children: []model.Section{
					{ID: "tpl-sec-16-1", Title: "公司概况", Level: 2, Content: ""},
					{ID: "tpl-sec-16-2", Title: "相关咨询业绩", Level: 2, Content: ""},
				},
			},
		},
	}
}

func (s *TemplateStore) List() []model.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Template, 0, len(s.templates))
	for _, t := range s.templates {
		result = append(result, t)
	}
	return result
}

func (s *TemplateStore) Get(id string) (model.Template, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.templates[id]
	return t, ok
}

func (s *TemplateStore) Save(t model.Template) {
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
