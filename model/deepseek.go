package model

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// ==================== 常量定义 ====================

const (
	DefaultDeepSeekAPIKey = ""
	DefaultDeepSeekModel  = "deepseek-chat"
	DefaultDeepSeekURL    = "https://api.deepseek.com"
)

// ==================== 数据结构 ====================

// DSModelConfig DeepSeek 模型配置
type DSModelConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// DSChatResponse 聊天响应
type DSChatResponse struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ==================== 工厂函数 ====================

// NewDSModelConfig 创建默认配置
func NewDSModelConfig() *DSModelConfig {
	// 从环境变量获取配置，如果环境变量未设置则使用默认值
	config := loadConfig()
	return &DSModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
	}
}

// ==================== 核心操作 ====================

// CreateDSChatModel 创建 DeepSeek ChatModel
// loadConfig 从环境变量加载DeepSeek配置
func loadConfig() *DSModelConfig {
	return &DSModelConfig{
		APIKey:  getEnv("DEEPSEEK_API_KEY", DefaultDeepSeekAPIKey),
		Model:   getEnv("DEEPSEEK_MODEL", DefaultDeepSeekModel),
		BaseURL: getEnv("DEEPSEEK_BASE_URL", DefaultDeepSeekURL),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func CreateDSChatModel(ctx context.Context, config *DSModelConfig) (*deepseek.ChatModel, error) {
	if config == nil {
		config = NewDSModelConfig()
	}

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  config.APIKey,
		Model:   config.Model,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 DeepSeek ChatModel 失败: %w", err)
	}

	return chatModel, nil
}

// SendMessage 发送单条消息到 DeepSeek
func SendMessage(ctx context.Context, chatModel *deepseek.ChatModel, userMessage string, systemMessage string) (*DSChatResponse, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("ChatModel 为空")
	}

	// 构建消息列表
	messages := []*schema.Message{
		schema.SystemMessage(systemMessage),
		schema.UserMessage(userMessage),
	}

	// 调用模型
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成响应失败: %w", err)
	}

	// 提取结果
	result := &DSChatResponse{
		Content: response.Content,
	}

	// 获取 Token 使用统计
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		result.PromptTokens = response.ResponseMeta.Usage.PromptTokens
		result.CompletionTokens = response.ResponseMeta.Usage.CompletionTokens
		result.TotalTokens = response.ResponseMeta.Usage.TotalTokens
	}

	return result, nil
}

// SendMessageWithHistory 发送消息（支持对话历史）
func SendMessageWithHistory(ctx context.Context, chatModel *deepseek.ChatModel, messages []*schema.Message) (*DSChatResponse, error) {
	if chatModel == nil {
		return nil, fmt.Errorf("ChatModel 为空")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("消息列表为空")
	}

	// 调用模型
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成响应失败: %w", err)
	}

	// 提取结果
	result := &DSChatResponse{
		Content: response.Content,
	}

	// 获取 Token 使用统计
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		result.PromptTokens = response.ResponseMeta.Usage.PromptTokens
		result.CompletionTokens = response.ResponseMeta.Usage.CompletionTokens
		result.TotalTokens = response.ResponseMeta.Usage.TotalTokens
	}

	return result, nil
}

// ==================== 工具函数 ====================

// LogDSChatResponse 打印响应信息
func LogDSChatResponse(response *DSChatResponse) {
	if response == nil {
		log.Println("响应为空")
		return
	}

	fmt.Printf("AI 响应: %s\n", response.Content)
	if response.TotalTokens > 0 {
		fmt.Printf("\nToken 使用统计:\n")
		fmt.Printf("  输入 Token: %d\n", response.PromptTokens)
		fmt.Printf("  输出 Token: %d\n", response.CompletionTokens)
		fmt.Printf("  总计 Token: %d\n", response.TotalTokens)
	}
}
