package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lineNews/agent/prompt"
	"lineNews/agent/tool"
)

// StreamTimelineWorkflow 支持流式输出的时间链生成工作流
type StreamTimelineWorkflow struct {
	streamLLMCaller *tool.StreamLLMCaller
}

// NewStreamTimelineWorkflow 创建支持流式输出的时间链工作流
func NewStreamTimelineWorkflow(streamLLMCaller *tool.StreamLLMCaller) *StreamTimelineWorkflow {
	return &StreamTimelineWorkflow{
		streamLLMCaller: streamLLMCaller,
	}
}

// GenerateStream 生成时间链并流式输出思考过程
func (w *StreamTimelineWorkflow) GenerateStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	// 发送开始事件
	startEvent := tool.StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("开始生成关键词「%s」的时间链", keyword),
		Stage:   "总览",
	}
	if err := sendEvent(startEvent); err != nil {
		return nil, err
	}

	// 第一步：初次生成
	timeline, err := w.generateInitialStream(ctx, keyword, sendEvent)
	if err != nil {
		errorEvent := tool.StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("初次生成失败: %v", err),
			Stage:   "时间链初次生成",
		}
		sendEvent(errorEvent)
		return nil, err
	}

	// 发送中间结果事件
	midResultEvent := tool.StreamEvent{
		Type:    "thinking",
		Content: fmt.Sprintf("初次生成完成，包含 %d 个事件，开始反思优化...", len(timeline.Events)),
		Stage:   "时间链初次生成",
	}
	if err := sendEvent(midResultEvent); err != nil {
		return nil, err
	}

	// 第二步：反思优化（最多3轮）
	const maxRefineRounds = 3
	currentTimeline := timeline
	for i := 0; i < maxRefineRounds; i++ {
		log.Printf("[StreamTimelineWorkflow] 第 %d 轮反思优化开始，当前事件数: %d", i+1, len(currentTimeline.Events))

		refinedTimeline, err := w.refineStream(ctx, keyword, currentTimeline, sendEvent, i+1)
		if err != nil {
			log.Printf("[StreamTimelineWorkflow] 第 %d 轮反思优化失败: %v", i+1, err)
			// 发送警告但继续
			warningEvent := tool.StreamEvent{
				Type:    "thinking",
				Content: fmt.Sprintf("第 %d 轮反思优化失败，使用上一版本结果: %v", i+1, err),
				Stage:   fmt.Sprintf("反思优化第%d轮", i+1),
			}
			sendEvent(warningEvent)
			break
		}
		if refinedTimeline == nil || len(refinedTimeline.Events) == 0 {
			log.Printf("[StreamTimelineWorkflow] 第 %d 轮反思优化返回空结果，停止进一步反思", i+1)
			// 发送警告但继续
			warningEvent := tool.StreamEvent{
				Type:    "thinking",
				Content: fmt.Sprintf("第 %d 轮反思优化返回空结果，停止进一步反思", i+1),
				Stage:   fmt.Sprintf("反思优化第%d轮", i+1),
			}
			sendEvent(warningEvent)
			break
		}

		log.Printf("[StreamTimelineWorkflow] 第 %d 轮反思优化后事件数: %d", i+1, len(refinedTimeline.Events))
		currentTimeline = refinedTimeline

		// 发送优化后结果
		refineResultEvent := tool.StreamEvent{
			Type:    "thinking",
			Content: fmt.Sprintf("第 %d 轮反思优化完成，当前事件数: %d", i+1, len(currentTimeline.Events)),
			Stage:   fmt.Sprintf("反思优化第%d轮", i+1),
		}
		if err := sendEvent(refineResultEvent); err != nil {
			return nil, err
		}

		// 如果事件数量已经在理想范围内，则提前结束循环
		if len(currentTimeline.Events) >= 15 && len(currentTimeline.Events) <= 100 {
			log.Printf("[StreamTimelineWorkflow] 反思优化后事件数量已满足要求（%d 条），结束反思循环", len(currentTimeline.Events))
			endOptEvent := tool.StreamEvent{
				Type:    "thinking",
				Content: fmt.Sprintf("反思优化后事件数量已满足要求（%d 条），结束反思循环", len(currentTimeline.Events)),
				Stage:   fmt.Sprintf("反思优化第%d轮", i+1),
			}
			if err := sendEvent(endOptEvent); err != nil {
				return nil, err
			}
			break
		}
	}

	log.Printf("[StreamTimelineWorkflow] 最终时间链生成完成，包含 %d 个事件", len(currentTimeline.Events))

	// 发送最终结果
	finalEvent := tool.StreamEvent{
		Type:    "result",
		Content: currentTimeline,
		Stage:   "总览",
	}
	if err := sendEvent(finalEvent); err != nil {
		return nil, err
	}

	// 发送完成事件
	doneEvent := tool.StreamEvent{
		Type:    "done",
		Content: "时间链生成完成",
		Stage:   "总览",
	}
	if err := sendEvent(doneEvent); err != nil {
		return nil, err
	}

	return currentTimeline, nil
}

// generateInitialStream 初次生成时间链（流式版本）
func (w *StreamTimelineWorkflow) generateInitialStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	userPrompt := fmt.Sprintf("请为关键词「%s」生成新闻时间链", keyword)

	var timeline TimelineResponse
	err := w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"时间链初次生成",
		&timeline,
		sendEvent,
	)
	if err != nil {
		return nil, fmt.Errorf("生成时间链失败: %w", err)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = keyword
	}

	log.Printf("[StreamTimelineWorkflow] 初次生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// refineStream 反思优化时间链（流式版本）
func (w *StreamTimelineWorkflow) refineStream(
	ctx context.Context,
	keyword string,
	original *TimelineResponse,
	sendEvent func(tool.StreamEvent) error,
	round int,
) (*TimelineResponse, error) {
	if original == nil {
		return nil, fmt.Errorf("原始时间链为空")
	}

	// 将原始时间链转换为 JSON 字符串
	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("序列化原始时间链失败: %w", err)
	}

	userPrompt := fmt.Sprintf(
		"下面是模型第一次为关键词「%s」生成的时间链 JSON：\n%s\n\n请在内部使用 ReAct 模式进行反思和推理，检查事件数量和内容是否满足上述要求，并在此基础上进行补充、合并和优化，生成最终的高质量时间链。请直接返回最终的 JSON，不要输出任何解释性文字。",
		keyword,
		string(originalJSON),
	)

	var refined TimelineResponse
	err = w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		prompt.TimelineRefinementSystemPrompt,
		userPrompt,
		fmt.Sprintf("时间链反思优化第%d轮", round),
		&refined,
		sendEvent,
	)
	if err != nil {
		return nil, fmt.Errorf("反思优化时间链失败: %w", err)
	}

	if refined.Keyword == "" {
		refined.Keyword = keyword
	}

	// 再次做数量上的兜底校验，如果仍然远少于 15 条，则保留原结果
	if len(refined.Events) < 5 {
		log.Printf("[StreamTimelineWorkflow] 反思后事件数过少(%d)，保留原始时间链(%d)", len(refined.Events), len(original.Events))
		return original, nil
	}

	return &refined, nil
}
