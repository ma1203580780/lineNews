package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"lineNews/agent/logutil"

	"github.com/gin-gonic/gin"
)

// ArkRequest API请求结构
type ArkRequest struct {
	Model               string    `json:"model"`
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"`
	Messages            []Message `json:"messages"`
	ReasoningEffort     string    `json:"reasoning_effort,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// ContentItem 消息内容项
type ContentItem struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	ImageUrl ImageURL `json:"image_url,omitempty"`
}

// ImageURL 图像URL
type ImageURL struct {
	URL string `json:"url"`
}

// ArkResponse API响应结构
type ArkResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择项
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// HandleArkChat 处理 Ark Chat 请求
func HandleArkChat(c *gin.Context) {
	message := c.Query("message")
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message 参数不能为空",
		})
		return
	}

	logutil.LogInfo("Ark Chat 请求: %s", message)

	// 从环境变量获取配置
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ARK_API_KEY 环境变量未设置",
		})
		return
	}

	var arkFlashModel = "doubao-seed-1-6-flash-250828" //Doubao-Seed-1.6-flash推理速度极致的多模态深度思考模型，TPOT低至10ms； 同时支持文本和视觉理解，文本理解能力超过上一代lite，视觉理解比肩友商pro系列模型。支持 256k 上下文窗口，输出长度支持最大 16k tokens。

	// 构建请求体 - 纯文本消息
	requestBody := ArkRequest{
		Model:               arkFlashModel,
		MaxCompletionTokens: 65535,
		// ReasoningEffort:     "medium",
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentItem{
					{
						Type: "text",
						Text: message,
					},
				},
			},
		},
	}

	// 序列化请求体
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		logutil.LogError("序列化请求体失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("序列化请求体失败: %v", err),
		})
		return
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://ark.cn-beijing.volces.com/api/v3/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		logutil.LogError("创建HTTP请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("创建HTTP请求失败: %v", err),
		})
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logutil.LogError("发送请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("发送请求失败: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logutil.LogError("读取响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("读取响应失败: %v", err),
		})
		return
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		logutil.LogError("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		c.JSON(resp.StatusCode, gin.H{
			"error":   fmt.Sprintf("API请求失败，状态码: %d", resp.StatusCode),
			"details": string(respBody),
		})
		return
	}

	// 解析响应
	var apiResp ArkResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logutil.LogError("解析响应失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("解析响应失败: %v", err),
		})
		return
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

	// 构建响应
	response := gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"response": responseText,
			"model":    arkFlashModel,
			"usage": gin.H{
				"prompt_tokens":     apiResp.Usage.PromptTokens,
				"completion_tokens": apiResp.Usage.CompletionTokens,
				"total_tokens":      apiResp.Usage.TotalTokens,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
