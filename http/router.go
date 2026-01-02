package http

import (
	"lineNews/http/controller"
	"lineNews/http/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 静态前端页面
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/main.html")
	})

	// API 路由组 - 只对API接口应用限流和CORS中间件
	api := r.Group("/api")
	{
		// 在API路由组上应用CORS和限流中间件
		api.Use(middleware.CORSMiddleware())
		api.Use(middleware.GlobalRateLimiter.Limit())

		// 基础服务路由
		api.GET("/health", controller.HandleHealthCheck)             // 健康检查
		api.GET("/timeline", controller.HandleTimeline)              // 时间链
		api.GET("/timeline/stream", controller.HandleTimelineStream) // 时间链流式
		api.GET("/graph", controller.HandleGraph)                    // 知识图谱

		// 百度深度搜索路由
		api.GET("/deepsearch/search", controller.HandleDeepSearch)        // GET /api/deepsearch/search?query=xxx
		api.POST("/deepsearch/custom", controller.HandleDeepSearchCustom) // POST /api/deepsearch/custom

		// 百度百科路由
		api.GET("/baike/search", controller.HandleBaikeSearch)         // GET /api/baike/search?keyword=xxx
		api.GET("/baike/lemma", controller.HandleBaikeSearchByLemmaId) // GET /api/baike/lemma?lemma_id=xxx

		// Ark Chat 路由
		api.GET("/arkchat/chat", controller.HandleArkChat)         // GET /api/arkchat/chat?message=xxx
		api.GET("/arkchat/stream", controller.HandleArkChatStream) // GET /api/arkchat/stream?message=xxx
	}

	return r
}
