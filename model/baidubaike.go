package model

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// API官方文档：https://cloud.baidu.com/doc/qianfan/s/bmh4stpbh
// ============== 示例用法 ==============
func baikedemo() {
	// 方式1: 使用简化接口(默认配置)
	response, err := BaiduBaikeSearchSimple("刘德华")
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}
	fmt.Printf("搜索成功:\n%s\n", response.RawResponse)

	// 方式2: 使用自定义配置
	// customOptions := &BaiduBaikeRequest{
	//     SearchType: "lemmaTitle",
	//     SearchKey:  "刘德华",
	//     RequestId:  "custom-request-id",
	// }
	// response, err := BaiduBaikeSearch(customOptions)

	// 方式3: 修改默认API Key
	// SetBaiduBaikeAPIKey("your-new-api-key")
	// response, err := BaiduBaikeSearchSimple("刘德华")
}

// ============== 常量定义 ==============

const (
	BaiduBaikeURL         = "https://appbuilder.baidu.com/v2/baike/lemma/get_content"
	DefaultSearchType     = "lemmaTitle" // lemmaTitle: 词条标题搜索, lemmaId: 词条ID搜索
	DefaultBaikeRequestID = ""
)

// ============== 请求相关结构体 ==============

// BaiduBaikeRequest 百度百科请求参数
type BaiduBaikeRequest struct {
	SearchType string `json:"search_type"` // lemmaTitle 或 lemmaId
	SearchKey  string `json:"search_key"`  // 搜索关键词或词条ID
	RequestId  string `json:"request_id"`  // 请求ID(可选)
}

// ============== 响应相关结构体 ==============

// BaikeRelation 实体关系
type BaikeRelation struct {
	LemmaId      int64  `json:"lemma_id"`       // 关联词条ID
	LemmaTitle   string `json:"lemma_title"`    // 关联词条标题
	RelationName string `json:"relation_name"`  // 关系名称
	SquarePicURL string `json:"square_pic_url"` // 方形图片URL
}

// BaikeResult 百度百科结果
type BaikeResult struct {
	ContentPlain       *string                  `json:"content_plain"`       // 纯内容(可能为null)
	LemmaDesc          string                   `json:"lemma_desc"`          // 词条描述
	LemmaId            int64                    `json:"lemma_id"`            // 词条ID
	LemmaTitle         string                   `json:"lemma_title"`         // 词条标题
	URL                string                   `json:"url"`                 // 词条URL
	Summary            string                   `json:"summary"`             // 词条摘要
	AbstractPlain      string                   `json:"abstract_plain"`      // 纯文本摘要
	AbstractHTML       string                   `json:"abstract_html"`       // HTML格式摘要
	AbstractStructured []map[string]interface{} `json:"abstract_structured"` // 结构化摘要
	PicURL             string                   `json:"pic_url"`             // 图片URL
	Relations          []BaikeRelation          `json:"relations"`           // 关系列表
}

// BaiduBaikeResponse 百度百科响应
type BaiduBaikeResponse struct {
	RequestId string       `json:"request_id"` // 请求ID
	Result    *BaikeResult `json:"result"`     // 搜索结果

	RawResponse string `json:"-"` // 原始响应，不序列化
}

// ============== 客户端配置 ==============

// BaiduBaikeClient 百度百科客户端
type BaiduBaikeClient struct {
	apiKey     string
	httpClient *http.Client
}

// 全局客户端实例
var defaultBaikeClient *BaiduBaikeClient

// init 初始化默认客户端
func init() {
	if defaultBaikeClient == nil {
		// 从环境变量获取API Key初始化客户端，如果未设置则使用默认值
		apiKey := os.Getenv("BAIDU_BAIKE_API_KEY")
		if apiKey == "" {
			apiKey = "bce-v3/ALTAK-jQLDiSgUGQoD1MPDkhPmt/46c9e4155b9a95a0e339dcdb3fd47e97048e53ba"
		}
		// 使用默认API Key初始化客户端
		defaultBaikeClient = &BaiduBaikeClient{
			apiKey:     apiKey,
			httpClient: &http.Client{},
		}
	}
}

// NewBaiduBaikeClient 创建新的客户端实例
func NewBaiduBaikeClient(apiKey string) *BaiduBaikeClient {
	return &BaiduBaikeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// SetBaiduBaikeAPIKey 设置默认客户端的API Key
func SetBaiduBaikeAPIKey(apiKey string) {
	defaultBaikeClient.apiKey = apiKey
}

// ============== 请求构建函数 ==============

// NewDefaultBaikeRequest 创建默认请求参数
func NewDefaultBaikeRequest(searchKey string) *BaiduBaikeRequest {
	return &BaiduBaikeRequest{
		SearchType: DefaultSearchType,
		SearchKey:  searchKey,
		RequestId:  DefaultBaikeRequestID,
	}
}

// ============== HTTP 请求辅助函数 ==============

// buildHTTPRequest 构建HTTP GET请求
func (c *BaiduBaikeClient) buildHTTPRequest(req *BaiduBaikeRequest) (*http.Request, error) {
	// 构建URL参数
	params := url.Values{}
	params.Add("search_type", req.SearchType)
	params.Add("search_key", req.SearchKey)

	// 拼接完整URL
	fullURL := fmt.Sprintf("%s?%s", BaiduBaikeURL, params.Encode())

	// 创建HTTP GET请求
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")
	if req.RequestId != "" {
		httpReq.Header.Set("X-Appbuilder-Request-Id", req.RequestId)
	}

	return httpReq, nil
}

// executeRequest 执行HTTP请求
func (c *BaiduBaikeClient) executeRequest(httpReq *http.Request) ([]byte, error) {
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
func (c *BaiduBaikeClient) parseResponse(body []byte) (*BaiduBaikeResponse, error) {
	var response BaiduBaikeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// 如果解析失败，仍然返回原始响应
		response.RawResponse = string(body)
		return &response, fmt.Errorf("解析响应失败: %w", err)
	}
	response.RawResponse = string(body)

	return &response, nil
}

// ============== 核心调用函数 ==============

// Search 执行百度百科搜索
func (c *BaiduBaikeClient) Search(req *BaiduBaikeRequest) (*BaiduBaikeResponse, error) {
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
	log.Println("Response:", string(body))

	// 3. 解析响应
	return c.parseResponse(body)
}

// SearchWithKey 使用简化接口执行搜索
func (c *BaiduBaikeClient) SearchWithKey(searchKey string) (*BaiduBaikeResponse, error) {
	req := NewDefaultBaikeRequest(searchKey)
	return c.Search(req)
}

// ============== 外部调用接口 ==============

// BaiduBaikeSearch 供外部调用的百度百科搜索函数
// 参数:
//   - req: 百度百科请求参数
//
// 返回:
//   - 搜索响应和错误信息
func BaiduBaikeSearch(req *BaiduBaikeRequest) (*BaiduBaikeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 设置默认值
	if req.SearchType == "" {
		req.SearchType = DefaultSearchType
	}
	if req.RequestId == "" {
		req.RequestId = DefaultBaikeRequestID
	}

	return defaultBaikeClient.Search(req)
}

// BaiduBaikeSearchSimple 简化版百科搜索(使用默认配置)
// 参数:
//   - searchKey: 搜索关键词
//
// 返回:
//   - 搜索响应和错误信息
func BaiduBaikeSearchSimple(searchKey string) (*BaiduBaikeResponse, error) {
	req := NewDefaultBaikeRequest(searchKey)
	return defaultBaikeClient.Search(req)
}

// BaiduBaikeSearchByLemmaId 通过词条ID搜索
// 参数:
//   - lemmaId: 词条ID
//
// 返回:
//   - 搜索响应和错误信息
func BaiduBaikeSearchByLemmaId(lemmaId string) (*BaiduBaikeResponse, error) {
	req := &BaiduBaikeRequest{
		SearchType: "lemmaId",
		SearchKey:  lemmaId,
		RequestId:  DefaultBaikeRequestID,
	}
	return defaultBaikeClient.Search(req)
}
