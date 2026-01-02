package controller

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"lineNews/agent"

	"github.com/gin-gonic/gin"
)

// AgentManager Agent 管理器
type AgentManager struct {
	agent *agent.NewsTimelineAgent
	mu    sync.RWMutex
	cache map[string]*CacheEntry
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Timeline  *agent.TimelineResponse
	Graph     *agent.GraphResponse
	Timestamp time.Time
}

var (
	agentManager *AgentManager
	agentOnce    sync.Once
)

// InitAgent 初始化 Agent
func InitAgent(ctx context.Context, apiKey string) error {
	var initErr error
	agentOnce.Do(func() {
		agentInstance, err := agent.NewNewsTimelineAgent(ctx, apiKey)
		if err != nil {
			initErr = err
			return
		}

		agentManager = &AgentManager{
			agent: agentInstance,
			cache: make(map[string]*CacheEntry),
		}

		fmt.Println("[Controller] Agent 初始化成功")
	})

	return initErr
}

// getOrGenerateTimeline 获取或生成时间链（带缓存）
func (am *AgentManager) getOrGenerateTimeline(ctx context.Context, keyword string) (*agent.TimelineResponse, error) {
	// 检查缓存
	am.mu.RLock()
	if entry, ok := am.cache[keyword]; ok {
		// 缓存10分钟有效
		if time.Since(entry.Timestamp) < 10*time.Minute {
			am.mu.RUnlock()
			fmt.Printf("[Controller] 使用缓存的时间链: %s\n", keyword)
			return entry.Timeline, nil
		}
	}
	am.mu.RUnlock()

	// 生成新的时间链
	fmt.Printf("[Controller] 开始从 Agent 生成时间链: %s\n", keyword)
	timeline, err := am.agent.GenerateTimeline(ctx, keyword)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	am.mu.Lock()
	am.cache[keyword] = &CacheEntry{
		Timeline:  timeline,
		Timestamp: time.Now(),
	}
	am.mu.Unlock()

	return timeline, nil
}

// getOrGenerateGraph 获取或生成知识图谱（带缓存）
func (am *AgentManager) getOrGenerateGraph(ctx context.Context, keyword string, timeline *agent.TimelineResponse) (*agent.GraphResponse, error) {
	// 检查缓存
	am.mu.RLock()
	if entry, ok := am.cache[keyword]; ok && entry.Graph != nil {
		if time.Since(entry.Timestamp) < 10*time.Minute {
			am.mu.RUnlock()
			fmt.Printf("[Controller] 使用缓存的知识图谱: %s\n", keyword)
			return entry.Graph, nil
		}
	}
	am.mu.RUnlock()

	// 生成新的图谱
	fmt.Printf("[Controller] 开始从 Agent 生成图谱: %s\n", keyword)
	graph, err := am.agent.GenerateGraph(ctx, timeline)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	am.mu.Lock()
	if entry, ok := am.cache[keyword]; ok {
		entry.Graph = graph
	} else {
		am.cache[keyword] = &CacheEntry{
			Timeline:  timeline,
			Graph:     graph,
			Timestamp: time.Now(),
		}
	}
	am.mu.Unlock()

	return graph, nil
}

// HandleTimeline 处理时间链请求
func HandleTimeline(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = "新闻"
	}

	ctx := c.Request.Context()

	// 使用 Agent 生成时间链
	timeline, err := agentManager.getOrGenerateTimeline(ctx, keyword)
	if err != nil {
		fmt.Printf("[Controller] 生成时间链失败: %v\n", err)
		// 失败时使用 mock 数据作为后备
		data := mockTimeline(keyword)
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// HandleGraph 处理知识图谱请求
func HandleGraph(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword query parameter is required"})
		return
	}

	ctx := c.Request.Context()

	// 先获取时间链
	timeline, err := agentManager.getOrGenerateTimeline(ctx, keyword)
	if err != nil {
		fmt.Printf("[Controller] 获取时间链失败: %v\n", err)
		data := mockGraph(keyword)
		c.JSON(http.StatusOK, data)
		return
	}

	// 再生成图谱
	graph, err := agentManager.getOrGenerateGraph(ctx, keyword, timeline)
	if err != nil {
		fmt.Printf("[Controller] 生成图谱失败: %v\n", err)
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
