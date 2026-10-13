package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
)

// ParseClient 调用解析微服务(officeparser 提取管线,见 ../parse-service).
type ParseClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewParseClient(baseURL string) *ParseClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return &ParseClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// Health 检查解析微服务是否在线.
func (c *ParseClient) Health() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("解析微服务不可达 %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("解析微服务健康检查失败: status %d", resp.StatusCode)
	}
	return nil
}

// ParseFile 上传本地文件并返回结构化解析结果.
func (c *ParseClient) ParseFile(filePath string) (*model.BidParsingOutput, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()
	return c.ParseReader(filepath.Base(filePath), f)
}

// ParseReader 上传文件流并返回结构化解析结果.
func (c *ParseClient) ParseReader(filename string, r io.Reader) (*model.BidParsingOutput, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("构造上传表单失败: %w", err)
	}
	if _, err := io.Copy(fw, r); err != nil {
		return nil, fmt.Errorf("写入上传内容失败: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("关闭上传表单失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/parse", &body)
	if err != nil {
		return nil, fmt.Errorf("构造解析请求失败: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用解析微服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("解析微服务返回 %d: %s", resp.StatusCode, string(raw))
	}

	var out model.BidParsingOutput
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("解析结果解码失败: %w", err)
	}
	// 废标条款独立且必填:空时兜底为空数组,保证下游检查工具可用
	if out.DisqualifyClauses == nil {
		out.DisqualifyClauses = []*model.DisqualifyClause{}
	}
	if out.TechnicalIndicators == nil {
		out.TechnicalIndicators = []*model.TechnicalIndicator{}
	}
	if out.CommercialReqs == nil {
		out.CommercialReqs = []*model.CommercialReq{}
	}
	if out.ScoringPoints == nil {
		out.ScoringPoints = []*model.ScoringPoint{}
	}
	return &out, nil
}
