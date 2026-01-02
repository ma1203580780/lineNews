package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	DeepSeekAPIKey        string
	DeepSeekModel         string
	DeepSeekBaseURL       string
	ArkAPIKey             string
	ArkModelID            string
	BaiduBaikeAPIKey      string
	BaiduDeepSearchAPIKey string
	ServerPort            string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
		// 如果 .env 文件不存在，继续使用环境变量
	}

	config := &Config{
		DeepSeekAPIKey:        getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:         getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		DeepSeekBaseURL:       getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		ArkAPIKey:             getEnv("ARK_API_KEY", ""),
		ArkModelID:            getEnv("ARK_MODEL_ID", "doubao-seed-1-6-251015"),
		BaiduBaikeAPIKey:      getEnv("BAIDU_BAIKE_API_KEY", ""),
		BaiduDeepSearchAPIKey: getEnv("BAIDU_DEEPSEARCH_API_KEY", ""),
		ServerPort:            getEnv("SERVER_PORT", "8080"),
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
