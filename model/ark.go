package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lineNews/agent/logutil"
	"net/http"
	"os"
)

// ==================== 常量定义 ====================

const (
	DefaultArkModel = "doubao-seed-1-6-251015"
	ArkFlashModel   = "doubao-seed-1-6-flash-250828"
)

// ==================== 数据结构 ====================

// ArkChatResponse 聊天响应
type ArkChatResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Tool 定义工具结构
type Tool struct {
	Type string `json:"type"`
}

// ArkInputContent 输入内容项
type ArkInputContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ArkInput 输入结构
type ArkInput struct {
	Role    string            `json:"role"`
	Content []ArkInputContent `json:"content"`
}

// ArkRequestWithTools 带工具的请求结构
type ArkRequestWithTools struct {
	Model           string     `json:"model"`
	Stream          bool       `json:"stream"`
	Tools           []Tool     `json:"tools,omitempty"`
	Input           []ArkInput `json:"input"`
	ReasoningEffort string     `json:"reasoning_effort,omitempty"`
}

// MessageModel 消息结构
type MessageModel struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ContentItemModel 消息内容项
type ContentItemModel struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageUrl ImageURLModel `json:"image_url,omitempty"`
}

// ImageURLModel 图像URL
type ImageURLModel struct {
	URL string `json:"url"`
}

// ArkResponseModel API响应结构
type ArkResponseModel struct {
	CreatedAt       int64        `json:"created_at"`
	ID              string       `json:"id"`
	MaxOutputTokens int          `json:"max_output_tokens"`
	Model           string       `json:"model"`
	Object          string       `json:"object"`
	Output          []OutputItem `json:"output"`
	ServiceTier     string       `json:"service_tier"`
	Status          string       `json:"status"`
	Tools           []Tool       `json:"tools"`
	Usage           UsageModel   `json:"usage"`
	Caching         Caching      `json:"caching"`
	Store           bool         `json:"store"`
	ExpireAt        int64        `json:"expire_at"`
}

// OutputItem 输出项
type OutputItem struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Summary []Summary     `json:"summary,omitempty"`
	Action  SearchAction  `json:"action,omitempty"`
	Content []ContentItem `json:"content,omitempty"`
	Status  string        `json:"status"`
	Role    string        `json:"role,omitempty"`
}

// Summary 摘要项
type Summary struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SearchAction 搜索动作
type SearchAction struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

// ContentItem 内容项
type ContentItem struct {
	Type        string       `json:"type"`
	Text        string       `json:"text,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// Annotation 注释
type Annotation struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	LogoURL     string    `json:"logo_url"`
	SiteName    string    `json:"site_name"`
	PublishTime string    `json:"publish_time"`
	Summary     string    `json:"summary"`
	CoverImage  ImageInfo `json:"cover_image,omitempty"`
}

// ImageInfo 图像信息
type ImageInfo struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Caching 缓存信息
type Caching struct {
	Type string `json:"type"`
}

// ChoiceModel 选择项
type ChoiceModel struct {
	Index        int          `json:"index"`
	Message      MessageModel `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

// UsageModel 使用情况
type UsageModel struct {
	PromptTokens        int                 `json:"prompt_tokens"`
	CompletionTokens    int                 `json:"completion_tokens"`
	TotalTokens         int                 `json:"total_tokens"`
	InputTokensDetails  InputTokensDetails  `json:"input_tokens_details"`
	OutputTokensDetails OutputTokensDetails `json:"output_tokens_details"`
	ToolUsage           map[string]int      `json:"tool_usage"`
	ToolUsageDetails    ToolUsageDetails    `json:"tool_usage_details"`
}

// InputTokensDetails 输入Token详情
type InputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// OutputTokensDetails 输出Token详情
type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// ToolUsageDetails 工具使用详情
type ToolUsageDetails struct {
	WebSearch WebSearchUsage `json:"web_search"`
}

// WebSearchUsage 联网搜索使用详情
type WebSearchUsage struct {
	SearchEngine int `json:"search_engine"`
}

// ==================== 核心操作 ====================

// CreateArkChatModel 创建 Ark ChatModel - 这个函数现在只是返回一个标识，实际调用在发送消息时进行
func CreateArkChatModel(ctx context.Context) (string, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ARK_API_KEY") // 兼容旧环境变量名
		if apiKey == "" {
			return "", fmt.Errorf("ARK_API_KEY 或 ARK_API_KEY 环境变量未设置")
		}
	}

	model := os.Getenv("ARK_MODEL_ID")
	if model == "" {
		model = os.Getenv("ARK_MODEL_ID") // 兼容旧环境变量名
		if model == "" {
			model = DefaultArkModel
		}
	}

	// 验证配置是否有效
	if apiKey != "" && model != "" {
		return model, nil
	}

	return "", fmt.Errorf("Ark 配置无效")
}

// SendArkMessage 发送单条消息到Ark
func SendArkMessage(ctx context.Context, modelID string, userMessage string, systemMessage string) (*ArkChatResponse, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY 环境变量未设置")
	}

	if modelID == "" {
		modelID = DefaultArkModel
	}

	// 构建请求体
	requestBody := ArkRequestWithTools{
		Model:  modelID,
		Stream: false,
		Tools:  []Tool{{Type: "web_search"}},
		Input: []ArkInput{
			{
				Role: "user",
				Content: []ArkInputContent{
					{
						Type: "input_text",
						Text: userMessage,
					},
				},
			},
		},
	}

	// 如果有系统消息，需要特殊处理（Ark API 不直接支持系统消息）
	if systemMessage != "" {
		// 将系统消息作为单独的输入项添加
		requestBody.Input = append([]ArkInput{{
			Role: "system",
			Content: []ArkInputContent{{
				Type: "input_text",
				Text: systemMessage,
			}},
		}}, requestBody.Input...)
	}

	// 序列化请求体
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://ark.cn-beijing.volces.com/api/v3/responses", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应 - 根据实际的Ark API响应结构
	var apiResp ArkResponseModel
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取响应内容 - 根据实际的API响应结构，查找type为message的output项
	responseText := ""
	for _, output := range apiResp.Output {
		if output.Type == "message" && output.Role == "assistant" {
			// 在content数组中查找text类型的内容
			for _, content := range output.Content {
				if content.Type == "output_text" && content.Text != "" {
					responseText = content.Text
					break
				}
			}
			if responseText != "" {
				break // 找到内容后退出外层循环
			}
		}
	}

	if responseText == "" {
		logutil.LogInfo("在API响应中未找到有效的输出文本，响应结构: %+v", apiResp)
		// 尝试从摘要中获取内容
		for _, output := range apiResp.Output {
			if output.Type == "reasoning" {
				for _, summary := range output.Summary {
					if summary.Type == "summary_text" && summary.Text != "" {
						responseText = summary.Text
						break
					}
				}
			}
			if responseText != "" {
				break
			}
		}
	}

	// 返回结果
	result := &ArkChatResponse{
		Content:          responseText,
		PromptTokens:     apiResp.Usage.PromptTokens,
		CompletionTokens: apiResp.Usage.CompletionTokens,
		TotalTokens:      apiResp.Usage.TotalTokens,
	}

	return result, nil
}

// ==================== 工具函数 ====================

// LogArkChatResponse 打印响应信息
func LogArkChatResponse(response *ArkChatResponse) {
	if response == nil {
		logutil.LogInfo("响应为空")
		return
	}

	// 使用json.Marshal来格式化输出
	respJSON, _ := json.Marshal(response)
	logutil.LogInfo("Ark 响应: %s", string(respJSON))

	if response.TotalTokens > 0 {
		logutil.LogInfo("Token 使用统计:\n  输入 Token: %d\n  输出 Token: %d\n  总计 Token: %d", response.PromptTokens, response.CompletionTokens, response.TotalTokens)
	}
}
