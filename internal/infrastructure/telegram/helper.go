package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/internal/entity"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (t *Handler) showWelcomeMessage(ctx context.Context, message *tgbotapi.Message, clientID, businessID string) {
	welcomeMsg := tgbotapi.NewMessage(message.Chat.ID, "Assalomu alaykum, botga xush kelibsiz.")
	if _, err := t.TelegramBot.Send(welcomeMsg); err != nil {
		fmt.Println("error sending welcome message: " + err.Error())
		t.SendErrorTelegramMessage(ctx, message.Chat.ID)
		return
	}

	check, err := t.Usecase.CheckClient(ctx, clientID, businessID)
	if err != nil {
		fmt.Println("error CheckClient: " + err.Error())
		t.SendErrorTelegramMessage(ctx, message.Chat.ID)
		return
	}

	if !check {
		client := entity.Client{
			PlatformID: clientID,
			BusinessId: businessID,
			ChatId:     message.Chat.ID,
			UserName:   message.From.UserName,
			FirstName:  message.From.FirstName,
		}

		if _, err := t.Usecase.CreateClient(ctx, client); err != nil {
			fmt.Println("error CreateClient: " + err.Error())
			t.SendErrorTelegramMessage(ctx, message.Chat.ID)
			return
		}
	}
}

func (t *Handler) SendTelegramMessage(ctx context.Context, data entity.SendMessageModel) {
	msg := tgbotapi.NewMessage(data.ChatID, data.Message)

	if data.ReplyToMessageID != nil {
		msg.ReplyToMessageID = *data.ReplyToMessageID
	}

	sentMessage, err := t.TelegramBot.Send(msg)
	if err != nil {
		fmt.Println("Telegram send error:", err)
		return
	}

	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "chat",
		ChatMessage: &entity.SendMessage{
			AIResponse:       data.Message,
			UserId:           t.UserId,
			BusinessId:       t.BusinessId,
			From:             "AI",
			Platform:         "BOT",
			Timestamp:        time.Now().Format(time.RFC3339),
			MessageId:        sentMessage.MessageID,
			ReplyToMessageID: safeDereference(data.ReplyToMessageID),
			Chatid:           data.ChatID,
		},
	})

	history := &entity.ChatHistory{
		MessageId:        sentMessage.MessageID,
		BusinessId:       t.BusinessId,
		ChatID:           data.ChatID,
		AIResponse:       data.Message,
		ReplyToMessageID: safeDereference(data.ReplyToMessageID),
		Platform:         "bot",
	}
	if err := t.Usecase.CreateChatHistory(ctx, history); err != nil {
		fmt.Println("Error while creating chat history:", err)
		t.SendMessageToAdmin(entity.SendMessageResponse{
			Type: "chat",
			ChatMessage: &entity.SendMessage{
				AIResponse:       "❌ Nimadur xato bo‘ldi. Iltimos, qaytadan urinib ko‘ring yoki admin bilan bog‘laning.",
				UserId:           t.UserId,
				BusinessId:       t.BusinessId,
				From:             "AI",
				Platform:         "BOT",
				Timestamp:        time.Now().Format(time.RFC3339),
				MessageId:        sentMessage.MessageID,
				ReplyToMessageID: safeDereference(data.ReplyToMessageID),
				Chatid:           data.ChatID,
			},
		})
		errorMsg := tgbotapi.NewMessage(data.ChatID, "❌ Nimadur xato bo‘ldi. Iltimos, qaytadan urinib ko‘ring yoki admin bilan bog‘laning.")
		_, sendErr := t.TelegramBot.Send(errorMsg)
		if sendErr != nil {
			fmt.Println("Failed to send error message:", sendErr)
			return
		}

	}
}

func safeDereference(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

func (t *Handler) SendTelegramMessageImages(chatID int64, message string, imageURLs []string) {
	var mediaGroup []interface{}

	for i, url := range imageURLs {
		photo := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(url))
		// Faqat birinchi rasmga caption qo‘shamiz
		if i == 0 && message != "" {
			photo.Caption = message
			photo.ParseMode = "Markdown" // yoki "HTML"
		}
		mediaGroup = append(mediaGroup, photo)
	}

	if len(mediaGroup) > 0 {
		mediaConfig := tgbotapi.MediaGroupConfig{
			ChatID: chatID,
			Media:  mediaGroup,
		}
		_, err := t.TelegramBot.SendMediaGroup(mediaConfig)
		if err != nil {
			log.Printf("SendMediaGroup error: %v", err)
		}
	} else {
		// Agar rasm bo'lmasa, faqat xabar yuboriladi
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		t.TelegramBot.Send(msg)
	}

	// Adminlarga xabar yuborish
	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "chat",
		ChatMessage: &entity.SendMessage{
			AIResponse: message,
			UserId:     t.UserId,
			BusinessId: t.BusinessId,
			From:       "AI",
			Platform:   "BOT",
			Timestamp:  time.Now().String(),
		},
	})
}

func SendTelegramNotification(guid, productId string) error {
	botHandler, ok := BotMap[guid]
	if !ok {
		return fmt.Errorf("Bot not found")
	}

	product, err := botHandler.Usecase.GetProductInfoForNotification(context.Background(), productId)
	if err != nil {
		return fmt.Errorf("failed to get product info: %w", err)
	}

	chatIDs, err := botHandler.Usecase.GetAllChatId(context.Background(), guid)
	if err != nil {
		return fmt.Errorf("failed to get chat ids: %w", err)
	}

	message := fmt.Sprintf(
		"🛍 *%s*\n\n📋 %s\n\n💬 %s\n\n💰 *Narxi:* ~%d so'm~\n🔥 *Chegirma:* %d%%\n✅ *Chegirmali narx:* %d so'm",
		product.Name,
		product.ShortInfo,
		product.Description,
		product.Cost,
		product.Discount,
		product.DiscountCost,
	)

	for _, chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown" // Markdown orqali bold, italic, etc.
		_, err := botHandler.TelegramBot.Send(msg)
		if err != nil {
			log.Printf("Error sending message to %d: %v", chatID, err)
		}
	}

	log.Printf("Bot sent notification to %d users: %s", len(chatIDs), guid)
	return nil
}

func (t *Handler) ShowMenu(chatID int64) {
	ctx := context.Background()

	menus, err := t.Usecase.Business.GetAllMenusByOwnerID(ctx, t.BusinessId)
	if err != nil {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "❌ Menyularni olishda xatolik: " + err.Error(),
		})
		return
	}

	if len(menus) == 0 {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "🚫 Hech qanday menyu topilmadi.",
		})
		return
	}

	msg := "📋 *Mavjud menyular:*\n\n"
	for _, menu := range menus {
		msg += fmt.Sprintf("➡️ `%s`\n", menu.Command)
	}
	msg += "\nQuyidagilardan birini yuboring."

	t.SendTelegramMessage(ctx, entity.SendMessageModel{
		ChatID:  chatID,
		Message: msg,
	})
}

func (t *Handler) HandleMenuCommand(chatID int64, command string) {
	ctx := context.Background()

	response, err := t.Usecase.AuthRepo.GetResponseByCommand(ctx, t.BusinessId, command)
	if err != nil {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "❌ Komandani olishda xatolik: " + err.Error(),
		})
		return
	}

	if response == "" {
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  chatID,
			Message: "🚫 Bunday menyu mavjud emas.",
		})
		return
	}

	t.SendTelegramMessage(ctx, entity.SendMessageModel{
		ChatID:  chatID,
		Message: response,
	})
}

func (t *Handler) SendMessageToAdmin(chat entity.SendMessageResponse) {
	body, err := json.Marshal(chat)
	if err != nil {
		fmt.Println("Error while marshaling chat message:", err)
		return
	}
	// http://ai-seller-admin:8080/v1/websocket/chat/send-message
	// http://localhost:8080/v1/websocket/chat/send-message
	resp, err := http.Post("http://ai-seller-admin:8080/v1/websocket/chat/send-message", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error while sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	var botResp entity.IntegrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&botResp); err != nil {
		fmt.Println("Error while decoding response:", err)
		return
	}

	if botResp.Code != 0 {
		fmt.Println("Bot returned error code:", botResp.Code)
		return
	}
}

func SendMessageBot(ctx context.Context, req entity.SendMessage) error {
	botHandler, ok := BotMap[req.BusinessId]
	if !ok {
		return fmt.Errorf("bot not found for business id: %s", req.BusinessId)
	}

	msg := tgbotapi.NewMessage(req.Chatid, req.Message)

	if req.ReplyToMessageID > 0 {
		msg.ReplyToMessageID = req.ReplyToMessageID
	}

	sentMessage, err := botHandler.TelegramBot.Send(msg)
	if err != nil {
		log.Printf("Telegram send error: %v", err)
		return err
	}

	history := &entity.ChatHistory{
		MessageId:        sentMessage.MessageID,
		BusinessId:       req.BusinessId,
		ChatID:           req.Chatid,
		AIResponse:       req.Message,
		ReplyToMessageID: req.ReplyToMessageID,
		Platform:         "bot",
	}

	if err := botHandler.Usecase.CreateChatHistory(ctx, history); err != nil {
		log.Printf("Error while creating chat history: %v", err)

		errorMsg := tgbotapi.NewMessage(req.Chatid, "❌ Nimadur xato bo‘ldi. Iltimos, qaytadan urinib ko‘ring yoki admin bilan bog‘laning.")
		if _, sendErr := botHandler.TelegramBot.Send(errorMsg); sendErr != nil {
			log.Printf("Failed to send fallback error message: %v", sendErr)
		}
	}
	settings, err := botHandler.Usecase.GetIntegrationSettingsByOwnerID(ctx, req.BusinessId, strconv.Itoa(int(req.Chatid)))
	if err != nil {
		log.Printf("Telegram GetIntegrationSettingsByOwnerID: %v", err)
		return err
	}
	if settings.IsPauze {
		err = botHandler.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID: strconv.Itoa(int(req.Chatid)),
			BusinessId: req.BusinessId,
			From:       "bot",
			IsPauzse:   boolPtr(true),
		})
		if err != nil {
			log.Printf("Telegram UpdateClientStatus: %v", err)
			botHandler.SendErrorTelegramMessage(ctx, req.Chatid)
			return err
		}
		log.Printf("Bot sent message: %s", req.Message)
		return nil
	}
	if !settings.IsStop {
		stopUntilDuration := time.Duration(settings.StopUntil) * time.Hour
		stopTime := time.Now().Add(stopUntilDuration)

		err = botHandler.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID: strconv.Itoa(int(req.Chatid)),
			BusinessId: req.BusinessId,
			From:       "bot",
			StopStatus: boolPtr(true),
			StopTime:   &stopTime,
		})
		if err != nil {
			log.Printf("Telegram UpdateClientStatus: %v", err)
			botHandler.SendErrorTelegramMessage(ctx, req.Chatid)
			return err
		}
	}

	log.Printf("Bot sent message: %s", req.Message)
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (t *Handler) AskForLocation(chatID int64) {
	locationButton := tgbotapi.KeyboardButton{
		Text:            "📍 Lokatsiyani yuborish",
		RequestLocation: true,
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(locationButton),
	)

	msg := tgbotapi.NewMessage(chatID, "Iltimos, lokatsiyangizni yuboring:")
	msg.ReplyMarkup = keyboard

	if _, err := t.TelegramBot.Send(msg); err != nil {
		log.Println("Lokatsiya so‘rovini yuborishda xatolik:", err)
		return
	}
}

func EstimateTokenCount(prompt string) int {
	words := strings.Fields(prompt)
	estimatedTokens := float64(len(words)) * 1.5
	return int(estimatedTokens)
}

func FindOrderStatusByName(list []*entity.OrderStatus, name string) (*string, bool) {
	for _, status := range list {
		if status.Type.Name == name {
			return &status.GUID, true
		}
	}
	return nil, false
}
