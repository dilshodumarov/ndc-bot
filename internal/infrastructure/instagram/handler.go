package telegramuser

import (
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/infrastructure/gemini"
	psql "ndc/ai_bot/internal/usecase/postgres"
	redis "ndc/ai_bot/internal/usecase/redis"
)

type Handler struct {
	GeminiModel  *gemini.Gemini
	cfg          *config.Config
	RedisUsecase *redis.Uscase
	ProductUscse *psql.UseCase
}

func NewHandler(cfg *config.Config, geminiModel *gemini.Gemini, Redis *redis.Uscase,ProductUscse *psql.UseCase ) (*Handler, error) {
	return &Handler{
		GeminiModel:  geminiModel,
		cfg:          cfg,
		RedisUsecase: Redis,
		ProductUscse: ProductUscse,
	}, nil
}
