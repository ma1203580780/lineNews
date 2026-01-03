package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ==================== 常量定义 ====================

const (
	DefaultArkModel = "doubao-seed-1-6-251015"
)

// ==================== 数据结构 ====================

// ArkChatResponse 聊天响应
type ArkChatResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ArkRequestModel API请求结构
type ArkRequestModel struct {
	Model               string         `json:"model"`
	MaxCompletionTokens int            `json:"max_completion_tokens,omitempty"`
	Messages            []MessageModel `json:"messages"`
	ReasoningEffort     string         `json:"reasoning_effort,omitempty"`
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
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChoiceModel `json:"choices"`
	Usage   UsageModel    `json:"usage"`
}

// ChoiceModel 选择项
type ChoiceModel struct {
	Index        int          `json:"index"`
	Message      MessageModel `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

// UsageModel 使用情况
type UsageModel struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
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
	requestBody := ArkRequestModel{
		Model:               modelID,
		MaxCompletionTokens: 65535,
		ReasoningEffort:     "medium",
		Messages: []MessageModel{
			{
				Role: "user",
				Content: []ContentItemModel{
					{
						Type: "text",
						Text: userMessage,
					},
				},
			},
		},
	}

	// 如果有系统消息，需要特殊处理（Ark API 不直接支持系统消息）
	if systemMessage != "" {
		// 将系统消息合并到用户消息中
		fullMessage := systemMessage + " " + userMessage
		requestBody.Messages[0].Content = []ContentItemModel{
			{
				Type: "text",
				Text: fullMessage,
			},
		}
	}

	// 序列化请求体
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://ark.cn-beijing.volces.com/api/v3/chat/completions", bytes.NewBuffer(reqBody))
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

	// 解析响应
	var apiResp ArkResponseModel
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取响应内容
	responseText := ""
	if len(apiResp.Choices) > 0 {
		choice := apiResp.Choices[0]
		// 处理响应内容
		if contentStr, ok := choice.Message.Content.(string); ok {
			responseText = contentStr
		} else if contentItems, ok := choice.Message.Content.([]interface{}); ok {
			// 如果内容是数组，提取文本部分
			for _, item := range contentItems {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemMap["type"] == "text" {
						if text, exists := itemMap["text"].(string); exists {
							responseText = text
							break
						}
					}
				}
			}
		} else if contentMap, ok := choice.Message.Content.(map[string]interface{}); ok {
			// 如果内容是单个对象，检查文本字段
			if text, exists := contentMap["text"].(string); exists {
				responseText = text
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

// SendArkMessageWithHistory 发送消息（支持对话历史）
func SendArkMessageWithHistory(ctx context.Context, modelID string, messages []MessageModel) (*ArkChatResponse, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY 环境变量未设置")
	}

	if modelID == "" {
		modelID = DefaultArkModel
	}

	// 构建请求体
	requestBody := ArkRequestModel{
		Model:               modelID,
		MaxCompletionTokens: 65535,
		ReasoningEffort:     "medium",
		Messages:            messages,
	}

	// 序列化请求体
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://ark.cn-beijing.volces.com/api/v3/chat/completions", bytes.NewBuffer(reqBody))
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

	// 解析响应
	var apiResp ArkResponseModel
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取响应内容
	responseText := ""
	if len(apiResp.Choices) > 0 {
		choice := apiResp.Choices[0]
		// 处理响应内容
		if contentStr, ok := choice.Message.Content.(string); ok {
			responseText = contentStr
		} else if contentItems, ok := choice.Message.Content.([]interface{}); ok {
			// 如果内容是数组，提取文本部分
			for _, item := range contentItems {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemMap["type"] == "text" {
						if text, exists := itemMap["text"].(string); exists {
							responseText = text
							break
						}
					}
				}
			}
		} else if contentMap, ok := choice.Message.Content.(map[string]interface{}); ok {
			// 如果内容是单个对象，检查文本字段
			if text, exists := contentMap["text"].(string); exists {
				responseText = text
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

// StreamArkMessage 流式发送单条消息到Ark - 由于HTTP方式的限制，这里返回错误
func StreamArkMessage(ctx context.Context, modelID string, userMessage string, systemMessage string) (<-chan *ArkChatResponse, <-chan error) {
	responseChan := make(chan *ArkChatResponse)
	errorChan := make(chan error)

	// 在 goroutine 中返回错误，因为HTTP方式不支持流式响应
	go func() {
		defer close(responseChan)
		defer close(errorChan)

		errorChan <- fmt.Errorf("HTTP方式不支持流式响应，请使用SSE流式接口")
	}()

	return responseChan, errorChan
}

// ==================== 工具函数 ====================

// LogArkChatResponse 打印响应信息
func LogArkChatResponse(response *ArkChatResponse) {
	if response == nil {
		log.Println("响应为空")
		return
	}

	// 使用json.Marshal来格式化输出
	respJSON, _ := json.Marshal(response)
	log.Printf("Ark 响应: %s", string(respJSON))

	if response.TotalTokens > 0 {
		fmt.Printf("\nToken 使用统计:\n")
		fmt.Printf("  输入 Token: %d\n", response.PromptTokens)
		fmt.Printf("  输出 Token: %d\n", response.CompletionTokens)
		fmt.Printf("  总计 Token: %d\n", response.TotalTokens)
	}
}
