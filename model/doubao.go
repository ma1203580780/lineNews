package model

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/arkbot"
)

func CreateArkChatModel(ctx context.Context) (*arkbot.ChatModel, error) {
	chatModel, err := arkbot.NewChatModel(ctx, &arkbot.Config{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_MODEL_ID"),
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Arkbot ChatModel 失败: %w", err)
	}

	return chatModel, nil
}
