package workflow

import (
	"context"
	"fmt"
	"strings"

	"lineNews/agent/logutil"
	"lineNews/agent/prompt"
	"lineNews/agent/tool"
	"lineNews/model"
)

// ModeWorkflow 不同模式的工作流
type ModeWorkflow struct {
	llmCaller *tool.LLMCaller
}

// NewModeWorkflow 创建模式工作流
func NewModeWorkflow(llmCaller *tool.LLMCaller) *ModeWorkflow {
	return &ModeWorkflow{
		llmCaller: llmCaller,
	}
}

// GenerateFastMode 实现Fast模式：调用百度百科接口，然后用ark模型来整理输出json结构
func (w *ModeWorkflow) GenerateFastMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行Fast模式，关键词: %s", keyword)

	// 1. 调用百度百科接口获取信息
	baikeResp, err := model.BaiduBaikeSearchSimple(keyword)
	if err != nil {
		logutil.LogError("百度百科搜索失败: %v", err)
		// 如果百度百科失败，使用默认流程
		return w.generateInitial(ctx, keyword)
	}

	// 2. 将百度百科结果格式化为适合时间线的数据
	var baikeContent string
	if baikeResp.Result != nil {
		baikeContent = fmt.Sprintf("词条标题: %s\n摘要: %s\n描述: %s",
			baikeResp.Result.LemmaTitle,
			baikeResp.Result.Summary,
			baikeResp.Result.LemmaDesc)
	} else {
		logutil.LogInfo("百度百科未找到相关结果，使用默认流程")
		return w.generateInitial(ctx, keyword)
	}

	// 3. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下百度百科内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		keyword,
		baikeContent,
	)

	var timeline TimelineResponse
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"Fast模式时间链生成",
		&timeline,
	)
	if err != nil {
		logutil.LogError("Fast模式生成失败: %v，回退到默认流程", err)
		return w.generateInitial(ctx, keyword)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = keyword
	}

	logutil.LogInfo("Fast模式完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// GenerateDeepSearchMode 实现Deepsearch模式：使用ReAct模式调用ark模型+联网工具
func (w *ModeWorkflow) GenerateDeepSearchMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行DeepSearch模式，关键词: %s", keyword)

	// 1. 使用ark模型澄清整理关键词
	refinedKeyword, err := w.refineKeyword(ctx, keyword)
	if err != nil {
		logutil.LogError("关键词澄清失败: %v，使用原始关键词", err)
		refinedKeyword = keyword
	}

	// 2. 使用百度深度搜索获取信息
	deepSearchResp, err := model.BaiduDeepSearchSimple(fmt.Sprintf("按照时间线梳理%s相关信息", refinedKeyword))
	if err != nil {
		logutil.LogError("百度深度搜索失败: %v", err)
		// 如果深度搜索失败，使用默认流程
		return w.generateInitial(ctx, refinedKeyword)
	}

	var searchContent string
	if len(deepSearchResp.Choices) > 0 {
		searchContent = deepSearchResp.Choices[0].Message.Content
	} else {
		logutil.LogInfo("深度搜索未返回内容，使用默认流程")
		return w.generateInitial(ctx, refinedKeyword)
	}

	// 3. 添加参考信息到内容中
	for _, ref := range deepSearchResp.References {
		searchContent += fmt.Sprintf("\n参考信息: %s - %s", ref.Title, ref.URL)
	}

	// 4. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下深度搜索内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		refinedKeyword,
		searchContent,
	)

	var timeline TimelineResponse
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"DeepSearch模式时间链生成",
		&timeline,
	)
	if err != nil {
		logutil.LogError("DeepSearch模式生成失败: %v，回退到默认流程", err)
		return w.generateInitial(ctx, refinedKeyword)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = refinedKeyword
	}

	logutil.LogInfo("DeepSearch模式完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// GenerateBalancedMode 实现均衡模式：使用百度AI搜索，ark模型整理输出
func (w *ModeWorkflow) GenerateBalancedMode(ctx context.Context, keyword string) (*TimelineResponse, error) {
	logutil.LogInfo("开始执行Balanced模式，关键词: %s", keyword)

	// 1. 使用ark模型澄清整理关键词
	refinedKeyword, err := w.refineKeyword(ctx, keyword)
	if err != nil {
		logutil.LogError("关键词澄清失败: %v，使用原始关键词", err)
		refinedKeyword = keyword
	}

	// 2. 使用百度深度搜索获取信息
	deepSearchResp, err := model.BaiduDeepSearchSimple(fmt.Sprintf("按照时间线梳理%s相关信息", refinedKeyword))
	if err != nil {
		logutil.LogError("百度深度搜索失败: %v", err)
		// 如果深度搜索失败，使用默认流程
		return w.generateInitial(ctx, refinedKeyword)
	}

	var searchContent string
	if len(deepSearchResp.Choices) > 0 {
		searchContent = deepSearchResp.Choices[0].Message.Content
	} else {
		logutil.LogInfo("深度搜索未返回内容，使用默认流程")
		return w.generateInitial(ctx, refinedKeyword)
	}

	// 3. 添加参考信息到内容中
	for _, ref := range deepSearchResp.References {
		searchContent += fmt.Sprintf("\n参考信息: %s - %s", ref.Title, ref.URL)
	}

	// 4. 使用ark模型整理输出json结构
	userPrompt := fmt.Sprintf(
		"根据以下搜索内容，生成关于「%s」的新闻时间链JSON格式数据：\n\n%s\n\n请严格按照JSON格式返回，包含Keyword和Events字段，Events中包含ID、Title、Time、Location、People、Summary等字段。",
		refinedKeyword,
		searchContent,
	)

	var timeline TimelineResponse
	err = w.llmCaller.CallAndUnmarshal(
		ctx,
		prompt.TimelineGenerationSystemPrompt,
		userPrompt,
		"Balanced模式时间链生成",
		&timeline,
	)
	if err != nil {
		logutil.LogError("Balanced模式生成失败: %v，回退到默认流程", err)
		return w.generateInitial(ctx, refinedKeyword)
	}

	// 确保 keyword 被正确设置
	if timeline.Keyword == "" {
		timeline.Keyword = refinedKeyword
	}

	logutil.LogInfo("Balanced模式完成，包含 %d 个事件", len(timeline.Events))
	return &timeline, nil
}

// refineKeyword 使用LLM澄清和优化关键词
func (w *ModeWorkflow) refineKeyword(ctx context.Context, keyword string) (string, error) {
	userPrompt := fmt.Sprintf("请澄清和优化以下搜索关键词：「%s」。返回优化后的关键词，直接输出，不要有其他解释。", keyword)

	var result string
	err := w.llmCaller.CallAndUnmarshal(
		ctx,
		"你是一个专业的关键词优化助手，能够澄清和优化搜索关键词。",
		userPrompt,
		"关键词澄清优化",
		&result,
	)
	if err != nil {
		return keyword, err
	}

	// 确保返回的关键词不为空
	if strings.TrimSpace(result) == "" {
		return keyword, fmt.Errorf("关键词澄清结果为空")
	}

	return strings.TrimSpace(result), nil
}

// generateInitial 初次生成时间链（使用原有的逻辑）
func (w *ModeWorkflow) generateInitial(ctx context.Context, keyword string) (*TimelineResponse, error) {
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
