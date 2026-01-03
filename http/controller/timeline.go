package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"lineNews/agent"
	"lineNews/agent/logutil"
	"lineNews/config"
	"lineNews/model"

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
	// 直接调用Ark模型生成时间链
	logutil.LogInfo("开始从 Ark 模型生成时间链: %s (模式: %s)", keyword, mode)

	// 从环境变量获取配置
	arkAPIKey := os.Getenv("ARK_API_KEY")
	if arkAPIKey == "" {
		return nil, fmt.Errorf("ARK_API_KEY 环境变量未设置")
	}

	arkModelID := os.Getenv("ARK_MODEL_ID")
	if arkModelID == "" {
		arkModelID = model.DefaultArkModel
	}

	// 构建用户请求
	userPrompt := fmt.Sprintf("请为关键词 '%s' 生成新闻时间线，返回有效的JSON格式结果，包含Keyword和Events字段。Events数组应包含至少5-10个独立的新闻事件，每个事件必须包含以下字段：ID（字符串类型，如\"1\", \"2\", \"3\"等）、Title（字符串，事件标题）、Time（字符串，具体时间如\"2024-01-15\"）、Location（字符串，地点）、People（字符串数组，涉及人物）、Summary（字符串，事件摘要）。确保时间线覆盖不同时间段，从早期到近期，每个事件都应有明确的时间、地点、人物和内容。", keyword)

	// 调用Ark模型
	response, err := model.SendArkMessage(ctx, arkModelID, userPrompt, "你是一个专业的新闻时间线生成助手。")
	if err != nil {
		return nil, fmt.Errorf("调用Ark模型失败: %w", err)
	}

	// 打印模型原始输出日志
	logutil.LogInfo("Ark模型原始输出: %s", response.Content)

	// 解析返回的JSON
	var timeline agent.TimelineResponse
	// 检查响应内容是否为空
	trimmedContent := strings.TrimSpace(response.Content)
	if trimmedContent == "" {
		logutil.LogInfo("Ark模型返回空响应")
		// 返回一个基本的响应
		timeline = agent.TimelineResponse{
			Keyword: keyword,
			Events: []agent.Event{
				{
					ID:       "1",
					Title:    fmt.Sprintf("%s 相关事件", keyword),
					Time:     "2024-01-01",
					Location: "未知地点",
					People:   []string{"未知人物"},
					Summary:  "模型未返回有效内容",
				},
			},
		}
	} else if err := json.Unmarshal([]byte(trimmedContent), &timeline); err != nil {
		// 如果直接解析失败，尝试从响应中提取JSON部分
		logutil.LogInfo("直接解析JSON失败，尝试提取: %v", err)
		jsonStart := findJSONStart(trimmedContent)
		if jsonStart != -1 {
			jsonContent := trimmedContent[jsonStart:]
			// 尝试找到JSON的结束位置
			jsonEnd := findJSONEnd(jsonContent)
			if jsonEnd != -1 {
				jsonContent = jsonContent[:jsonEnd+1]
			}
			if err := json.Unmarshal([]byte(jsonContent), &timeline); err == nil {
				logutil.LogInfo("成功提取并解析JSON")
			} else {
				logutil.LogInfo("提取JSON后仍解析失败: %v", err)
				// 如果JSON解析失败，创建一个基本的响应
				timeline = agent.TimelineResponse{
					Keyword: keyword,
					Events: []agent.Event{
						{
							ID:       "1",
							Title:    fmt.Sprintf("%s 相关事件", keyword),
							Time:     "2024-01-01",
							Location: "未知地点",
							People:   []string{"未知人物"},
							Summary:  response.Content,
						},
					},
				}
			}
		}
	}

	return &timeline, nil
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

	// 检查是否需要流式响应
	stream := c.Query("stream")
	if stream == "true" || stream == "1" {
		// 使用流式响应
		HandleTimelineStream(c)
		return
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
		c.JSON(http.StatusOK, gin.H{"error": "keyword query parameter is required"})
		return
	}
	mode := c.Query("mode")
	if mode == "" {
		mode = "fast" // 默认模式
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ctx := c.Request.Context()

	// 发送开始事件
	c.SSEvent("start", gin.H{"message": "开始生成时间链", "keyword": keyword})
	c.Writer.Flush()

	// 直接调用Ark模型生成时间链（流式版本）
	logutil.LogInfo("开始从 Ark 模型生成时间链（流式）: %s (模式: %s)", keyword, mode)

	// 从环境变量获取配置
	arkAPIKey := os.Getenv("ARK_API_KEY")
	if arkAPIKey == "" {
		logutil.LogError("ARK_API_KEY 环境变量未设置")
		c.SSEvent("error", gin.H{"error": "ARK_API_KEY 环境变量未设置"})
		c.Writer.Flush()
		return
	}

	arkModelID := os.Getenv("ARK_MODEL_ID")
	if arkModelID == "" {
		arkModelID = model.ArkFlashModel
	}

	// 构建用户请求
	userPrompt := fmt.Sprintf("请为关键词 '%s' 生成新闻时间线，返回有效的JSON格式结果，包含Keyword和Events字段。Events数组应包含至少5-10个独立的新闻事件，每个事件必须包含以下字段：ID（字符串类型，如\"1\", \"2\", \"3\"等）、Title（字符串，事件标题）、Time（字符串，具体时间如\"2024-01-15\"）、Location（字符串，地点）、People（字符串数组，涉及人物）、Summary（字符串，事件摘要）。确保时间线覆盖不同时间段，从早期到近期，每个事件都应有明确的时间、地点、人物和内容。", keyword)

	// 发送思考过程
	c.SSEvent("thinking", gin.H{"message": "正在分析关键词并规划时间线生成"})
	c.Writer.Flush()

	// 调用Ark模型
	response, err := model.SendArkMessage(ctx, arkModelID, userPrompt, "你是一个专业的新闻时间线生成助手。")
	if err != nil {
		logutil.LogError("调用Ark模型失败: %v", err)
		c.SSEvent("error", gin.H{"error": fmt.Sprintf("调用Ark模型失败: %v", err)})
		c.Writer.Flush()
		// 返回 mock 数据
		data := mockTimeline(keyword)
		c.SSEvent("data", data)
		c.Writer.Flush()
		return
	}

	// 打印模型原始输出日志
	logutil.LogInfo("Ark模型原始输出: %s", response.Content)

	// 发送处理中事件
	c.SSEvent("processing", gin.H{"message": "正在解析模型响应"})
	c.Writer.Flush()

	// 解析返回的JSON
	var timeline agent.TimelineResponse
	// 检查响应内容是否为空
	trimmedContent := strings.TrimSpace(response.Content)
	if trimmedContent == "" {
		logutil.LogInfo("Ark模型返回空响应")
		// 返回一个基本的响应
		timeline = agent.TimelineResponse{
			Keyword: keyword,
			Events: []agent.Event{
				{
					ID:       "1",
					Title:    fmt.Sprintf("%s 相关事件", keyword),
					Time:     "2024-01-01",
					Location: "未知地点",
					People:   []string{"未知人物"},
					Summary:  "模型未返回有效内容",
				},
			},
		}
	} else if err := json.Unmarshal([]byte(trimmedContent), &timeline); err != nil {
		// 如果直接解析失败，尝试从响应中提取JSON部分
		logutil.LogInfo("直接解析JSON失败，尝试提取: %v", err)
		jsonStart := findJSONStart(trimmedContent)
		if jsonStart != -1 {
			jsonContent := trimmedContent[jsonStart:]
			// 尝试找到JSON的结束位置
			jsonEnd := findJSONEnd(jsonContent)
			if jsonEnd != -1 {
				jsonContent = jsonContent[:jsonEnd+1]
			}
			if err := json.Unmarshal([]byte(jsonContent), &timeline); err == nil {
				logutil.LogInfo("成功提取并解析JSON")
			} else {
				logutil.LogInfo("提取JSON后仍解析失败: %v", err)
				// 如果JSON解析失败，创建一个基本的响应
				timeline = agent.TimelineResponse{
					Keyword: keyword,
					Events: []agent.Event{
						{
							ID:       "1",
							Title:    fmt.Sprintf("%s 相关事件", keyword),
							Time:     "2024-01-01",
							Location: "未知地点",
							People:   []string{"未知人物"},
							Summary:  response.Content,
						},
					},
				}
			}
		}
	}

	// 发送最终数据
	c.SSEvent("data", timeline)
	c.Writer.Flush()

	// 发送完成事件
	c.SSEvent("complete", gin.H{"message": "时间链生成完成"})
	c.Writer.Flush()
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

// findJSONStart 查找内容中的 JSON 起始位置
func findJSONStart(content string) int {
	start := -1
	for i, char := range content {
		if char == '{' || char == '[' {
			start = i
			break
		}
	}
	return start
}

// findJSONEnd 查找内容中的 JSON 结束位置
func findJSONEnd(content string) int {
	// 寻找匹配的括号或方括号
	stack := 0
	startChar := byte(0)
	for i := 0; i < len(content); i++ {
		char := content[i]
		if char == '{' || char == '[' {
			if startChar == 0 {
				startChar = char
			}
			stack++
		} else if (char == '}' && startChar == '{') || (char == ']' && startChar == '[') {
			stack--
			if stack == 0 {
				return i
			}
		}
	}
	return -1
}
