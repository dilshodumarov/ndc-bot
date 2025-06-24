package telegramuser

import (
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/infrastructure/gemini"
	"ndc/ai_bot/internal/usecase/product"
	redis "ndc/ai_bot/internal/usecase/redis"
)

type Handler struct {
	GeminiModel  *gemini.Gemini
	cfg          *config.Config
	RedisUsecase *redis.Uscase
	ProductUscse *product.UseCase
}

func NewHandler(cfg *config.Config, geminiModel *gemini.Gemini, Redis *redis.Uscase,ProductUscse *product.UseCase ) (*Handler, error) {
	return &Handler{
		GeminiModel:  geminiModel,
		cfg:          cfg,
		RedisUsecase: Redis,
		ProductUscse: ProductUscse,
	}, nil
}
