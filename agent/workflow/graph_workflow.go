package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lineNews/agent/prompt"
	"lineNews/agent/tool"
)

// GraphResponse 图谱响应结构（从types.go复制）
type GraphResponse struct {
	Keyword string      `json:"keyword"`
	Nodes   []GraphNode `json:"nodes"`
	Links   []GraphLink `json:"links"`
}

// GraphNode 图谱节点（从types.go复制）
type GraphNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// GraphLink 图谱连接（从types.go复制）
type GraphLink struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

// GraphWorkflow 知识图谱生成工作流
type GraphWorkflow struct {
	llmCaller *tool.LLMCaller
}

// NewGraphWorkflow 创建知识图谱工作流
func NewGraphWorkflow(llmCaller *tool.LLMCaller) *GraphWorkflow {
	return &GraphWorkflow{
		llmCaller: llmCaller,
	}
}

// Generate 生成知识图谱（包含初次生成和反思优化）
func (w *GraphWorkflow) Generate(ctx context.Context, timeline *TimelineResponse) (*GraphResponse, error) {
	if timeline == nil {
		return nil, fmt.Errorf("时间链为空")
	}

	// 第一步：初次生成
	graph, err := w.generateInitial(ctx, timeline)
	if err != nil {
		return nil, err
	}

	// 第二步：反思优化（最多3轮）
	const maxGraphRefineRounds = 3
	for i := 0; i < maxGraphRefineRounds; i++ {
		log.Printf("[GraphWorkflow] 第 %d 轮反思优化开始，当前节点数: %d，边数: %d", i+1, len(graph.Nodes), len(graph.Links))

		refinedGraph, err := w.refine(ctx, timeline.Keyword, graph)
		if err != nil {
			log.Printf("[GraphWorkflow] 第 %d 轮反思优化失败: %v", i+1, err)
			break
		}
		if refinedGraph == nil || len(refinedGraph.Nodes) == 0 {
			log.Printf("[GraphWorkflow] 第 %d 轮反思优化返回空结果，停止进一步反思", i+1)
			break
		}

		log.Printf("[GraphWorkflow] 第 %d 轮反思优化后节点数: %d，边数: %d", i+1, len(refinedGraph.Nodes), len(refinedGraph.Links))
		graph = refinedGraph

		// 如果节点数量已经在理想范围内，则提前结束循环
		if len(graph.Nodes) >= 20 && len(graph.Nodes) <= 100 {
			log.Printf("[GraphWorkflow] 反思优化后节点数量已满足要求（%d 个），结束反思循环", len(graph.Nodes))
			break
		}
	}

	log.Printf("[GraphWorkflow] 最终知识图谱生成完成，包含 %d 个节点和 %d 条边", len(graph.Nodes), len(graph.Links))
	return graph, nil
}

// generateInitial 初次生成知识图谱
func (w *GraphWorkflow) generateInitial(ctx context.Context, timeline *TimelineResponse) (*GraphResponse, error) {
	// 将时间链转换为 JSON 字符串
	timelineJSON, err := json.Marshal(timeline)
	if err != nil {
		return nil, fmt.Errorf("序列化时间链失败: %w", err)
	}

	userPrompt := fmt.Sprintf("请根据以下时间链构建知识图谱：\n%s", string(timelineJSON))

	var graph GraphResponse
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.GraphGenerationSystemPrompt,
		userPrompt,
		"知识图谱初次生成",
		&graph,
	)
	if err != nil {
		return nil, fmt.Errorf("生成知识图谱失败: %w", err)
	}

	// 确保 keyword 被正确设置
	if graph.Keyword == "" {
		graph.Keyword = timeline.Keyword
	}

	log.Printf("[GraphWorkflow] 初次生成完成，包含 %d 个节点和 %d 条边", len(graph.Nodes), len(graph.Links))
	return &graph, nil
}

// refine 反思优化知识图谱
func (w *GraphWorkflow) refine(ctx context.Context, keyword string, original *GraphResponse) (*GraphResponse, error) {
	if original == nil {
		return nil, fmt.Errorf("原始知识图谱为空")
	}

	// 将原始知识图谱转换为 JSON 字符串
	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("序列化原始知识图谱失败: %w", err)
	}

	userPrompt := fmt.Sprintf(
		"下面是模型第一次为关键词「%s」生成的知识图谱 JSON：\n%s\n\n请在内部使用 ReAct 模式进行反思和推理，检查节点数量和内容是否满足上述要求，并在此基础上进行补充、合并和优化，生成最终的高质量知识图谱。请直接返回最终的 JSON，不要输出任何解释性文字。",
		keyword,
		string(originalJSON),
	)

	var refined GraphResponse
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.GraphRefinementSystemPrompt,
		userPrompt,
		"知识图谱反思优化",
		&refined,
	)
	if err != nil {
		return nil, fmt.Errorf("反思优化知识图谱失败: %w", err)
	}

	if refined.Keyword == "" {
		refined.Keyword = keyword
	}

	// 再次做数量上的兜底校验，如果仍然过少，则保留原结果
	if len(refined.Nodes) < 10 {
		log.Printf("[GraphWorkflow] 反思后节点数过少(%d)，保留原始知识图谱(%d)", len(refined.Nodes), len(original.Nodes))
		return original, nil
	}

	return &refined, nil
}
