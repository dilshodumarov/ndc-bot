// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"log"
	"ndc/ai_bot/api"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/gemini"
	"ndc/ai_bot/internal/infrastructure/telegram"
	"ndc/ai_bot/internal/infrastructure/telegramuser"
	"ndc/ai_bot/internal/repo/persistent"
	redisRepo "ndc/ai_bot/internal/repo/redis"
	"ndc/ai_bot/internal/usecase/product"
	uscaseredis "ndc/ai_bot/internal/usecase/redis"
	instagram "ndc/ai_bot/internal/infrastructure/instagram"
	"ndc/ai_bot/pkg/logger"
	"ndc/ai_bot/pkg/postgres"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

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

	// Use case
	translationUseCase := product.New(
		persistent.New(pg),
		persistent.NewOrderRepo(pg),
		persistent.NewAuthRepo(pg),
	)

	redisUscase := uscaseredis.NewRedisRepo(
		redisRepo.NewClientStateRepo(rdb),
	)

	// Gemini Model
	newGeminiModel, err := gemini.NewGeminiModel(cfg)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - gemini.NewGeminiModel: %w", err))
	}

	// Get bot integrations
	res, err := translationUseCase.AuthRepo.GetBotIntegrations(context.Background())
	if err != nil {
		log.Fatalf("Xatolik: bot integrationlarni olishda muammo: %v", err)
		return
	}
	telegramUscase, err := telegramuser.NewHandler(cfg, newGeminiModel, redisUscase, translationUseCase)
	if err != nil {
		log.Fatalf("Xatolik: telegramUscase: %v", err)
		return
	}
	instagramUscase, err := instagram.NewHandler(cfg, newGeminiModel, redisUscase, translationUseCase)
	if err != nil {
		log.Fatalf("Xatolik: telegramUscase: %v", err)
		return
	}
	telegram.CreateGlobalVar(newGeminiModel, translationUseCase, redisUscase)
	for _, bot := range res {
		go func(bot *entity.BotIntegration) {

			tgBot, err := telegram.NewHandler(cfg, bot.Token, bot.BusinessID, bot.UserID,newGeminiModel, translationUseCase, redisUscase)
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

	c := cron.New()
	c.AddFunc("0 0 * * *", func() {
		fmt.Println("Cron ishladi")

		orders, err := translationUseCase.GetUsersByLastOrder(context.Background())
		if err != nil {
			fmt.Println("Xatolik:", err)
			return
		}

		for i := 0; i < len(orders); i++ {
			order := orders[i]
			bot := GetBotByGuid(order.BotGuid) // ← Botni to‘g‘ri topamiz

			if bot != nil {

				bot.SendTelegramMessage(context.Background(), entity.SendMessageModel{
					ChatID:  order.ChatId,
					Message: "Buyurtma qilining!!!",
				})
			} else {
				fmt.Printf("Bot topilmadi: %s\n", order.BotGuid)
			}
		}
	})

	c.Start()
	r := gin.Default()
	api.NewTelegramRoutes(*telegramUscase, *instagramUscase,cfg, r)
	api.NewRouter(r)
	r.Run(":8081")

}

func AddBotMap(guid string, handler *telegram.Handler) {
	botMap[guid] = handler
}

func GetBotByGuid(guid string) *telegram.Handler {
	return botMap[guid]
}
