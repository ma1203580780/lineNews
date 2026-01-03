package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"lineNews/agent/logutil"
	"lineNews/agent/prompt"
	"lineNews/agent/tool"
)

// TimelineResponse 时间链响应结构（从types.go复制）
type TimelineResponse struct {
	Keyword string  `json:"keyword"`
	Events  []Event `json:"events"`
}

// Event 事件数据结构（从types.go复制）
type Event struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Time     string   `json:"time"`
	Location string   `json:"location"`
	People   []string `json:"people"`
	Summary  string   `json:"summary"`
}

// TimelineWorkflow 时间链生成工作流
type TimelineWorkflow struct {
	llmCaller *tool.LLMCaller
}

// NewTimelineWorkflow 创建时间链工作流
func NewTimelineWorkflow(llmCaller *tool.LLMCaller) *TimelineWorkflow {
	return &TimelineWorkflow{
		llmCaller: llmCaller,
	}
}

// Generate 生成时间链（包含初次生成和反思优化）
func (w *TimelineWorkflow) Generate(ctx context.Context, keyword string) (*TimelineResponse, error) {
	// 第一步：初次生成
	timeline, err := w.generateInitial(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 第二步：反思优化（最多3轮）
	const maxRefineRounds = 3
	for i := 0; i < maxRefineRounds; i++ {
		logutil.LogInfo("第 %d 轮反思优化开始，当前事件数: %d", i+1, len(timeline.Events))

		refinedTimeline, err := w.refine(ctx, keyword, timeline)
		if err != nil {
			logutil.LogError("第 %d 轮反思优化失败: %v", i+1, err)
			break
		}
		if refinedTimeline == nil || len(refinedTimeline.Events) == 0 {
			logutil.LogInfo("第 %d 轮反思优化返回空结果，停止进一步反思", i+1)
			break
		}

		logutil.LogInfo("第 %d 轮反思优化后事件数: %d", i+1, len(refinedTimeline.Events))
		timeline = refinedTimeline

		// 如果事件数量已经在理想范围内，则提前结束循环
		if len(timeline.Events) >= 15 && len(timeline.Events) <= 100 {
			logutil.LogInfo("反思优化后事件数量已满足要求（%d 条），结束反思循环", len(timeline.Events))
			break
		}
	}

	logutil.LogInfo("最终时间链生成完成，包含 %d 个事件", len(timeline.Events))
	return timeline, nil
}

// generateInitial 初次生成时间链
func (w *TimelineWorkflow) generateInitial(ctx context.Context, keyword string) (*TimelineResponse, error) {
	userPrompt := fmt.Sprintf("请为关键词「%s」生成新闻时间链", keyword)

	var timeline TimelineResponse
	err := w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"时间链初次生成",
		&timeline,
	)
	if err != nil {
		return nil, fmt.Errorf("生成时间链失败: %w", err)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = keyword
	}

	logutil.LogInfo("初次生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// refine 反思优化时间链
func (w *TimelineWorkflow) refine(ctx context.Context, keyword string, original *TimelineResponse) (*TimelineResponse, error) {
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
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.TimelineRefinementSystemPrompt,
		userPrompt,
		"时间链反思优化",
		&refined,
	)
	if err != nil {
		return nil, fmt.Errorf("反思优化时间链失败: %w", err)
	}

	if refined.Keyword == "" {
		refined.Keyword = keyword
	}

	// 再次做数量上的兜底校验，如果仍然远少于 15 条，则保留原结果
	if len(refined.Events) < 5 {
		logutil.LogInfo("反思后事件数过少(%d)，保留原始时间链(%d)", len(refined.Events), len(original.Events))
		return original, nil
	}

	return &refined, nil
}
