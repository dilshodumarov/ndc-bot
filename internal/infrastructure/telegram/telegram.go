package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/config"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/gemini"
	uscase "ndc/ai_bot/internal/usecase/postgres"
	uscaseredis "ndc/ai_bot/internal/usecase/redis"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	newGeminiModel     *gemini.Gemini
	translationUseCase *uscase.UseCase
	redisUscase        *uscaseredis.Uscase
	errorMEssage       string
)

var BotMap = make(map[string]*Handler)

func (t *Handler) HandleTelegramMessage(message *tgbotapi.Message) {
	var (
		ctx        = context.Background()
		userInput  = message.Text
		chatID     = message.Chat.ID
		MessageId  = message.MessageID
		BussinesId = t.BusinessId
		UserId     = fmt.Sprintf("%d", message.From.ID)
	)
	OrderStatus, err := t.Usecase.GetOrderStatusesByBusinessID(ctx, BussinesId)
	if err != nil {
		fmt.Println("Error while GetOrderStatusesByBusinessID:", err)
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}
	SettingsAi, err := t.Usecase.GetIntegrationSettingsByOwnerID(ctx, BussinesId, UserId)
	if err != nil {
		fmt.Println("Error while GetIntegrationSettingsByOwnerID:", err)
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}
	errorMEssage = SettingsAi.ErrorMessage
	fmt.Println(1111, SettingsAi.PromtOrder)
	if SettingsAi.IsBlocked {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "Assalomu alaykum uzur siz bloklangansiz",
		})
		return
	}
	if message.Photo != nil {

		photo := (message.Photo)[len(message.Photo)-1]

		caption := message.Caption

		fileConfig, err := t.TelegramBot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
		if err != nil {
			log.Println("Faylni olishda xatolik:", err)
			return
		}

		userInput = fileConfig.Link(t.TelegramBot.Token)
		if caption != "" {
			userInput += "caption: " + caption
		}
	}
	replyToMessageID := 0
	if message.ReplyToMessage != nil {
		replyToMessageID = message.ReplyToMessage.MessageID

		var replyText string

		// Agar oddiy text bo‘lsa
		if message.ReplyToMessage.Text != "" {
			replyText = message.ReplyToMessage.Text
		}

		// Agar rasm bo‘lsa va reply text bo‘sh bo‘lsa
		if replyText == "" && message.ReplyToMessage.Photo != nil && len(message.ReplyToMessage.Photo) > 0 {
			photo := message.ReplyToMessage.Photo[len(message.ReplyToMessage.Photo)-1]
			fileConfig, err := t.TelegramBot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
			if err != nil {
				log.Println("Faylni olishda xatolik:", err)
				return
			}
			link := fileConfig.Link(t.TelegramBot.Token)
			replyText = message.ReplyToMessage.Caption + "\nimage_url: " + link
		}

		// Agar location bo‘lsa
		if message.ReplyToMessage.Location != nil {
			latitude := message.ReplyToMessage.Location.Latitude
			longitude := message.ReplyToMessage.Location.Longitude
			locationURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", latitude, longitude)
			replyText += "\nlocation: " + locationURL
		}

		// Foydalanuvchi reply qilgan message kontentini biriktiramiz
		if replyText != "" {
			userInput += " reply message: " + replyText
		}

		// Agar hozirgi message image bo‘lsa (yangi yuborilgan)
		if message.Photo != nil && len(message.Photo) > 0 {
			photo := message.Photo[len(message.Photo)-1]
			fileConfig, err := t.TelegramBot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
			if err != nil {
				log.Println("Faylni olishda xatolik:", err)
				return
			}
			link := fileConfig.Link(t.TelegramBot.Token)
			userInput += "\nimage_url: " + link
		}
	}

	chatHistory, err := t.Usecase.GetChatHistory(ctx, &entity.GetChatHistoryRequest{
		BusinessID: BussinesId,
		ChatID:     chatID,
		TokenLimit: SettingsAi.ChatToken,
	})
	if err != nil {
		fmt.Println("Error while getting chatHistory:", err)
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}

	err = t.Usecase.CreateChatHistory(ctx, &entity.ChatHistory{
		MessageId:        MessageId,
		BusinessId:       BussinesId,
		PlatformID:       UserId,
		ChatID:           chatID,
		Message:          userInput,
		ReplyToMessageID: replyToMessageID,
		Platform:         "bot",
	})

	if message.Location != nil {
		latitude := message.Location.Latitude
		longitude := message.Location.Longitude
		fmt.Printf("Lokatsiya: Lat: %f, Lng: %f\n", latitude, longitude)

		locationURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", latitude, longitude)
		userInput = locationURL
	}
	Message := entity.SendMessage{
		Message:          userInput,
		UserId:           t.UserId,
		BusinessId:       t.BusinessId,
		From:             "USER",
		Platform:         "BOT",
		Timestamp:        time.Now().String(),
		ReplyToMessageID: replyToMessageID,
		MessageId:        MessageId,
		Chatid:           chatID,
	}
	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type:        "chat",
		ChatMessage: &Message,
	})
	if SettingsAi.IsPauze {
		return
	}

	if SettingsAi.StopTime.String() > time.Now().String() {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID: chatID,
			Message: fmt.Sprintf("Bot vaqtincha to‘xtatilgan. Qayta faollashish vaqti: %s",
				SettingsAi.StopTime.Format("15:04:05 02-Jan-2006")),
		})

		return
	}

	if SettingsAi.IsStop {
		err = t.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID: UserId,
			From:       "bot",
			BusinessId: BussinesId,
			StopStatus: boolPtr(false),
		})
		if err != nil {
			log.Println("Statusni yangilashda xatolik:", err)
		}
	}

	fmt.Println("BussinessID: ", BussinesId)
	state, err := t.RedisUsecase.RedisRepo.Get(ctx, chatID)
	if err != nil {
		fmt.Println("❌ Redis xatosi: " + err.Error())
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}

	if state != nil && state.State != "" {
		t.handleRegistrationInput(message, chatID, state)
		return
	}
	if userInput == "/register" {
		t.startRegistration(chatID)
		return
	}
	if userInput == "/start" {
		t.showWelcomeMessage(chatID)
		return
	}

	chatHistoryStr, err := json.Marshal(chatHistory)
	if err != nil {
		fmt.Println("❌ JSON Marshal xatosi:", err)
		return
	}

	TokenRequest := EstimateTokenCount(userInput + SettingsAi.PromptText + SettingsAi.PromtOrder + string(chatHistoryStr))
	fmt.Println(TokenRequest, SettingsAi.TokenLimit)
	if TokenRequest > SettingsAi.TokenLimit {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "kechirasiz token koplik qildi",
		})
		return
	}
	FirstRes, err := t.GeminiModel.FirstStep(ctx, SettingsAi.IntelligenceLevel, userInput, SettingsAi.PromptText, SettingsAi.PromtOrder, chatHistory)
	if err != nil {
		fmt.Println("❌ AI xatosi: FirstStep: " + err.Error())
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}
	ByteRes, err := json.Marshal(FirstRes)
	if err != nil {
		fmt.Println("error Marshal " + err.Error())
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}
	ResponseToken := EstimateTokenCount(string(ByteRes))
	t.SendTelegramMessage(ctx, entity.SendMessageModel{
		ChatID:  chatID,
		Message: FirstRes.UserMessage,
	})
	if FirstRes.IsAiResponse {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: FirstRes.AiResponse,
		})
		err = t.Usecase.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
			BusinessID:     BussinesId,
			SourceType:     "bot",
			UsedFor:        "text",
			RequestTokens:  TokenRequest,
			ResponseTokens: ResponseToken,
		})
		if err != nil {
			fmt.Println("Error while CreateTokenUsage:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		return
	}

	if FirstRes.IsProductSearch {
		ListProducts, err := t.Usecase.ListProductsForAI(ctx, BussinesId)
		if err != nil {
			fmt.Println("Error while getting ListProductsForAI:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		listProductStr, err := json.Marshal(ListProducts)
		if err != nil {
			fmt.Println("Error while  Marshal:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}

		FoudnProducts, err := t.GeminiModel.ExtractProductName(ctx, userInput, SettingsAi.PromtProdcut, listProductStr, chatHistory)
		if err != nil {
			fmt.Println("Error while getting ExtractProductName:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		ByteRes, err := json.Marshal(FoudnProducts)
		if err != nil {
			fmt.Println("error Marshal " + err.Error())
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		ResponseToken += EstimateTokenCount(string(ByteRes))
		TokenRequest += EstimateTokenCount(userInput + string(listProductStr) + SettingsAi.PromtProdcut + string(chatHistoryStr))
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID: chatID,
		})
		err = t.Usecase.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
			BusinessID:     BussinesId,
			SourceType:     "bot",
			UsedFor:        "text",
			RequestTokens:  TokenRequest,
			ResponseTokens: ResponseToken,
		})
		if err != nil {
			fmt.Println("Error while CreateTokenUsage:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		if FoudnProducts.IsAiResponse {
			t.SendTelegramMessage(ctx, entity.SendMessageModel{
				ChatID:  chatID,
				Message: FoudnProducts.AiResponse,
			})
			return
		}
		statusId, found := FindOrderStatusByName(OrderStatus, "yangi")
		if !found {
			statusId = nil
		}
		var products entity.CreateProductOrder
		order := entity.Order{
			UserID:       message.From.ID,
			Status:       "yangi",
			BusinessId:   BussinesId,
			StatusId:     statusId,
			StatusNumber: 1,
			Platform:     "bot",
		}
		for i := 0; i < len(FoudnProducts.Products); i++ {
			product, err := t.Usecase.GetProductById(ctx, entity.GetProductByIDRequest{
				ProductId:  int64(FoudnProducts.Products[i].Id),
				BusinessID: BussinesId,
			})
			if err != nil {
				fmt.Println("Error while getting GetProductById:", err)
				t.SendErrorTelegramMessage(ctx, chatID)
				return
			}

			t.SendTelegramMessageImages(chatID, FoudnProducts.Products[i].UserMessage, product.Image_urls)
			products.Id = product.ID
			products.Count = 1
			order.ProductOrders = append(order.ProductOrders, products)

		}

		_, err = t.Usecase.CreateOrder(ctx, order)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		err = t.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID:  UserId,
			From:        "bot",
			BusinessId:  BussinesId,
			OrderStatus: order.Status,
		})
		if err != nil {
			fmt.Println("Error while UpdateClientStatus:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		return
	}

	err = t.Usecase.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
		BusinessID:     BussinesId,
		SourceType:     "bot",
		UsedFor:        "text",
		RequestTokens:  TokenRequest,
		ResponseTokens: ResponseToken,
	})
	if err != nil {
		fmt.Println("Error while CreateTokenUsage:", err)
		t.SendErrorTelegramMessage(ctx, chatID)
		return
	}
	if FirstRes.Action == "confirm_order" {
		IsTrue, err := t.Usecase.CheckProductCount(ctx, BussinesId, FirstRes.Products)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		if !IsTrue.Valid {
			t.SendTelegramMessage(ctx, entity.SendMessageModel{
				ChatID:  chatID,
				Message: IsTrue.Message,
			})
			t.SendMessageToAdmin(entity.SendMessageResponse{
				Type: "notification",
				Notifications: &entity.Notification{
					UserId:    t.UserId,
					Title:     "KAM MAHSULOT",
					Content:   IsTrue.Message,
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			})
			return

		}
		statusId, found := FindOrderStatusByName(OrderStatus, "buyurtma_qilmoqchi")
		if !found {
			statusId = nil
		}
		OrderCrear := entity.CreateOrder{
			ID:       FirstRes.Products,
			ChatId:   chatID,
			UserId:   message.From.ID,
			StatusId: statusId,
		}
		err = t.CreateOrder(ctx, &OrderCrear)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}

		return
	}

	if FirstRes.Action == "get_order_status" {
		OrderRes, err := t.Usecase.GetOrderByID(ctx, entity.GetOrderByID{
			OrderID:    FirstRes.OrderID,
			PlatformId: UserId,
			BussnesId:  BussinesId,
		})
		if OrderRes == nil && err == nil {
			t.SendTelegramMessage(ctx, entity.SendMessageModel{
				ChatID:  chatID,
				Message: "sizda buyurtmalar mavjud emas",
			})
			return
		}
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		orders := []*entity.OrderResponseByOrderId{}
		orders = append(orders, OrderRes)

		t.ClientOrders(orders, chatID)
	}

	if FirstRes.Action == "get_order_status_all" {
		OrderRes, err := t.Usecase.GetClientOrders(ctx, UserId, BussinesId)
		if OrderRes == nil && err == nil {
			t.SendTelegramMessage(ctx, entity.SendMessageModel{
				ChatID:  chatID,
				Message: "sizda buyurtmalar mavjud emas",
			})
			return
		}
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		t.ClientOrders(OrderRes, chatID)
	}

	if FirstRes.Action == "confirm_payment" {
		isupdate, err := t.handleOrderStatus(ctx, chatID, UserId, BussinesId, "online_tolov_tasdigi", FirstRes.OrderID, &entity.UpdateOrderRequest{
			ImageUrl:     FirstRes.PaymentScreenshotURL,
			StatusNumber: 4,
		}, OrderStatus)
		if err != nil {
			fmt.Println("Error in handleConfirmPayment:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
		}
		if isupdate {
			err = t.SendNotification("Yangi buyurtma", fmt.Sprintf("Sizda yangi buyurtma bor {id}: %s", FirstRes.OrderID))
			if err != nil {
				fmt.Println("Error while SendNotification", err)
				t.SendErrorTelegramMessage(ctx, chatID)
				return
			}
		}
		return
	}

	if FirstRes.Action == "notification_to_admin" {
		FirstRes.Message += fmt.Sprintf("/nuserid: %s, chatid: %d, messageid: %d", UserId, chatID, MessageId)

		if err := t.SendNotification(FirstRes.Title, FirstRes.Message); err != nil {
			fmt.Println("Error while SendNotification:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}

		return
	}

	if FirstRes.Action == "cancel_order" {
		OrterId, err := strconv.Atoi(FirstRes.OrderID)
		if err != nil {
			fmt.Println("Error while strconv.Atoi(FirstRes.OrderID):", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		statusId, found := FindOrderStatusByName(OrderStatus, "yangi")
		if !found {
			statusId = nil
		}
		message, err := t.Usecase.RestoreProductCounts(ctx, entity.CanseledOrder{
			OrderID:   int64(OrterId),
			NewStatus: "canceled",
			Reason:    FirstRes.Reason,
			StatusId:  statusId,
		})
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: message.Message,
		})
		if message.Isupdate {
			t.SendMessageToAdmin(entity.SendMessageResponse{
				Type: "notification",
				Notifications: &entity.Notification{
					UserId:    t.UserId,
					Title:     "📦 MAHSULOT BEKOR QILINDI",
					Content:   fmt.Sprintf("❗ Foydalanuvchi mahsulotni bekor qildi!\n\n🆔 Buyurtma ID: %d\n📝 Sabab: %s", OrterId, FirstRes.Reason),
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			})

		}

		return
	}

	if FirstRes.Action == "set_payment_method" {
		_, err := t.handleOrderStatus(ctx, chatID, UserId, BussinesId, "tolov_qilmoqchi", FirstRes.OrderID, &entity.UpdateOrderRequest{
			PaymentMethod: FirstRes.Method,
			StatusNumber:  3,
		}, OrderStatus)
		if err != nil {
			fmt.Println("Error in handleConfirmPayment:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
		}

		return
	}

	if FirstRes.Action == "set_order_location" {

		OrderIdStr, err := strconv.Atoi(FirstRes.OrderID)
		if err != nil {
			fmt.Println("Error while strconv.Atoi(FirstRes.OrderID):", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		updateMessage, err := t.Usecase.UpdateOrderStatus(ctx, &entity.UpdateOrderRequest{
			OrderID:     int64(OrderIdStr),
			LocationUrl: FirstRes.LocationUrl,
			Location:    FirstRes.Location,
			UserNote:    FirstRes.UserNote,
		})
		if err != nil {
			fmt.Println("Error while UpdateOrderStatus:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		err = t.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID:   UserId,
			From:         "bot",
			BusinessId:   BussinesId,
			Location:     FirstRes.LocationUrl,
			LocationText: FirstRes.Location,
		})
		if err != nil {
			fmt.Println("Error while UpdateClientStatus:", err)
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: updateMessage.Message,
		})
		return
	}

}

func (t *Handler) handleRegistrationInput(message *tgbotapi.Message, chatID int64, state *entity.ClientState) {
	ctx := context.TODO()

	switch state.State {
	case "first_name":
		state.ClientDetails.FirstName = message.Text
		state.State = "phone"
		t.RedisUsecase.RedisRepo.Set(ctx, chatID, state)
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "phone number kiriting:",
		})

	case "phone":
		state.ClientDetails.Phone = message.Text
		state.ClientDetails.PlatformID = fmt.Sprintf("%d", message.From.ID)
		state.ClientDetails.ChatId = chatID
		state.ClientDetails.BusinessId = t.BusinessId
		state.ClientDetails.UserName = message.From.UserName
		state.ClientDetails.From = "bot"
		res, err := t.Usecase.CreateClient(ctx, state.ClientDetails)
		if err != nil {
			fmt.Println("❌ Mijozni ro'yxatga olishda xatolik: " + err.Error())
			t.SendErrorTelegramMessage(ctx, chatID)
			return
		}

		err = t.RedisUsecase.RedisRepo.Delete(ctx, chatID)
		if err != nil {
			fmt.Println("❌ Redisda state o'chirishda xatolik: " + err.Error())
			return
		}

		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: fmt.Sprintf("✅ Ro'yxatdan o'tdingiz, %s!", res.Username),
		})
	}
}

func AddNewBot(req entity.BotIntegration) error {
	tgBot, err := NewHandler(&config.Config{}, req.Token, req.BusinessID, req.UserID, newGeminiModel, translationUseCase, redisUscase)
	if err != nil {
		log.Printf("Bot yaratishda xatolik (BusinessID: %s): %v", req.BusinessID, err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	tgBot.cancelFunc = cancel
	AddBotMap(req.BusinessID, tgBot)

	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates := tgBot.TelegramBot.GetUpdatesChan(u)

		for {
			select {
			case <-ctx.Done():
				log.Printf("Bot %s stopped by context", req.BusinessID)
				return
			case update := <-updates:
				if update.Message != nil {
					go tgBot.HandleTelegramMessage(update.Message)
				}
			}
		}
	}()

	return nil
}

func StopBot(guid string) error {
	botHandler, ok := BotMap[guid]
	if !ok {
		return fmt.Errorf("Bot not found")
	}

	if botHandler.cancelFunc != nil {
		botHandler.cancelFunc()
	}

	botHandler.TelegramBot.StopReceivingUpdates()

	delete(BotMap, guid)
	log.Printf("Bot stopped: %s", guid)
	return nil
}

func CreateGlobalVar(geminiModel *gemini.Gemini, UseCase *uscase.UseCase, redisUs *uscaseredis.Uscase) {
	newGeminiModel = geminiModel
	translationUseCase = UseCase
	redisUscase = redisUs
}

func AddBotMap(guid string, tgBot *Handler) {
	BotMap[guid] = tgBot
}

func (h *Handler) SetCancelFunc(cancel context.CancelFunc) {
	h.cancelFunc = cancel
}

func (t *Handler) SendErrorTelegramMessage(ctx context.Context, chatid int64) {

	msg := tgbotapi.NewMessage(chatid, errorMEssage)

	sentMessage, err := t.TelegramBot.Send(msg)
	if err != nil {
		fmt.Println("Telegram send error:", err)
		return
	}

	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "chat",
		ChatMessage: &entity.SendMessage{
			AIResponse: errorMEssage,
			UserId:     t.UserId,
			BusinessId: t.BusinessId,
			From:       "AI",
			Platform:   "BOT",
			Timestamp:  time.Now().Format(time.RFC3339),
			MessageId:  sentMessage.MessageID,
		},
	})

	history := &entity.ChatHistory{
		MessageId:  sentMessage.MessageID,
		BusinessId: t.BusinessId,
		ChatID:     chatid,
		AIResponse: errorMEssage,
	}
	if err := t.Usecase.CreateChatHistory(ctx, history); err != nil {
		fmt.Println("Error while creating chat history:", err)

		errorMsg := tgbotapi.NewMessage(chatid, errorMEssage)
		_, sendErr := t.TelegramBot.Send(errorMsg)
		if sendErr != nil {
			fmt.Println("Failed to send error message:", sendErr)
			return
		}

	}
}
