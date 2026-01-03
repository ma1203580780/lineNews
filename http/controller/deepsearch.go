package controller

import (
	"lineNews/model"
	"net/http"

	"lineNews/agent/logutil"

	"github.com/gin-gonic/gin"
)

// HandleDeepSearch 处理百度深度搜索请求
func HandleDeepSearch(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query 参数不能为空",
		})
		return
	}

	logutil.LogInfo("百度深度搜索请求: %s", query)

	// 调用 model 层
	response, err := model.BaiduDeepSearchSimple(query)
	if err != nil {
		logutil.LogError("深度搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "深度搜索失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"query":   query,
		"data":    response,
	})
}

// HandleDeepSearchCustom 处理自定义配置的深度搜索请求
func HandleDeepSearchCustom(c *gin.Context) {
	var req struct {
		Query              string `json:"query" binding:"required"`
		EnableDeepSearch   bool   `json:"enable_deep_search"`
		MaxCompletionToken int    `json:"max_completion_tokens"`
		SearchRecency      string `json:"search_recency_filter"` // week, month, year
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
		})
		return
	}

	logutil.LogInfo("自定义深度搜索请求: %s", req.Query)

	// 构建自定义配置
	options := model.NewDefaultRequest(req.Query)
	if req.EnableDeepSearch {
		options.EnableDeepSearch = true
	}
	if req.MaxCompletionToken > 0 {
		options.MaxCompletionTokens = req.MaxCompletionToken
	}
	if req.SearchRecency != "" {
		options.SearchRecencyFilter = req.SearchRecency
	}

	// 调用 model 层
	response, err := model.BaiduDeepSearch(req.Query, options)
	if err != nil {
		logutil.LogError("自定义深度搜索失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "深度搜索失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"query":   req.Query,
		"data":    response,
	})
}
