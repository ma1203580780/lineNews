package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"lineNews/agent/logutil"
	"lineNews/agent/tool"
	"lineNews/model"
)

// ModeWorkflow 不同模式的工作流
type ModeWorkflow struct {
	llmCaller  *tool.LLMCaller
	arkModelID string
}

// NewModeWorkflow 创建模式工作流
func NewModeWorkflow(llmCaller *tool.LLMCaller, arkModelID string) *ModeWorkflow {
	return &ModeWorkflow{
		llmCaller:  llmCaller,
		arkModelID: arkModelID,
	}
}

// callArkModelAndUnmarshal 调用Ark模型并解析JSON响应
func (w *ModeWorkflow) callArkModelAndUnmarshal(ctx context.Context, systemPrompt, userPrompt, stage string, result interface{}) error {
	message := fmt.Sprintf("%s %s", systemPrompt, userPrompt)

	arkResponse, err := model.SendArkMessage(ctx, w.arkModelID, message, "")
	if err != nil {
		return fmt.Errorf("调用Ark模型失败: %w", err)
	}

	if err := json.Unmarshal([]byte(arkResponse.Content), result); err != nil {
		return fmt.Errorf("解析JSON失败: %w, 原始内容: %s", err, arkResponse.Content)
	}

	return nil
}

// ArkInputContent 输入内容结构
type ArkInputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ArkInput 输入结构
type ArkInput struct {
	Role    string            `json:"role"`
	Content []ArkInputContent `json:"content"`
}

// ModeEvent 模式工作流中的事件数据结构
type ModeEvent struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Time     string   `json:"time"`
	Location string   `json:"location"`
	People   []string `json:"people"`
	Summary  string   `json:"summary"`
}

// ModeTimelineResponse 模式工作流中的时间链响应结构
type ModeTimelineResponse struct {
	Keyword string      `json:"keyword"`
	Events  []ModeEvent `json:"events"`
}

// ArkRequestWithTools 带工具的请求结构
type ArkRequestWithTools struct {
	Model               string       `json:"model"`
	MaxCompletionTokens int          `json:"max_completion_tokens,omitempty"`
	ReasoningEffort     string       `json:"reasoning_effort,omitempty"`
	Stream              bool         `json:"stream,omitempty"`
	Tools               []model.Tool `json:"tools"`
	Input               []ArkInput   `json:"input"`
}

// callArkModelWithWebSearch 调用Ark模型并使用网络搜索工具来补充信息
func (w *ModeWorkflow) callArkModelWithWebSearch(ctx context.Context, systemPrompt, userPrompt string, result interface{}) error {
	// 构建带工具的请求
	requestBody := ArkRequestWithTools{
		Model: "doubao-seed-1-6-flash-250828", //w.arkModelID,
		// ReasoningEffort: "low",                          //控制思维链长度 [ 新增 ] https://www.volcengine.com/docs/82379/2123288?lang=zh
		Stream: false,
		Tools: []model.Tool{
			{
				Type: "web_search",
			},
		},
		Input: []ArkInput{
			{
				Role: "user",
				Content: []ArkInputContent{
					{
						Type: "input_text",
						Text: fmt.Sprintf("%s %s", systemPrompt, userPrompt),
					},
				},
			},
		},
	}

	// 序列化请求体
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://ark.cn-beijing.volces.com/api/v3/responses", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ARK_API_KEY 环境变量未设置")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	log.Println("API响应:", string(respBody))

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应 - API可能返回包装过的响应
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResponse); err != nil {
		return fmt.Errorf("解析API响应失败: %w, 原始内容: %s", err, string(respBody))
	}

	// 检查是否有错误信息
	if errorInfo, hasError := rawResponse["error"].(map[string]interface{}); hasError {
		return fmt.Errorf("API返回错误: %v", errorInfo)
	}

	// 尝试从响应中提取实际内容，可能在不同的字段中
	var contentStr string
	if content, ok := rawResponse["content"].(string); ok {
		contentStr = content
	} else if text, ok := rawResponse["text"].(string); ok {
		contentStr = text
	} else if choices, ok := rawResponse["choices"].([]interface{}); ok && len(choices) > 0 {
		// 兼容类似chat completions的格式
		if choiceMap, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choiceMap["message"].(map[string]interface{}); ok {
				if msgContent, ok := message["content"].(string); ok {
					contentStr = msgContent
				}
			}
		}
	} else if output, ok := rawResponse["output"].([]interface{}); ok && len(output) > 0 {
		// 处理新的API响应格式，其中数据在output数组中
		for _, item := range output {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemType, ok := itemMap["type"].(string); ok && itemType == "summary" {
					if summaries, ok := itemMap["summary"].([]interface{}); ok && len(summaries) > 0 {
						if summaryMap, ok := summaries[0].(map[string]interface{}); ok {
							if text, ok := summaryMap["text"].(string); ok {
								contentStr = text
								break // 找到后就退出循环
							}
						}
					}
				}
			}
		}
	}

	if contentStr == "" {
		// 如果没有找到特定字段，尝试将整个响应作为JSON字符串
		contentStr = string(respBody)
	}

	if err := json.Unmarshal([]byte(contentStr), result); err != nil {
		return fmt.Errorf("解析JSON内容失败: %w, 原始内容: %s", err, contentStr)
	}

	return nil
}

// GenerateMockMode 生成模拟模式数据
func (w *ModeWorkflow) GenerateMockMode(ctx context.Context, keyword string) (*ModeTimelineResponse, error) {
	logutil.LogInfo("开始执行Mock模式，关键词: %s", keyword)

	// 返回模拟的时间链数据
	timeline := &ModeTimelineResponse{
		Keyword: keyword,
		Events: []ModeEvent{
			{
				ID:       "1",
				Title:    fmt.Sprintf("%s 相关新闻一：事件起源", keyword),
				Time:     "2023-01-10",
				Location: "北京",
				People:   []string{"张三", "李四"},
				Summary:  fmt.Sprintf("围绕 %s 的最初报道和背景信息。", keyword),
			},
			{
				ID:       "2",
				Title:    fmt.Sprintf("%s 相关新闻二：事态发展", keyword),
				Time:     "2023-03-05",
				Location: "上海",
				People:   []string{"王五"},
				Summary:  fmt.Sprintf("%s 相关事件在区域内的进一步发酵与反应。", keyword),
			},
			{
				ID:       "3",
				Title:    fmt.Sprintf("%s 相关新闻三：官方回应", keyword),
				Time:     "2023-05-20",
				Location: "广州",
				People:   []string{"官方发言人"},
				Summary:  fmt.Sprintf("有关部门针对 %s 发布官方说明与政策。", keyword),
			},
		},
	}

	logutil.LogInfo("Mock模式完成，包含 %d 个事件", len(timeline.Events))
	return timeline, nil
}
