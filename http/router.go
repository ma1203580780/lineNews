package http

import (
	"net/http"

	"lineNews/http/controller"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS，允许从 file:// 等来源访问 http://localhost:8080 的接口
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 静态前端页面
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// API 路由
	api := r.Group("/api")
	{
		// 时间链和知识图谱路由
		api.GET("/timeline", controller.HandleTimeline)
		api.GET("/graph", controller.HandleGraph)

		// 健康检查
		api.GET("/health", controller.HandleHealthCheck)

		// 百度深度搜索路由
		deepSearch := api.Group("/deepsearch")
		{
			deepSearch.GET("/search", controller.HandleDeepSearch)        // GET /api/deepsearch/search?query=xxx
			deepSearch.POST("/custom", controller.HandleDeepSearchCustom) // POST /api/deepsearch/custom
		}

		// 百度百科路由
		baike := api.Group("/baike")
		{
			baike.GET("/search", controller.HandleBaikeSearch)         // GET /api/baike/search?keyword=xxx
			baike.GET("/lemma", controller.HandleBaikeSearchByLemmaId) // GET /api/baike/lemma?lemma_id=xxx
		}

		// Ark Chat 路由
		arkchat := api.Group("/arkchat")
		{
			arkchat.GET("/chat", controller.HandleArkChat)         // GET /api/arkchat/chat?message=xxx
			arkchat.GET("/stream", controller.HandleArkChatStream) // GET /api/arkchat/stream?message=xxx
		}
	}

	return r
}
