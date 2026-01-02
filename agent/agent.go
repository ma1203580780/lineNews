package agent

import (
	"context"
	"fmt"

	"lineNews/agent/tool"
	"lineNews/agent/workflow"
	"lineNews/model"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
)

// NewsTimelineAgent 新闻时间链 Agent
type NewsTimelineAgent struct {
	chatModel              *deepseek.ChatModel
	apiKey                 string
	timelineWorkflow       *workflow.TimelineWorkflow
	graphWorkflow          *workflow.GraphWorkflow
	streamTimelineWorkflow *workflow.StreamTimelineWorkflow
}

// NewNewsTimelineAgent 创建新闻时间链 Agent
func NewNewsTimelineAgent(ctx context.Context, apiKey string) (*NewsTimelineAgent, error) {
	if apiKey == "" {
		apiKey = model.DefaultDeepSeekAPIKey
	}

	config := &model.DSModelConfig{
		APIKey:  apiKey,
		Model:   model.DefaultDeepSeekModel,
		BaseURL: model.DefaultDeepSeekURL,
	}

	chatModel, err := model.CreateDSChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 创建LLM调用器
	llmCaller := tool.NewLLMCaller(chatModel)
	streamLLMCaller := tool.NewStreamLLMCaller(chatModel)

	// 创建工作流
	timelineWorkflow := workflow.NewTimelineWorkflow(llmCaller)
	graphWorkflow := workflow.NewGraphWorkflow(llmCaller)
	streamTimelineWorkflow := workflow.NewStreamTimelineWorkflow(streamLLMCaller)

	return &NewsTimelineAgent{
		chatModel:              chatModel,
		apiKey:                 apiKey,
		timelineWorkflow:       timelineWorkflow,
		graphWorkflow:          graphWorkflow,
		streamTimelineWorkflow: streamTimelineWorkflow,
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

// GenerateTimelineStream 流式生成新闻时间链
func (a *NewsTimelineAgent) GenerateTimelineStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	// 将workflow包的类型转换为agent包的类型
	result, err := a.streamTimelineWorkflow.GenerateStream(ctx, keyword, sendEvent)
	if err != nil {
		return nil, err
	}

	return &TimelineResponse{
		Keyword: result.Keyword,
		Events:  convertEvents(result.Events),
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
