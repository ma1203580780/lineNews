package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// StreamEvent 流式事件结构
type StreamEvent struct {
	Type    string      `json:"type"`            // "thinking", "result", "error", "done"
	Content interface{} `json:"content"`         // 根据类型包含不同内容
	Stage   string      `json:"stage,omitempty"` // 当前处理阶段
}

// StreamLLMCaller 支持流式输出的LLM调用器
type StreamLLMCaller struct {
	chatModel *deepseek.ChatModel
}

// NewStreamLLMCaller 创建支持流式输出的LLM调用器
func NewStreamLLMCaller(chatModel *deepseek.ChatModel) *StreamLLMCaller {
	return &StreamLLMCaller{
		chatModel: chatModel,
	}
}

// CallWithPromptStream 使用系统提示词和用户提示词调用LLM并流式输出思考过程
func (c *StreamLLMCaller) CallWithPromptStream(
	ctx context.Context,
	systemPrompt, userPrompt, stage string,
	sendEvent func(StreamEvent) error,
) error {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 发送思考过程
	thinkEvent := StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("正在处理 %s 阶段...", stage),
		Stage:   stage,
	}
	if err := sendEvent(thinkEvent); err != nil {
		return err
	}

	// 发送提示词信息
	promptEvent := StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("System Prompt: %s", systemPrompt),
		Stage:   stage,
	}
	if err := sendEvent(promptEvent); err != nil {
		return err
	}

	promptEvent2 := StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("User Prompt: %s", userPrompt),
		Stage:   stage,
	}
	if err := sendEvent(promptEvent2); err != nil {
		return err
	}

	log.Printf("[StreamLLMCaller] %s阶段 System Prompt:\n%s", stage, systemPrompt)
	log.Printf("[StreamLLMCaller] %s阶段 User Prompt:\n%s", stage, userPrompt)
	log.Printf("[StreamLLMCaller] 正在调用LLM: %s", stage)

	// 调用模型
	response, err := c.chatModel.Generate(ctx, messages)
	if err != nil {
		errorEvent := StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("LLM调用失败: %v", err),
			Stage:   stage,
		}
		sendEvent(errorEvent) // 尝试发送错误事件，即使失败也不影响错误返回
		return fmt.Errorf("LLM调用失败: %w", err)
	}

	log.Printf("[StreamLLMCaller] %s阶段 AI 响应: %s", stage, response.Content)

	// 发送结果
	resultEvent := StreamEvent{
		Type:    "result",
		Content: response.Content,
		Stage:   stage,
	}
	if err := sendEvent(resultEvent); err != nil {
		return err
	}

	return nil
}

// CallAndUnmarshalStream 调用LLM并解析JSON响应，同时流式输出思考过程
func (c *StreamLLMCaller) CallAndUnmarshalStream(
	ctx context.Context,
	systemPrompt, userPrompt, stage string,
	result interface{},
	sendEvent func(StreamEvent) error,
) error {
	// 发送开始事件
	startEvent := StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("开始 %s 阶段", stage),
		Stage:   stage,
	}
	if err := sendEvent(startEvent); err != nil {
		return err
	}

	// 先收集响应内容
	var responseContent string
	err := c.CallWithPromptStream(ctx, systemPrompt, userPrompt, stage, func(event StreamEvent) error {
		// 如果是结果类型，保存内容
		if event.Type == "result" {
			if contentStr, ok := event.Content.(string); ok {
				responseContent = contentStr
			}
		}
		// 将事件转发给原始的sendEvent函数
		return sendEvent(event)
	})
	if err != nil {
		return err
	}

	// 发送解析开始事件
	parseEvent := StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("正在解析 %s 阶段的响应...", stage),
		Stage:   stage,
	}
	if err := sendEvent(parseEvent); err != nil {
		return err
	}

	// 尝试解析JSON
	if err := json.Unmarshal([]byte(responseContent), result); err != nil {
		errorEvent := StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("解析JSON失败: %v, 原始内容: %s", err, responseContent),
			Stage:   stage,
		}
		sendEvent(errorEvent) // 尝试发送错误事件
		return fmt.Errorf("解析JSON失败: %w, 原始内容: %s", err, responseContent)
	}

	// 发送解析完成事件
	doneEvent := StreamEvent{
		Type:    "done",
		Content: fmt.Sprintf("%s 阶段完成", stage),
		Stage:   stage,
	}
	if err := sendEvent(doneEvent); err != nil {
		return err
	}

	return nil
}
