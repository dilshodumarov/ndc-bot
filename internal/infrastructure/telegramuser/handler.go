package telegramuser

import (
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/infrastructure/chatgpt"
	"ndc/ai_bot/internal/infrastructure/gemini"
	uscase "ndc/ai_bot/internal/usecase/postgres"
	redis "ndc/ai_bot/internal/usecase/redis"
)

type Handler struct {
	GeminiModel  *gemini.Gemini
	cfg          *config.Config
	RedisUsecase *redis.Uscase
	ProductUscse *uscase.UseCase
	chatgpt      *chatgpt.ChatGpt
}

func NewHandler(cfg *config.Config, geminiModel *gemini.Gemini, chatgpt *chatgpt.ChatGpt, Redis *redis.Uscase, ProductUscse *uscase.UseCase) (*Handler, error) {
	return &Handler{
		GeminiModel:  geminiModel,
		cfg:          cfg,
		RedisUsecase: Redis,
		ProductUscse: ProductUscse,
		chatgpt:      chatgpt,
	}, nil
}
