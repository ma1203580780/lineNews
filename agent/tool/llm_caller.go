package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"lineNews/agent/logutil"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// LLMCaller LLM调用器
type LLMCaller struct {
	chatModel *deepseek.ChatModel
}

// NewLLMCaller 创建LLM调用器
func NewLLMCaller(chatModel *deepseek.ChatModel) *LLMCaller {
	return &LLMCaller{
		chatModel: chatModel,
	}
}

// CallWithPrompt 使用系统提示词和用户提示词调用LLM
func (c *LLMCaller) CallWithPrompt(ctx context.Context, systemPrompt, userPrompt, stage string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 打印提示词，便于调试观察
	logutil.LogInfo("[LLMCaller] %s阶段 System Prompt:\n%s", stage, systemPrompt)
	logutil.LogInfo("[LLMCaller] %s阶段 User Prompt:\n%s", stage, userPrompt)
	logutil.LogInfo("[LLMCaller] 正在调用LLM: %s", stage)

	response, err := c.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM调用失败: %w", err)
	}

	logutil.LogInfo("[LLMCaller] %s阶段 AI 响应: %s", stage, response.Content)
	return response.Content, nil
}

// CallAndUnmarshal 调用LLM并解析JSON响应
func (c *LLMCaller) CallAndUnmarshal(ctx context.Context, systemPrompt, userPrompt, stage string, result interface{}) error {
	content, err := c.CallWithPrompt(ctx, systemPrompt, userPrompt, stage)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(content), result); err != nil {
		return fmt.Errorf("解析JSON失败: %w, 原始内容: %s", err, content)
	}

	return nil
}
