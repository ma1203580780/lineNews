package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"lineNews/agent"
	"lineNews/agent/logutil"
	"lineNews/agent/tool"
	"lineNews/config"

	"github.com/gin-gonic/gin"
)

// AgentManager Agent 管理器
type AgentManager struct {
	agent *agent.NewsTimelineAgent
}

var (
	agentManager *AgentManager
)

// InitAgent 初始化 Agent
func InitAgent(ctx context.Context, cfg *config.Config) error {
	if agentManager != nil {
		return nil // 已经初始化过了
	}

	agentInstance, err := agent.NewNewsTimelineAgent(ctx, cfg)
	if err != nil {
		return err
	}

	agentManager = &AgentManager{
		agent: agentInstance,
	}

	logutil.LogInfo("Agent 初始化成功")
	return nil
}

// generateTimeline 生成时间链
func (am *AgentManager) generateTimeline(ctx context.Context, keyword string, mode string) (*agent.TimelineResponse, error) {
	// 生成时间链
	logutil.LogInfo("开始从 Agent 生成时间链: %s (模式: %s)", keyword, mode)
	timeline, err := am.agent.GenerateTimelineWithMode(ctx, keyword, mode)
	if err != nil {
		return nil, err
	}

	return timeline, nil
}

// generateGraph 生成知识图谱
func (am *AgentManager) generateGraph(ctx context.Context, keyword string, timeline *agent.TimelineResponse, mode string) (*agent.GraphResponse, error) {
	// 生成图谱
	logutil.LogInfo("开始从 Agent 生成图谱: %s (模式: %s)", keyword, mode)
	graph, err := am.agent.GenerateGraph(ctx, timeline)
	if err != nil {
		return nil, err
	}

	return graph, nil
}

// HandleTimeline 处理时间链请求
func HandleTimeline(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusOK, gin.H{"error": "keyword query parameter is required"})
		return
	}
	mode := c.Query("mode")
	if mode == "" {
		mode = "fast" // 默认模式
	}

	ctx := c.Request.Context()

	// 使用 Agent 生成时间链
	timeline, err := agentManager.generateTimeline(ctx, keyword, mode)
	if err != nil {
		logutil.LogError("生成时间链失败: %v", err)
		// 失败时使用 mock 数据作为后备
		data := mockTimeline(keyword)
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// HandleTimelineStream 处理时间链流式请求
func HandleTimelineStream(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = "新闻"
	}
	mode := c.Query("mode")
	if mode == "" {
		mode = "fast" // 默认模式
	}

	ctx := c.Request.Context()

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建发送事件的函数
	sendEvent := func(event tool.StreamEvent) error {
		// 将事件序列化为JSON
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}

		// 发送SSE事件
		_, err = c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
		if err != nil {
			return err
		}

		// 立即刷新响应
		c.Writer.Flush()

		return nil
	}

	// 使用 Agent 流式生成时间链，根据模式选择
	_, err := agentManager.agent.GenerateTimelineStreamWithMode(ctx, keyword, mode, sendEvent)
	if err != nil {
		logutil.LogError("流式生成时间链失败: %v", err)
		// 发送错误事件
		errorEvent := tool.StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("生成时间链失败: %v", err),
			Stage:   "总览",
		}
		data, _ := json.Marshal(errorEvent)
		c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
		c.Writer.Flush()
		return
	}
}

// HandleGraph 处理知识图谱请求
func HandleGraph(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = "新闻"
	}
	mode := c.Query("mode")
	if mode == "" {
		mode = "fast" // 默认模式
	}

	ctx := c.Request.Context()

	// 先获取时间链
	timeline, err := agentManager.generateTimeline(ctx, keyword, mode)
	if err != nil {
		logutil.LogError("获取时间链失败: %v", err)
		data := mockGraph(keyword)
		c.JSON(http.StatusOK, data)
		return
	}

	// 再生成图谱
	graph, err := agentManager.generateGraph(ctx, keyword, timeline, mode)
	if err != nil {
		logutil.LogError("生成图谱失败: %v", err)
		data := mockGraph(keyword)
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(http.StatusOK, graph)
}

// mockTimeline 生成 mock 时间链数据
func mockTimeline(keyword string) agent.TimelineResponse {
	return agent.TimelineResponse{
		Keyword: keyword,
		Events: []agent.Event{
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
			{
				ID:       "4",
				Title:    fmt.Sprintf("%s 相关新闻四：后续影响", keyword),
				Time:     "2023-08-01",
				Location: "深圳",
				People:   []string{"媒体", "专家"},
				Summary:  fmt.Sprintf("%s 对社会、产业或公众情绪产生的长期影响分析。", keyword),
			},
		},
	}
}

// mockGraph 生成 mock 知识图谱数据
func mockGraph(keyword string) agent.GraphResponse {
	nodes := []agent.GraphNode{
		{ID: "e1", Name: fmt.Sprintf("%s 核心事件", keyword), Category: "事件"},
		{ID: "e2", Name: fmt.Sprintf("%s 延伸事件", keyword), Category: "事件"},
		{ID: "p1", Name: "张三", Category: "人物"},
		{ID: "p2", Name: "李四", Category: "人物"},
		{ID: "l1", Name: "北京", Category: "地点"},
		{ID: "l2", Name: "上海", Category: "地点"},
		{ID: "t1", Name: fmt.Sprintf("%s 政策", keyword), Category: "主题"},
	}

	links := []agent.GraphLink{
		{Source: "e1", Target: "p1", Relation: "相关人物"},
		{Source: "e1", Target: "l1", Relation: "发生地点"},
		{Source: "e1", Target: "t1", Relation: "涉及主题"},
		{Source: "e2", Target: "p2", Relation: "相关人物"},
		{Source: "e2", Target: "l2", Relation: "发生地点"},
		{Source: "e2", Target: "t1", Relation: "政策影响"},
		{Source: "e1", Target: "e2", Relation: "事件演化"},
	}

	return agent.GraphResponse{
		Keyword: keyword,
		Nodes:   nodes,
		Links:   links,
	}
}
