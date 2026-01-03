package workflow

import (
	"context"
	"fmt"

	"lineNews/agent/logutil"
	"lineNews/agent/prompt"
	"lineNews/agent/tool"
	"lineNews/model"
)

// StreamModeWorkflow 支持流式的模式工作流
type StreamModeWorkflow struct {
	streamLLMCaller *tool.StreamLLMCaller
}

// NewStreamModeWorkflow 创建支持流式的模式工作流
func NewStreamModeWorkflow(streamLLMCaller *tool.StreamLLMCaller) *StreamModeWorkflow {
	return &StreamModeWorkflow{
		streamLLMCaller: streamLLMCaller,
	}
}

// GenerateFastModeStream 流式生成Fast模式时间链
func (w *StreamModeWorkflow) GenerateFastModeStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行Fast模式流式生成，关键词: %s", keyword)

	// 1. 调用百度百科接口获取信息
	baikeResp, err := model.BaiduBaikeSearchSimple(keyword)
	if err != nil {
		logutil.LogError("百度百科搜索失败: %v", err)
		// 如果百度百科失败，使用默认流程
		return w.generateInitialStream(ctx, keyword, sendEvent)
	}

	// 2. 将百度百科结果格式化为适合时间线的数据
	var baikeContent string
	if baikeResp.Result != nil {
		baikeContent = fmt.Sprintf("词条标题: %s\n摘要: %s\n描述: %s",
			baikeResp.Result.LemmaTitle,
			baikeResp.Result.Summary,
			baikeResp.Result.LemmaDesc)

		// 发送思考过程
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: fmt.Sprintf("从百度百科获取到%s相关信息", baikeResp.Result.LemmaTitle),
			Stage:   "数据获取",
		})
	} else {
		logutil.LogInfo("百度百科未找到相关结果，使用默认流程")
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "百度百科未找到相关信息，使用默认生成流程",
			Stage:   "数据获取",
		})
		return w.generateInitialStream(ctx, keyword, sendEvent)
	}

	// 3. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下百度百科内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		keyword,
		baikeContent,
	)

	var timeline TimelineResponse
	err = w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"Fast模式时间链生成",
		&timeline,
		sendEvent,
	)
	if err != nil {
		logutil.LogError("Fast模式流式生成失败: %v，回退到默认流程", err)
		sendEvent(tool.StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("Fast模式生成失败: %v", err),
			Stage:   "生成",
		})
		return w.generateInitialStream(ctx, keyword, sendEvent)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = keyword
	}

	sendEvent(tool.StreamEvent{
		Type:    "done",
		Content: "Fast模式时间链生成完成",
		Stage:   "完成",
	})

	logutil.LogInfo("Fast模式流式生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// GenerateDeepSearchModeStream 流式生成Deepsearch模式时间链
func (w *StreamModeWorkflow) GenerateDeepSearchModeStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行DeepSearch模式流式生成，关键词: %s", keyword)

	// 1. 使用ark模型澄清整理关键词
	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "正在澄清和优化搜索关键词",
		Stage:   "关键词澄清",
	})

	refinedKeyword, err := w.refineKeywordStream(ctx, keyword, sendEvent)
	if err != nil {
		logutil.LogError("关键词澄清失败: %v，使用原始关键词", err)
		refinedKeyword = keyword
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "关键词澄清失败，使用原始关键词",
			Stage:   "关键词澄清",
		})
	} else {
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: fmt.Sprintf("关键词已优化为: %s", refinedKeyword),
			Stage:   "关键词澄清",
		})
	}

	// 2. 使用百度深度搜索获取信息
	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "正在使用深度搜索获取相关信息",
		Stage:   "信息检索",
	})

	deepSearchResp, err := model.BaiduDeepSearchSimple(fmt.Sprintf("按照时间线梳理%s相关信息", refinedKeyword))
	if err != nil {
		logutil.LogError("百度深度搜索失败: %v", err)
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "深度搜索失败，使用默认流程",
			Stage:   "信息检索",
		})
		// 如果深度搜索失败，使用默认流程
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	var searchContent string
	if len(deepSearchResp.Choices) > 0 {
		searchContent = deepSearchResp.Choices[0].Message.Content
	} else {
		logutil.LogInfo("深度搜索未返回内容，使用默认流程")
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "深度搜索未返回内容，使用默认流程",
			Stage:   "信息检索",
		})
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	// 3. 添加参考信息到内容中
	for _, ref := range deepSearchResp.References {
		searchContent += fmt.Sprintf("\n参考信息: %s - %s", ref.Title, ref.URL)
	}

	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "已获取搜索结果，正在整理时间线数据",
		Stage:   "数据整理",
	})

	// 4. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下深度搜索内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		refinedKeyword,
		searchContent,
	)

	var timeline TimelineResponse
	err = w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"DeepSearch模式时间链生成",
		&timeline,
		sendEvent,
	)
	if err != nil {
		logutil.LogError("DeepSearch模式流式生成失败: %v", err)
		sendEvent(tool.StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("DeepSearch模式生成失败: %v", err),
			Stage:   "生成",
		})
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = refinedKeyword
	}

	sendEvent(tool.StreamEvent{
		Type:    "done",
		Content: "DeepSearch模式时间链生成完成",
		Stage:   "完成",
	})

	logutil.LogInfo("DeepSearch模式流式生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// GenerateBalancedModeStream 流式生成均衡模式时间链
func (w *StreamModeWorkflow) GenerateBalancedModeStream(
	ctx context.Context,
	keyword string,
	sendEvent func(tool.StreamEvent) error,
) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行Balanced模式流式生成，关键词: %s", keyword)

	// 1. 使用ark模型澄清整理关键词
	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "正在澄清和优化搜索关键词",
		Stage:   "关键词澄清",
	})

	refinedKeyword, err := w.refineKeywordStream(ctx, keyword, sendEvent)
	if err != nil {
		logutil.LogError("关键词澄清失败: %v，使用原始关键词", err)
		refinedKeyword = keyword
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "关键词澄清失败，使用原始关键词",
			Stage:   "关键词澄清",
		})
	} else {
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: fmt.Sprintf("关键词已优化为: %s", refinedKeyword),
			Stage:   "关键词澄清",
		})
	}

	// 2. 使用百度深度搜索获取信息
	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "正在使用AI搜索获取相关信息",
		Stage:   "信息检索",
	})

	deepSearchResp, err := model.BaiduDeepSearchSimple(fmt.Sprintf("按照时间线梳理%s相关信息", refinedKeyword))
	if err != nil {
		logutil.LogError("百度深度搜索失败: %v", err)
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "AI搜索失败，使用默认流程",
			Stage:   "信息检索",
		})
		// 如果深度搜索失败，使用默认流程
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	var searchContent string
	if len(deepSearchResp.Choices) > 0 {
		searchContent = deepSearchResp.Choices[0].Message.Content
	} else {
		logutil.LogInfo("深度搜索未返回内容，使用默认流程")
		sendEvent(tool.StreamEvent{
			Type:    "thinking",
			Content: "AI搜索未返回内容，使用默认流程",
			Stage:   "信息检索",
		})
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	// 3. 添加参考信息到内容中
	for _, ref := range deepSearchResp.References {
		searchContent += fmt.Sprintf("\n参考信息: %s - %s", ref.Title, ref.URL)
	}

	sendEvent(tool.StreamEvent{
		Type:    "thinking",
		Content: "已获取搜索结果，正在整理时间线数据",
		Stage:   "数据整理",
	})

	// 4. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下搜索内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		refinedKeyword,
		searchContent,
	)

	var timeline TimelineResponse
	err = w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"Balanced模式时间链生成",
		&timeline,
		sendEvent,
	)
	if err != nil {
		logutil.LogError("Balanced模式流式生成失败: %v", err)
		sendEvent(tool.StreamEvent{
			Type:    "error",
			Content: fmt.Sprintf("Balanced模式生成失败: %v", err),
			Stage:   "生成",
		})
		return w.generateInitialStream(ctx, refinedKeyword, sendEvent)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = refinedKeyword
	}

	sendEvent(tool.StreamEvent{
		Type:    "done",
		Content: "Balanced模式时间链生成完成",
		Stage:   "完成",
	})

	logutil.LogInfo("Balanced模式流式生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// refineKeywordStream 使用LLM澄清和优化关键词（流式）
func (w *StreamModeWorkflow) refineKeywordStream(ctx context.Context, keyword string, sendEvent func(tool.StreamEvent) error) (string, error) {
	userPrompt := fmt.Sprintf("请澄清和优化以下搜索关键词：「%s」。返回优化后的关键词，直接输出，不要有其他解释。", keyword)

	var result string
	err := w.streamLLMCaller.CallAndUnmarshalStream(
		ctx,
		"你是一个专业的关键词优化助手，能够澄清和优化搜索关键词。",
		userPrompt,
		"关键词澄清优化",
		&result,
		sendEvent,
	)
	if err != nil {
		return keyword, err
	}

	// 确保返回的关键词不为空
	if result == "" {
		return keyword, fmt.Errorf("关键词澄清结果为空")
	}

	return result, nil
}

// generateInitialStream 流式生成初始时间链
func (w *StreamModeWorkflow) generateInitialStream(
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
		return nil, fmt.Errorf("流式生成时间链失败: %w", err)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = keyword
	}

	logutil.LogInfo("流式初次生成完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}
