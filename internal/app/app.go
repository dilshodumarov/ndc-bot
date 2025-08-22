// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"log"
	"ndc/ai_bot/api"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/chatgpt"
	"ndc/ai_bot/internal/infrastructure/gemini"
	instagram "ndc/ai_bot/internal/infrastructure/instagram"
	"ndc/ai_bot/internal/infrastructure/telegram"
	"ndc/ai_bot/internal/infrastructure/telegramuser"
	psql "ndc/ai_bot/internal/repo/postgres"
	redisRepo "ndc/ai_bot/internal/repo/redis"
	uscase "ndc/ai_bot/internal/usecase/postgres"
	uscaseredis "ndc/ai_bot/internal/usecase/redis"
	"ndc/ai_bot/pkg/logger"
	"ndc/ai_bot/pkg/postgres"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botMap = make(map[string]*telegram.Handler)
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {

	l := logger.New(cfg.Log.Level)

	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: "",
		DB:       0,
	})

	// Use uscase
	translationUseCase := uscase.New(
		uscase.UseCase{
			Repo:      psql.New(pg),
			RepoOrder: psql.NewOrderRepo(pg),
			AuthRepo:  psql.NewAuthRepo(pg),
			Chat:      psql.NewChatepo(pg),
			Business:  psql.NewBusinessRepo(pg),
		},
	)

	redisUscase := uscaseredis.NewRedisRepo(
		redisRepo.NewClientStateRepo(rdb),
	)

	// Gemini Model
	newGeminiModel, err := gemini.NewGeminiModel(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - gemini.NewGeminiModel: %w", err))
		return
	}

	NewChatgptModel, err := chatgpt.NewChatGPTModel(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - gemini.NewChatgptModel: %w", err))
		return
	}
	fmt.Println(1111,NewChatgptModel)
	// Get bot integrations
	res, err := translationUseCase.AuthRepo.GetBotIntegrations(context.Background())
	if err != nil {
		log.Fatalf("Xatolik: bot integrationlarni olishda muammo: %v", err)
		return
	}
	telegramUscase, err := telegramuser.NewHandler(cfg, newGeminiModel, NewChatgptModel, redisUscase, translationUseCase)
	if err != nil {
		log.Fatalf("Xatolik: telegramUscase: %v", err)
		return
	}
	instagramUscase, err := instagram.NewHandler(cfg, newGeminiModel, redisUscase, translationUseCase)
	if err != nil {
		log.Fatalf("Xatolik: telegramUscase: %v", err)
		return
	}
	telegram.CreateGlobalVar(newGeminiModel, NewChatgptModel,translationUseCase, redisUscase)
	for _, bot := range res {
		go func(bot *entity.BotIntegration) {

			tgBot, err := telegram.NewHandler(cfg, bot.Token, bot.BusinessID, bot.UserID, newGeminiModel, NewChatgptModel,translationUseCase, redisUscase)
			if err != nil {
				log.Printf("Bot yaratishda xatolik (BusinessID: %s): %v", bot.BusinessID, err)
				return
			}
			AddBotMap(bot.BusinessID, tgBot)
			ctx, cancel := context.WithCancel(context.Background())

			tgBot.SetCancelFunc(cancel)

			telegram.AddBotMap(bot.BusinessID, tgBot)

			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60
			updates := tgBot.TelegramBot.GetUpdatesChan(u)

			for {
				select {
				case <-ctx.Done():
					log.Printf("Bot %s kontekst orqali to‘xtatildi", bot.BusinessID)
					return
				case update := <-updates:
					if update.Message != nil {
						go tgBot.HandleTelegramMessage(update.Message)
					}
				}
			}
		}(bot)
	}

	r := gin.Default()
	api.NewTelegramRoutes(*telegramUscase, *instagramUscase, cfg, r)
	api.NewRouter(r)
	r.Run(":8089")

}

func AddBotMap(guid string, handler *telegram.Handler) {
	botMap[guid] = handler
}

func GetBotByGuid(guid string) *telegram.Handler {
	return botMap[guid]
}
