package agent

import (
	"context"
	"fmt"
	"os"

	"lineNews/agent/tool"
	"lineNews/agent/workflow"
	"lineNews/config"
	"lineNews/model"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
)

// NewsTimelineAgent 新闻时间链 Agent
type NewsTimelineAgent struct {
	chatModel        *deepseek.ChatModel
	apiKey           string
	timelineWorkflow *workflow.TimelineWorkflow
	graphWorkflow    *workflow.GraphWorkflow
	modeWorkflow     *workflow.ModeWorkflow
}

// NewNewsTimelineAgent 创建新闻时间链 Agent
func NewNewsTimelineAgent(ctx context.Context, cfg *config.Config) (*NewsTimelineAgent, error) {
	if cfg.DeepSeekAPIKey == "" {
		return nil, fmt.Errorf("DeepSeek API Key 不能为空")
	}

	dsConfig := &model.DSModelConfig{
		APIKey:  cfg.DeepSeekAPIKey,
		Model:   cfg.DeepSeekModel,
		BaseURL: cfg.DeepSeekBaseURL,
	}

	chatModel, err := model.CreateDSChatModel(ctx, dsConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 创建LLM调用器
	llmCaller := tool.NewLLMCaller(chatModel)

	// 创建工作流
	timelineWorkflow := workflow.NewTimelineWorkflow(llmCaller)
	graphWorkflow := workflow.NewGraphWorkflow(llmCaller)
	arkModelID := os.Getenv("ARK_MODEL_ID")
	if arkModelID == "" {
		arkModelID = model.DefaultArkModel
	}
	modeWorkflow := workflow.NewModeWorkflow(llmCaller, arkModelID)

	return &NewsTimelineAgent{
		chatModel:        chatModel,
		apiKey:           cfg.DeepSeekAPIKey,
		timelineWorkflow: timelineWorkflow,
		graphWorkflow:    graphWorkflow,
		modeWorkflow:     modeWorkflow,
	}, nil
}

// GenerateTimeline 生成新闻时间链
func (a *NewsTimelineAgent) GenerateTimeline(ctx context.Context, keyword string) (*TimelineResponse, error) {
	result, err := a.timelineWorkflow.Generate(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 将workflow包的类型转换为agent包的类型
	return &TimelineResponse{
		Keyword: result.Keyword,
		Events:  convertEvents(result.Events),
	}, nil
}

// GenerateTimelineWithMode 根据模式生成新闻时间链
func (a *NewsTimelineAgent) GenerateTimelineWithMode(ctx context.Context, keyword string, mode string) (*TimelineResponse, error) {
	switch mode {
	case "fast":
		// Fast模式：调用ark模型+联网功能+整理输出json结构
		return a.generateFastMode(ctx, keyword)
	case "deepsearch":
		// Deepsearch模式：调用ark模型澄清整理关键词，React模式调用ark模型+联网TOOLS，反思通过后，ark整理输出json结构
		return a.generateDeepSearchMode(ctx, keyword)
	case "balanced":
		// 均衡模式：调用ark模型澄清整理关键词，React模式调用使用百度AI搜索，ark模型调用整理输出json结构
		return a.generateBalancedMode(ctx, keyword)
	default:
		// 默认使用fast模式
		return a.generateFastMode(ctx, keyword)
	}
}

// generateFastMode 实现Fast模式：返回模拟数据
func (a *NewsTimelineAgent) generateFastMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	result, err := a.modeWorkflow.GenerateMockMode(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 将workflow包的类型转换为agent包的类型
	convertedEvents := make([]Event, len(result.Events))
	for i, e := range result.Events {
		convertedEvents[i] = Event{
			ID:       e.ID,
			Title:    e.Title,
			Time:     e.Time,
			Location: e.Location,
			People:   e.People,
			Summary:  e.Summary,
		}
	}
	return &TimelineResponse{
		Keyword: result.Keyword,
		Events:  convertedEvents,
	}, nil
}

// generateDeepSearchMode 实现Deepsearch模式：返回模拟数据
func (a *NewsTimelineAgent) generateDeepSearchMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	result, err := a.modeWorkflow.GenerateMockMode(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 将workflow包的类型转换为agent包的类型
	convertedEvents := make([]Event, len(result.Events))
	for i, e := range result.Events {
		convertedEvents[i] = Event{
			ID:       e.ID,
			Title:    e.Title,
			Time:     e.Time,
			Location: e.Location,
			People:   e.People,
			Summary:  e.Summary,
		}
	}
	return &TimelineResponse{
		Keyword: result.Keyword,
		Events:  convertedEvents,
	}, nil
}

// generateBalancedMode 实现均衡模式：返回模拟数据
func (a *NewsTimelineAgent) generateBalancedMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	result, err := a.modeWorkflow.GenerateMockMode(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 将workflow包的类型转换为agent包的类型
	convertedEvents := make([]Event, len(result.Events))
	for i, e := range result.Events {
		convertedEvents[i] = Event{
			ID:       e.ID,
			Title:    e.Title,
			Time:     e.Time,
			Location: e.Location,
			People:   e.People,
			Summary:  e.Summary,
		}
	}
	return &TimelineResponse{
		Keyword: result.Keyword,
		Events:  convertedEvents,
	}, nil
}

// GenerateGraph 生成知识图谱
func (a *NewsTimelineAgent) GenerateGraph(ctx context.Context, timeline *TimelineResponse) (*GraphResponse, error) {
	// 将agent包的类型转换为workflow包的类型
	workflowTimeline := &workflow.TimelineResponse{
		Keyword: timeline.Keyword,
		Events:  convertToWorkflowEvents(timeline.Events),
	}

	result, err := a.graphWorkflow.Generate(ctx, workflowTimeline)
	if err != nil {
		return nil, err
	}

	// 将workflow包的类型转换为agent包的类型
	return &GraphResponse{
		Keyword: result.Keyword,
		Nodes:   convertNodes(result.Nodes),
		Links:   convertLinks(result.Links),
	}, nil
}

// convertEvents 将workflow.Event转换为agent.Event
func convertEvents(events []workflow.Event) []Event {
	result := make([]Event, len(events))
	for i, e := range events {
		result[i] = Event{
			ID:       e.ID,
			Title:    e.Title,
			Time:     e.Time,
			Location: e.Location,
			People:   e.People,
			Summary:  e.Summary,
		}
	}
	return result
}

// convertToWorkflowEvents 将agent.Event转换为workflow.Event
func convertToWorkflowEvents(events []Event) []workflow.Event {
	result := make([]workflow.Event, len(events))
	for i, e := range events {
		result[i] = workflow.Event{
			ID:       e.ID,
			Title:    e.Title,
			Time:     e.Time,
			Location: e.Location,
			People:   e.People,
			Summary:  e.Summary,
		}
	}
	return result
}

// convertNodes 将workflow.GraphNode转换为agent.GraphNode
func convertNodes(nodes []workflow.GraphNode) []GraphNode {
	result := make([]GraphNode, len(nodes))
	for i, n := range nodes {
		result[i] = GraphNode{
			ID:       n.ID,
			Name:     n.Name,
			Category: n.Category,
		}
	}
	return result
}

// convertLinks 将workflow.GraphLink转换为agent.GraphLink
func convertLinks(links []workflow.GraphLink) []GraphLink {
	result := make([]GraphLink, len(links))
	for i, l := range links {
		result[i] = GraphLink{
			Source:   l.Source,
			Target:   l.Target,
			Relation: l.Relation,
		}
	}
	return result
}
