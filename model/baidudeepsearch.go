package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// 外部模块可以直接调用 BaiduDeepSearch() 或 BaiduDeepSearchSimple() 进行搜索，无需关心客户端初始化细节！
// API官方文档： https://cloud.baidu.com/doc/qianfan/s/Omh4su4s0
// ============== 示例用法 ==============
func deepsearchdemo() {
	// 方式1: 使用简化接口(默认配置)
	response, err := BaiduDeepSearchSimple("按照时间线梳理川普生平")
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}
	fmt.Printf("搜索成功:\n%s\n", response.RawResponse)

	// 方式2: 使用自定义配置
	// customOptions := NewDefaultRequest("")
	// customOptions.EnableDeepSearch = true
	// customOptions.MaxCompletionTokens = 4096
	// customOptions.SearchRecencyFilter = "month"
	// response, err := BaiduDeepSearch("按照时间线梳理川普生平", customOptions)

	// 方式3: 修改默认API Key
	// SetDefaultAPIKey("your-new-api-key")
	// response, err := BaiduDeepSearchSimple("查询内容")
}

// ============== 常量定义 ==============

const (
	BaiduDeepSearchURL = "https://qianfan.baidubce.com/v2/ai_search/chat/completions"
	DefaultModel       = "ernie-3.5-8k"
	DefaultTemperature = "1e-10"
	DefaultTopP        = "1e-10"
)

// ============== 请求相关结构体 ==============

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResourceTypeFilter 资源类型过滤器
type ResourceTypeFilter struct {
	Type string `json:"type"`
	TopK int    `json:"top_k"`
}

// BaiduDeepSearchRequest 百度深度搜索请求参数
type BaiduDeepSearchRequest struct {
	Messages              []Message            `json:"messages"`
	SearchSource          string               `json:"search_source"`
	ResourceTypeFilter    []ResourceTypeFilter `json:"resource_type_filter"`
	SearchRecencyFilter   string               `json:"search_recency_filter"`
	Model                 string               `json:"model"`
	Temperature           string               `json:"temperature"`
	TopP                  string               `json:"top_p"`
	SearchMode            string               `json:"search_mode"`
	EnableReasoning       bool                 `json:"enable_reasoning"`
	EnableDeepSearch      bool                 `json:"enable_deep_search"`
	MaxCompletionTokens   int                  `json:"max_completion_tokens"`
	ResponseFormat        string               `json:"response_format"`
	EnableCornerMarkers   bool                 `json:"enable_corner_markers"`
	EnableFollowupQueries bool                 `json:"enable_followup_queries"`
	Stream                bool                 `json:"stream"`
	SafetyLevel           string               `json:"safety_level"`
	MaxSearchQueryNum     int                  `json:"max_search_query_num"`
}

// ============== 响应相关结构体 ==============

// BaiduDeepSearchResponse 百度深度搜索响应
type BaiduDeepSearchResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Result  string `json:"result"`
	// 根据实际响应结构添加更多字段
	RawResponse string `json:"-"` // 原始响应，不序列化
}

// ============== 客户端配置 ==============

// BaiduDeepSearchClient 百度深度搜索客户端
type BaiduDeepSearchClient struct {
	apiKey     string
	httpClient *http.Client
}

// 全局客户端实例
var defaultClient *BaiduDeepSearchClient

// init 初始化默认客户端
func init() {
	// 从环境变量获取API Key初始化客户端，如果未设置则使用默认值
	apiKey := os.Getenv("BAIDU_DEEPSEARCH_API_KEY")
	if apiKey == "" {
		apiKey = "bce-v3/ALTAK-jQLDiSgUGQoD1MPDkhPmt/46c9e4155b9a95a0e339dcdb3fd47e97048e53ba"
	}
	// 使用默认API Key初始化客户端
	defaultClient = &BaiduDeepSearchClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// NewBaiduDeepSearchClient 创建新的客户端实例
func NewBaiduDeepSearchClient(apiKey string) *BaiduDeepSearchClient {
	return &BaiduDeepSearchClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SetDefaultAPIKey 设置默认客户端的API Key
func SetDefaultAPIKey(apiKey string) {
	defaultClient.apiKey = apiKey
}

// ============== 请求构建函数 ==============

// NewDefaultRequest 创建默认请求参数
func NewDefaultRequest(userMessage string) *BaiduDeepSearchRequest {
	return &BaiduDeepSearchRequest{
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
		SearchSource: "baidu_search_v1",
		ResourceTypeFilter: []ResourceTypeFilter{
			{Type: "image", TopK: 4},
			{Type: "video", TopK: 4},
			{Type: "web", TopK: 4},
		},
		SearchRecencyFilter:   "week",
		Model:                 DefaultModel,
		Temperature:           DefaultTemperature,
		TopP:                  DefaultTopP,
		SearchMode:            "auto",
		EnableReasoning:       true,
		EnableDeepSearch:      false,
		MaxCompletionTokens:   2048,
		ResponseFormat:        "auto",
		EnableCornerMarkers:   true,
		EnableFollowupQueries: false,
		Stream:                false,
		SafetyLevel:           "standard",
		MaxSearchQueryNum:     10,
	}
}

// ============== HTTP 请求辅助函数 ==============

// buildHTTPRequest 构建HTTP请求
func (c *BaiduDeepSearchClient) buildHTTPRequest(req *BaiduDeepSearchRequest) (*http.Request, error) {
	// 序列化请求体
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求参数失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", BaiduDeepSearchURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	return httpReq, nil
}

// executeRequest 执行HTTP请求
func (c *BaiduDeepSearchClient) executeRequest(httpReq *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("执行HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// parseResponse 解析响应
func (c *BaiduDeepSearchClient) parseResponse(body []byte) (*BaiduDeepSearchResponse, error) {
	var response BaiduDeepSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// 如果解析失败，仍然返回原始响应
		response.RawResponse = string(body)
		return &response, fmt.Errorf("解析响应失败: %w", err)
	}
	response.RawResponse = string(body)
	return &response, nil
}

// ============== 核心调用函数 ==============

// Search 执行百度深度搜索
func (c *BaiduDeepSearchClient) Search(req *BaiduDeepSearchRequest) (*BaiduDeepSearchResponse, error) {
	// 1. 构建HTTP请求
	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return nil, err
	}

	// 2. 执行HTTP请求
	body, err := c.executeRequest(httpReq)
	if err != nil {
		return nil, err
	}

	// 3. 解析响应
	return c.parseResponse(body)
}

// SearchWithMessage 使用简化接口执行搜索
func (c *BaiduDeepSearchClient) SearchWithMessage(message string) (*BaiduDeepSearchResponse, error) {
	req := NewDefaultRequest(message)
	return c.Search(req)
}

// ============== 外部调用接口 ==============

// BaiduDeepSearch 供外部调用的百度深度搜索函数
// 参数:
//   - message: 用户查询消息
//   - options: 可选配置参数(可传nil使用默认配置)
//
// 返回:
//   - 搜索响应和错误信息
func BaiduDeepSearch(message string, options *BaiduDeepSearchRequest) (*BaiduDeepSearchResponse, error) {
	var req *BaiduDeepSearchRequest

	if options == nil {
		// 使用默认配置
		req = NewDefaultRequest(message)
	} else {
		// 使用自定义配置
		req = options
		// 确保消息被设置
		if len(req.Messages) == 0 {
			req.Messages = []Message{{Role: "user", Content: message}}
		} else {
			req.Messages[0].Content = message
		}
	}

	return defaultClient.Search(req)
}

// BaiduDeepSearchSimple 简化版深度搜索(使用默认配置)
func BaiduDeepSearchSimple(message string) (*BaiduDeepSearchResponse, error) {
	return BaiduDeepSearch(message, nil)
}
