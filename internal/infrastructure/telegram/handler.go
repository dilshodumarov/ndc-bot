package telegram

import (
	"context"
	"fmt"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/gemini"
	repo "ndc/ai_bot/internal/usecase/postgres"
	redis "ndc/ai_bot/internal/usecase/redis"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	TelegramBot  *tgbotapi.BotAPI
	GeminiModel  *gemini.Gemini
	Usecase      *repo.UseCase
	cfg          *config.Config
	BusinessId   string
	UserId       string
	ClientStates map[int64]*entity.ClientState
	RedisUsecase *redis.Uscase

	cancelFunc context.CancelFunc 
}


func NewHandler(cfg *config.Config, token,  BusinessId,UserId string, geminiModel *gemini.Gemini, usecase *repo.UseCase, Redis *redis.Uscase) (*Handler, error) {
	tgBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("Error creating Telegram bot: %w", err)
	}
	tgBot.Debug = true

	return &Handler{
		TelegramBot:  tgBot,
		GeminiModel:  geminiModel,
		Usecase:      usecase,
		cfg:          cfg,
		BusinessId:   BusinessId,
		ClientStates: make(map[int64]*entity.ClientState),
		RedisUsecase: Redis,
		UserId:       UserId,
	}, nil
}
