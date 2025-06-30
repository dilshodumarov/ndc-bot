package telegram

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"strconv"
	"strings"
	"time"
)

func (t *Handler) CreateOrder(ctx context.Context, res *entity.CreateOrder) (int64,error) {
	var products entity.CreateProductOrder
	order := entity.Order{
		UserID:       res.UserId,
		Status:       "olishga_tayyor",
		StatusId:     res.StatusId,
		StatusNumber: 2,
		Platform:     "bot",
	}

	for i := 0; i < len(res.ID); i++ {

		product, err := t.Usecase.GetProductById(ctx, entity.GetProductByIDRequest{
			ProductId:  int64(res.ID[i].ProductID),
			BusinessID: t.BusinessId,
		})
		if err != nil || product == nil {
			t.SendTelegramMessage(ctx, entity.SendMessageModel{
				ChatID:  res.ChatId,
				Message: fmt.Sprintf("❌ Mahsulot topilmadi (ID: %d).", res.ID[i]),
			})
			return 0,err
		}
		products.Id = product.ID
		products.Count = res.ID[i].Count
		products.TotalPrice = res.ID[i].Count * product.Cost
		products.Price = product.Cost
		order.TotalPrice += res.ID[i].Count * product.Cost
		order.BusinessId = t.BusinessId
		order.ProductOrders = append(order.ProductOrders, products)
	}

	orderResponse, err := t.Usecase.CreateOrder(ctx, order)
	if err != nil {
		fmt.Println("❌ Buyurtma yaratishda xatolik: " + err.Error())
		t.SendErrorTelegramMessage(ctx, res.ChatId)
		return 0,err
	}

	if orderResponse.Message == "norows" {
		state := &entity.ClientState{
			State: "first_name",
		}

		err := t.RedisUsecase.RedisRepo.Set(ctx, res.ChatId, state)
		if err != nil {
			fmt.Println("❌ Redisda state saqlashda xatolik: " + err.Error())
			t.SendErrorTelegramMessage(ctx, res.ChatId)
			return 0,err
		}
		t.SendTelegramMessage(ctx, entity.SendMessageModel{
			ChatID:  res.ChatId,
			Message: "Ismingizni kiriting:",
		})
		return 0,err
	}
	platformID := fmt.Sprintf("%d", order.UserID)
	err = t.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
		PlatformID:  platformID,
		From:        "bot",
		BusinessId:  order.BusinessId,
		Goal:        "sotipOlmoqchoi",
		OrderStatus: order.Status,
	})
	if err != nil {
		fmt.Println("Update statusda hatolik: " + err.Error())
		t.SendErrorTelegramMessage(ctx, res.ChatId)
		return 0,err
	}
	// response := fmt.Sprintf(
	// 	"✅ Buyurtmangiz muvaffaqiyatli rasmiylashtirildi!\n\n🆔 Buyurtma Id: #%d\n💵 Umumiy narx: %d so'm\n\nEndi locatsiangizni yuboring.",
	// 	orderResponse.OrderIdSerial,
	// 	order.TotalPrice,
	// )
	// t.SendTelegramMessage(ctx, entity.SendMessageModel{
	// 	ChatID:  res.ChatId,
	// 	Message: response,
	// })

	return orderResponse.OrderIdSerial,nil

}

func (t *Handler) handleOrderStatus(
	ctx context.Context,
	chatID int64,
	UserId, BussinesId, status, OrderID string,
	Order *entity.UpdateOrderRequest,
	OrderStatus []*entity.OrderStatus,
) (bool, error) {

	orderID, err := strconv.Atoi(OrderID)
	if err != nil {
		return false, fmt.Errorf("converting OrderID to int: %w", err)
	}

	findStatusID, found := FindOrderStatusByName(OrderStatus, status)
	if !found {
		findStatusID = nil
	}
	Order.NewStatus = status
	Order.AdminStatus = findStatusID
	Order.OrderID = int64(orderID)

	updateMessage, err := t.Usecase.UpdateOrderStatus(ctx, Order)
	if err != nil {
		return false, fmt.Errorf("updating order status: %w", err)
	}
	fmt.Println(Order)
	err = t.Usecase.UpdateClientStatus(ctx, &entity.UpdateClientStatusRequest{
		PlatformID:  UserId,
		From:        "bot",
		BusinessId:  BussinesId,
		OrderStatus: status,
	})
	if err != nil {
		return false, fmt.Errorf("updating client status: %w", err)
	}

	t.SendTelegramMessage(ctx, entity.SendMessageModel{
		ChatID:  chatID,
		Message: updateMessage.Message,
	})

	return updateMessage.Isupdate, nil
}

func (t *Handler) SendNotification(title, content string) error {

	t.SendMessageToAdmin(entity.SendMessageResponse{
		Type: "notification",
		Notifications: &entity.Notification{
			UserId:    t.UserId,
			Title:     title,
			Content:   content,
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	})

	return nil
}

func (t *Handler) ClientOrders(orders []*entity.OrderResponseByOrderId, chatID int64) {
	var builder strings.Builder

	if len(orders) == 0 {
		builder.WriteString("❌ Sizda hali hech qanday buyurtma mavjud emas.")
		return
	} else {
		builder.WriteString("📦 Buyurtmalaringiz:\n\n")
		for i, order := range orders {

			if order != nil {
				builder.WriteString(fmt.Sprintf("📌 #%d - Holat: %s\n🕒 Yaratilgan: %s\n", i+1, order.Status, order.CreatedAt.Format("2006-01-02 15:04")))
			}

			for _, p := range order.Products {
				builder.WriteString(fmt.Sprintf("  🛒 %s x%d\n", p.ProductName, p.Count))
			}
			builder.WriteString("\n")
		}
	}
	t.SendTelegramMessage(context.Background(), entity.SendMessageModel{
		ChatID:  chatID,
		Message: builder.String(),
	})
}

