package main

import (
	"context"

	"lineNews/agent/logutil"
	"lineNews/config"
	"lineNews/http"
	"lineNews/http/controller"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化 Agent
	ctx := context.Background()
	if err := controller.InitAgent(ctx, cfg); err != nil {
		logutil.LogError("初始化失败: %v", err)
	}

	// 设置路由
	r := http.SetupRouter()

	// 使用配置中的端口
	port := ":" + cfg.ServerPort
	logutil.LogInfo("服务启动在 http://localhost:%s", cfg.ServerPort)
	if err := r.Run(port); err != nil {
		logutil.LogError("服务启动失败: %v", err)
	}
}
