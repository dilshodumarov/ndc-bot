package telegramuser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/infrastructure/telegram"
	"strconv"
	"strings"
	"time"
)

var (
	ctx                = context.Background()
	errorMEssage       string
	BusinessId, UserId string
	Chatid             int64
)

func (t *Handler) AiResponse(msg entity.Messaging) {

	BussinessInfo, err := t.ProductUscse.GetBusinessByPhone(ctx, entity.GetBussinesId{Phone: msg.Sender.ID})
	if err != nil || BussinessInfo.BusinessID == "" {
		log.Println("GetBusinessByPhone error:", err)
		return
	}
	chatID, err := strconv.Atoi(msg.Sender.ID)
	if err != nil {
		log.Println("strconv.Atoi error:", err)
		t.sendErrorMessage(msg)
		return
	}
	Chatid = int64(chatID)
	BusinessId = BussinessInfo.BusinessID
	UserId = BussinessInfo.OwnerID
	text := msg.Message.Text
	Message := entity.SendMessage{
		Message:    text,
		UserId:     UserId,
		BusinessId: BusinessId,
		From:       "USER",
		Platform:   "TELEGRAM",
		Timestamp:  time.Now().String(),
		//ReplyToMessageID: *msg.ReplyToMessageID,
		//MessageId: msg.MessageID,
		Chatid: Chatid,
	}
	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type:        "chat",
		ChatMessage: &Message,
	})
	OrderStatus, err := t.ProductUscse.GetOrderStatusesByBusinessID(ctx, BusinessId)
	if err != nil {
		fmt.Println("Error while GetOrderStatusesByBusinessID:", err)
		t.sendErrorMessage(msg)
		return
	}
	SettingsAi, err := t.ProductUscse.GetIntegrationSettingsByOwnerID(ctx, BusinessId, msg.Recipient.ID)
	if err != nil {
		fmt.Println("Error while GetIntegrationSettingsByOwnerID:", err)
		t.sendErrorMessage(msg)
		return
	}

	errorMEssage = SettingsAi.ErrorMessage
	fmt.Println(1111, SettingsAi.PromptText)
	if SettingsAi.IsBlocked {
		t.SendInstagramMessage(msg, "siz bloklangansiz")
	}
	history, err := t.ProductUscse.GetChatHistory(ctx, &entity.GetChatHistoryRequest{
		BusinessID: BusinessId,
		ChatID:     int64(chatID),
		TokenLimit: 100,
	})
	if err != nil {
		log.Println("GetChatHistory error:", err)
		t.sendErrorMessage(msg)
		return
	}

	err = t.ProductUscse.CreateChatHistory(ctx, &entity.ChatHistory{
		//MessageId:  msg.MessageID,
		BusinessId: BusinessId,
		PlatformID: msg.Recipient.ID,
		ChatID:     int64(chatID),
		Message:    text,
		//ReplyToMessageID: *msg.ReplyToMessageID,
		Platform: "instagram",
	})
	if err != nil {
		log.Println("CreateChatHistory error:", err)
		t.sendErrorMessage(msg)
		return
	}
	chatHistoryStr, err := json.Marshal(history)
	if err != nil {
		fmt.Println("❌ JSON Marshal xatosi:", err)
		return
	}

	TokenRequest := EstimateTokenCount(text + SettingsAi.PromptText + SettingsAi.PromtOrder + string(chatHistoryStr))
	fmt.Println(TokenRequest, SettingsAi.TokenLimit)
	if TokenRequest > SettingsAi.TokenLimit {

		t.SendInstagramMessage(msg, "kechirasiz token koplik qildi")
		return
	}
	firstRes, err := t.GeminiModel.FirstStep(ctx, SettingsAi.IntelligenceLevel, text, SettingsAi.PromptText, SettingsAi.PromtOrder, history)
	if err != nil {
		log.Println("FirstStep error:", err)
		return
	}
	ByteRes, err := json.Marshal(firstRes)
	if err != nil {
		fmt.Println("error Marshal " + err.Error())
		t.sendErrorMessage(msg)
		return
	}
	ResponseToken := EstimateTokenCount(string(ByteRes))

	if firstRes.IsAiResponse {
		t.SendInstagramMessage(msg, firstRes.AiResponse)
		err = t.ProductUscse.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
			BusinessID:     BusinessId,
			SourceType:     "telegram",
			UsedFor:        "text",
			RequestTokens:  TokenRequest,
			ResponseTokens: ResponseToken,
		})
		if err != nil {
			fmt.Println("Error while CreateTokenUsage:", err)
			t.sendErrorMessage(msg)
			return
		}
		return
	}
	// if firstRes.UserMessage!=""{
	// 	t.SendInstagramMessage(msg, firstRes.UserMessage)
	// }

	if firstRes.IsProductSearch {
		list, err := t.ProductUscse.ListProductsForAI(ctx, BusinessId)
		if err != nil {
			log.Println("ListProductsForAI error:", err)
			t.sendErrorMessage(msg)
			return
		}
		listProductStr, err := json.Marshal(list)
		if err != nil {
			log.Println("ListProductsForAI error:", err)
			t.sendErrorMessage(msg)
			return
		}
		FoudnProducts, err := t.GeminiModel.ExtractProductName(ctx, text, SettingsAi.PromtProdcut, listProductStr, history)
		if err != nil {
			log.Println("ExtractProductName error:", err)
			t.sendErrorMessage(msg)
			return
		}
		ByteRes, err := json.Marshal(FoudnProducts)
		if err != nil {
			fmt.Println("error Marshal " + err.Error())
			t.sendErrorMessage(msg)
			return
		}
		ResponseToken += EstimateTokenCount(string(ByteRes))
		TokenRequest += EstimateTokenCount(text + string(listProductStr) + SettingsAi.PromtProdcut + string(chatHistoryStr))
		//t.SendInstagramMessage(msg, FoudnProducts.)
		err = t.ProductUscse.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
			BusinessID:     BusinessId,
			SourceType:     "telegram",
			UsedFor:        "text",
			RequestTokens:  TokenRequest,
			ResponseTokens: ResponseToken,
		})
		if err != nil {
			fmt.Println("Error while CreateTokenUsage:", err)
			t.sendErrorMessage(msg)
			return
		}
		if FoudnProducts.IsAiResponse {
			t.SendInstagramMessage(msg, FoudnProducts.AiResponse)
			return
		}
		statusId, found := telegram.FindOrderStatusByName(OrderStatus, "yangi")
		if !found {
			statusId = nil
		}
		var products entity.CreateProductOrder
		order := entity.Order{
			UserID:       Chatid,
			Status:       "yangi",
			BusinessId:   BusinessId,
			StatusId:     statusId,
			StatusNumber: 1,
			Platform:     "telegram",
		}
		for i := 0; i < len(FoudnProducts.Products); i++ {
			product, err := t.ProductUscse.GetProductById(ctx, entity.GetProductByIDRequest{
				ProductId:  int64(FoudnProducts.Products[i].Id),
				BusinessID: BusinessId,
			})
			if err != nil {
				fmt.Println("Error while getting GetProductById:", err)
				t.sendErrorMessage(msg)
				return
			}

			t.SendImageMessage(entity.ImageMessageRequest{
				//Phone:     msg.Phone,
				UserID:    chatID,
				Message:   FoudnProducts.Products[i].UserMessage,
				ImageURLs: product.Image_urls,
			})
			t.SendMessageToAdmin(entity.SendMessageResponse{
				Type: "chat",
				ChatMessage: &entity.SendMessage{
					AIResponse: FoudnProducts.Products[i].UserMessage,
					UserId:     UserId,
					BusinessId: BusinessId,
					From:       "AI",
					Platform:   "TELEGRAM",
					Timestamp:  time.Now().String(),
				},
			})
			products.Id = product.ID
			products.Count = 1
			order.ProductOrders = append(order.ProductOrders, products)

		}

		_, err = t.ProductUscse.CreateOrder(ctx, order)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			t.sendErrorMessage(msg)
			return
		}
		err = t.ProductUscse.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID:  msg.Recipient.ID,
			From:        "instagram",
			BusinessId:  BusinessId,
			OrderStatus: order.Status,
		})
		if err != nil {
			fmt.Println("Error while UpdateClientStatus:", err)
			t.sendErrorMessage(msg)
			return
		}
		return

	}

	err = t.ProductUscse.CreateTokenUsage(ctx, &entity.ClientTokenUsage{
		BusinessID:     BusinessId,
		SourceType:     "telegram",
		UsedFor:        "text",
		RequestTokens:  TokenRequest,
		ResponseTokens: ResponseToken,
	})
	if err != nil {
		fmt.Println("Error while CreateTokenUsage:", err)
		t.sendErrorMessage(msg)
		return
	}
	if firstRes.Action == "confirm_order" {
		IsTrue, err := t.ProductUscse.CheckProductCount(ctx, BusinessId, firstRes.Products)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			t.sendErrorMessage(msg)
			return
		}
		if !IsTrue.Valid {
			t.SendInstagramMessage(msg, IsTrue.Message)
			t.SendMessageToAdmin(entity.SendMessageResponse{
				Type: "notification",
				Notifications: &entity.Notification{
					UserId:    UserId,
					Title:     "KAM MAHSULOT",
					Content:   IsTrue.Message,
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			})
			return

		}
		statusId, found := telegram.FindOrderStatusByName(OrderStatus, "buyurtma_qilmoqchi")
		if !found {
			statusId = nil
		}
		OrderCrear := entity.CreateOrder{
			ID:       firstRes.Products,
			ChatId:   Chatid,
			UserId:   Chatid,
			StatusId: statusId,
		}
		err = t.CreateOrder(ctx, &OrderCrear, msg)
		if err != nil {
			fmt.Println("Error while CreateOrder:", err)
			return
		}

		return
	}
	if firstRes.Action == "get_order_status" {
		OrderRes, err := t.ProductUscse.GetOrderByID(ctx, entity.GetOrderByID{})
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.sendErrorMessage(msg)
			return
		}
		orders := []*entity.OrderResponseByOrderId{}
		orders = append(orders, OrderRes)

		t.ClientOrders(orders, msg)
	}

	if firstRes.Action == "get_order_status_all" {
		OrderRes, err := t.ProductUscse.GetClientOrders(ctx, UserId, BusinessId)
		if OrderRes == nil && err == nil {
			t.SendInstagramMessage(msg, "sizda buyurtmalar mavjud emas")
			return
		}
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.sendErrorMessage(msg)
			return
		}
		t.ClientOrders(OrderRes, msg)
	}

	if firstRes.Action == "confirm_payment" {
		isupdate, err := t.handleOrderStatus(ctx, Chatid, UserId, BusinessId, "online_tolov_tasdigi", firstRes.OrderID, &entity.UpdateOrderRequest{
			ImageUrl:     firstRes.PaymentScreenshotURL,
			StatusNumber: 4,
		}, OrderStatus, msg)
		if err != nil {
			fmt.Println("Error in handleConfirmPayment:", err)
			t.sendErrorMessage(msg)
		}
		if isupdate {
			err = t.SendNotification("Yangi buyurtma", fmt.Sprintf("Sizda yangi buyurtma bor {id}: %s", firstRes.OrderID))
			if err != nil {
				fmt.Println("Error while SendNotification", err)
				t.sendErrorMessage(msg)
				return
			}
		}
		return
	}

	if firstRes.Action == "notification_to_admin" {
		firstRes.Message += fmt.Sprintf("/nuserid: %s, chatid: %d", UserId, chatID)

		if err := t.SendNotification(firstRes.Title, firstRes.Message); err != nil {
			fmt.Println("Error while SendNotification:", err)
			t.sendErrorMessage(msg)
			return
		}

		return
	}

	if firstRes.Action == "cancel_order" {
		OrterId, err := strconv.Atoi(firstRes.OrderID)
		if err != nil {
			fmt.Println("Error while strconv.Atoi(firstRes.OrderID):", err)
			t.sendErrorMessage(msg)
			return
		}
		statusId, found := telegram.FindOrderStatusByName(OrderStatus, "yangi")
		if !found {
			statusId = nil
		}
		message, err := t.ProductUscse.RestoreProductCounts(ctx, entity.CanseledOrder{
			OrderID:   int64(OrterId),
			NewStatus: "canceled",
			Reason:    firstRes.Reason,
			StatusId:  statusId,
		})
		if err != nil {
			fmt.Println("Error while GetOrderByID:", err)
			t.sendErrorMessage(msg)
			return
		}
		t.SendInstagramMessage(msg, message.Message)
		if message.Isupdate {
			t.SendMessageToAdmin(entity.SendMessageResponse{
				Type: "notification",
				Notifications: &entity.Notification{
					UserId:    UserId,
					Title:     "📦 MAHSULOT BEKOR QILINDI",
					Content:   fmt.Sprintf("❗ Foydalanuvchi mahsulotni bekor qildi!\n\n🆔 Buyurtma ID: %d\n📝 Sabab: %s", OrterId, firstRes.Reason),
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			})

		}

		return
	}

	if firstRes.Action == "set_payment_method" {
		_, err := t.handleOrderStatus(ctx, Chatid, UserId, BusinessId, "tolov_qilmoqchi", firstRes.OrderID, &entity.UpdateOrderRequest{
			PaymentMethod: firstRes.Method,
			StatusNumber:  3,
		}, OrderStatus, msg)
		if err != nil {
			fmt.Println("Error in handleConfirmPayment:", err)
			t.sendErrorMessage(msg)
		}

		return
	}

	if firstRes.Action == "set_order_location" {
		OrderIdStr, err := strconv.Atoi(firstRes.OrderID)
		if err != nil {
			fmt.Println("Error while strconv.Atoi(firstRes.OrderID):", err)
			t.sendErrorMessage(msg)
			return
		}
		updateMessage, err := t.ProductUscse.UpdateOrderStatus(ctx, &entity.UpdateOrderRequest{
			OrderID:     int64(OrderIdStr),
			LocationUrl: firstRes.LocationUrl,
			Location:    firstRes.Location,
			UserNote:    firstRes.UserNote,
		})
		if err != nil {
			fmt.Println("Error while UpdateOrderStatus:", err)
			t.sendErrorMessage(msg)
			return
		}
		err = t.ProductUscse.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
			PlatformID:  UserId,
			From:        "telegram",
			BusinessId:  BusinessId,
			OrderStatus: "yetkazish_kerak",
		})
		if err != nil {
			fmt.Println("Error while UpdateClientStatus:", err)
			t.sendErrorMessage(msg)
			return
		}
		t.SendInstagramMessage(msg, updateMessage.Message)
		return
	}

}

func (t *Handler) sendErrorMessage(msg entity.Messaging) {

	Resp, err := t.SendInstagramMessage(msg, errorMEssage)
	if err != nil {
		fmt.Println("Error while SendInstagramMessage:", err)
		return
	}
	if Resp.Code != 0 {
		fmt.Println("Error while SendInstagramMessage::", Resp.Message)
		return
	}
	err = t.ProductUscse.CreateChatHistory(ctx, &entity.ChatHistory{
		//MessageId:  Resp.Message_id,
		BusinessId: BusinessId,
		ChatID:     Chatid,
		AIResponse: errorMEssage,
		Platform:   "instagram",
	})
	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "chat",
		ChatMessage: &entity.SendMessage{
			AIResponse: errorMEssage,
			UserId:     UserId,
			BusinessId: BusinessId,
			From:       "AI",
			Platform:   "instagram",
			Timestamp:  time.Now().Format(time.RFC3339),
			//MessageId:  Resp.Message_id,
			Chatid:     Chatid,
		},
	})
	if err != nil {
		log.Println("CreateChatHistory error:", err)
		return
	}

}

// func (t *Handler) Sendtext(msg entity.MessageResponse, text string) {
// 	Code, err := t.SendTelegramMessage(entity.MessageRequest{
// 		Phone:  msg.Phone,
// 		UserID: msg.FromID,
// 		Text:   text,
// 	})

// 	if err != nil {
// 		log.Println("SendTelegramMessage error:", err)
// 		t.sendErrorMessage(msg)
// 		return
// 	}
// 	if Code.Code != 0 {
// 		log.Println("SendTelegramMessage error:", Code.Message)
// 		t.sendErrorMessage(msg)
// 		return
// 	}
// 	t.SendMessageToAdmin(entity.SendMessageResponse{
// 		Type: "chat",
// 		ChatMessage: &entity.SendMessage{
// 			AIResponse:       text,
// 			UserId:           UserId,
// 			BusinessId:       BusinessId,
// 			From:             "AI",
// 			Platform:         "TELEGRAM",
// 			Timestamp:        time.Now().Format(time.RFC3339),
// 			MessageId:        Code.Message_id,
// 			ReplyToMessageID: safeDereference(msg.ReplyToMessageID),
// 			Chatid:           Chatid,
// 		},
// 	})
// 	err = t.ProductUscse.CreateChatHistory(ctx, &entity.ChatHistory{
// 		MessageId:  Code.Message_id,
// 		BusinessId: BusinessId,
// 		PlatformID: msg.FromID,
// 		ChatID:     Chatid,
// 		AIResponse: text,
// 		Phone:      msg.Phone,
// 		Platform:   "telegram",
// 	})
// 	if err != nil {
// 		log.Println("CreateChatHistory error:", err)

// 		return
// 	}

// }

func (t *Handler) ClientOrders(orders []*entity.OrderResponseByOrderId, msg entity.Messaging) {
	var builder strings.Builder
	if len(orders) == 0 {
		builder.WriteString("❌ Sizda hali hech qanday buyurtma mavjud emas.")
	} else {
		builder.WriteString("📦 Buyurtmalaringiz:\n\n")
		for i, order := range orders {
			builder.WriteString(fmt.Sprintf("📌 #%d - Holat: %s\n🕒 Yaratilgan: %s\n", i+1, order.Status, order.CreatedAt.Format("2006-01-02 15:04")))

			for _, p := range order.Products {
				builder.WriteString(fmt.Sprintf("  🛒 %s x%d\n", p.ProductName, p.Count))
			}
			builder.WriteString("\n")
		}
	}
	t.SendInstagramMessage(msg, builder.String())

}

func safeDereference(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

func (t *Handler) CreateOrder(ctx context.Context, res *entity.CreateOrder, msg entity.Messaging) error {
	var products entity.CreateProductOrder
	order := entity.Order{
		UserID:       res.UserId,
		Status:       "olishga_tayyor",
		StatusId:     res.StatusId,
		StatusNumber: 2,
		Platform:     "telegram",
	}

	for i := 0; i < len(res.ID); i++ {

		product, err := t.ProductUscse.GetProductById(ctx, entity.GetProductByIDRequest{
			ProductId:  int64(res.ID[i].ProductID),
			BusinessID: BusinessId,
		})
		if err != nil || product == nil {
			t.SendInstagramMessage(msg, fmt.Sprintf("❌ Mahsulot topilmadi (ID: %d).", res.ID[i]))
			return err
		}
		products.Id = product.ID
		products.Count = res.ID[i].Count
		products.TotalPrice = res.ID[i].Count * product.Cost
		products.Price = product.Cost
		order.TotalPrice += res.ID[i].Count * product.Cost
		order.BusinessId = BusinessId
		order.ProductOrders = append(order.ProductOrders, products)
	}

	orderResponse, err := t.ProductUscse.CreateOrder(ctx, order)
	if err != nil {
		fmt.Println("❌ Buyurtma yaratishda xatolik: " + err.Error())
		t.sendErrorMessage(msg)
		return err
	}

	if orderResponse.Message == "norows" {
		state := &entity.ClientState{
			State: "first_name",
		}

		err := t.RedisUsecase.RedisRepo.Set(ctx, res.ChatId, state)
		if err != nil {
			fmt.Println("❌ Redisda state saqlashda xatolik: " + err.Error())
			t.sendErrorMessage(msg)
			return err
		}
		t.SendInstagramMessage(msg, "Ismingizni kiriting:")
		return err
	}
	platformID := fmt.Sprintf("%d", order.UserID)
	err = t.ProductUscse.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
		PlatformID:  platformID,
		From:        "telegram",
		BusinessId:  order.BusinessId,
		Goal:        "sotipOlmoqchoi",
		OrderStatus: order.Status,
	})
	if err != nil {
		fmt.Println("Update statusda hatolik: " + err.Error())
		t.sendErrorMessage(msg)
		return err
	}
	response := fmt.Sprintf(
		"✅ Buyurtmangiz muvaffaqiyatli rasmiylashtirildi!\n\n🆔 Buyurtma Id: #%d\n💵 Umumiy narx: %d so'm\n\nEndi locatsiangizni yuboring.",
		orderResponse.OrderIdSerial,
		order.TotalPrice,
	)
	t.SendInstagramMessage(msg, response)

	return nil

}

func (t *Handler) handleOrderStatus(
	ctx context.Context,
	chatID int64,
	UserId, BussinesId, status, OrderID string,
	Order *entity.UpdateOrderRequest,
	OrderStatus []*entity.OrderStatus, msg entity.Messaging,
) (bool, error) {

	orderID, err := strconv.Atoi(OrderID)
	if err != nil {
		return false, fmt.Errorf("converting OrderID to int: %w", err)
	}

	findStatusID, found := telegram.FindOrderStatusByName(OrderStatus, status)
	if !found {
		findStatusID = nil
	}
	Order.NewStatus = status
	Order.AdminStatus = findStatusID
	Order.OrderID = int64(orderID)

	updateMessage, err := t.ProductUscse.UpdateOrderStatus(ctx, Order)
	if err != nil {
		return false, fmt.Errorf("updating order status: %w", err)
	}
	fmt.Println(Order)
	err = t.ProductUscse.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
		PlatformID:  UserId,
		From:        "telegram",
		BusinessId:  BussinesId,
		OrderStatus: status,
	})
	if err != nil {
		return false, fmt.Errorf("updating client status: %w", err)
	}

	t.SendInstagramMessage(msg, updateMessage.Message)

	return updateMessage.Isupdate, nil
}

func (t *Handler) SendNotification(title, content string) error {

	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "notification",
		Notifications: &entity.Notification{
			UserId:    UserId,
			Title:     title,
			Content:   content,
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	})

	return nil
}



