package controller

import (
	"fmt"
	"net/http"

	"lineNews/agent/logutil"
	"lineNews/model"

	"github.com/gin-gonic/gin"
)

// HandleDeepSeekChat 处理 DeepSeek Chat 请求
func HandleDeepSeekChat(c *gin.Context) {
	message := c.Query("message")
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message 参数不能为空",
		})
		return
	}

	logutil.LogInfo("DeepSeek Chat 请求: %s", message)

	// 创建 DeepSeek 聊天模型
	ctx := c.Request.Context()
	config := model.NewDSModelConfig()
	chatModel, err := model.CreateDSChatModel(ctx, config)
	if err != nil {
		logutil.LogError("创建 DeepSeek 模型失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("创建模型失败: %v", err),
		})
		return
	}

	// 发送消息到 DeepSeek 模型
	response, err := model.SendMessage(ctx, chatModel, message, "你是一个有用的AI助手，请回答用户的问题。")
	if err != nil {
		logutil.LogError("发送消息到 DeepSeek 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("发送消息失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"response": response.Content,
			"model":    "deepseek",
			"usage": gin.H{
				"prompt_tokens":     response.PromptTokens,
				"completion_tokens": response.CompletionTokens,
				"total_tokens":      response.TotalTokens,
			},
		},
	})
}
