package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealthCheck 健康检查接口
func HandleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": gin.H{
			"deep_search": "available",
			"baike":       "available",
			"ark_chat":    "not_implemented",
		},
	})
}
