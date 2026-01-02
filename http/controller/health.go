package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealthCheck 健康检查接口
func HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
