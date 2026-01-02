package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleArkChat 处理 Ark Chat 请求
func HandleArkChat(c *gin.Context) {
	message := c.Query("message")
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message 参数不能为空",
		})
		return
	}

	log.Printf("[Controller] Ark Chat 请求: %s", message)

	// TODO: 调用 model 层的 Ark Chat 接口
	// 目前返回模拟响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"response": fmt.Sprintf("这是 Ark Chat 对 '%s' 的回复（待实现）", message),
			"model":    "ark-chat",
		},
	})
}

// HandleArkChatStream 处理 Ark Chat 流式请求
func HandleArkChatStream(c *gin.Context) {
	message := c.Query("message")
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message 参数不能为空",
		})
		return
	}

	log.Printf("[Controller] Ark Chat 流式请求: %s", message)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// TODO: 实现流式响应
	c.String(http.StatusOK, "data: 流式响应待实现\n\n")
}
