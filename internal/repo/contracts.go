// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"ndc/ai_bot/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// ProductRepo -.
	ProductRepo interface {
		GetProduct(ctx context.Context, req entity.GetProductRequest) ([]entity.Product, error)
		GetProductsByAlternatives(ctx context.Context, names []string, businessID string) ([]entity.Product, error)
		GetProductInfoForNotification(ctx context.Context, productID string) (*entity.ProductNotificationInfo, error)
		GetProductById(ctx context.Context, req entity.GetProductByIDRequest) (*entity.Product, error)
		ListProductsForAI(ctx context.Context, businessID string) ([]entity.ProductAI, error)
		CheckProductCount(ctx context.Context, businessID string, products []entity.ProductOrder) (*entity.ProductCheckResponse, error)
	}
	OrderRepo interface {
		CreateOrder(ctx context.Context, order entity.Order) (*entity.CreateOrderResponse, error)
		GetClientOrders(ctx context.Context, platformId, bussnesid string) ([]*entity.OrderResponseByOrderId, error)
		GetUsersByLastOrder(ctx context.Context) ([]*entity.LastOrders, error)
		GetOrderByID(ctx context.Context, order entity.GetOrderByID) (*entity.OrderResponseByOrderId, error)
		UpdateOrderStatus(ctx context.Context, req *entity.UpdateOrderRequest) (*entity.UpdateOrderResponse, error)
		RestoreProductCounts(ctx context.Context, req entity.CanseledOrder) (*entity.UpdateOrderResponse, error)
	}

	AuthRepo interface {
		CreateClient(ctx context.Context, client entity.Client) (*entity.ClientResponse, error)
		GetBotIntegrations(ctx context.Context) ([]*entity.BotIntegration, error)
		GetResponseByCommand(ctx context.Context, ownerID, command string) (string, error)
		UpdateClientStatus(ctx context.Context, req *entity.UpdateClientStatusRequest) error
		CreateTokenUsage(ctx context.Context, usage *entity.ClientTokenUsage) error
	}

	ChatRepo interface {
		CreateChatHistory(ctx context.Context, chatHistory *entity.ChatHistory) error
		GetChatHistory(ctx context.Context, req *entity.GetChatHistoryRequest) ([]map[string]interface{}, error)
		GetAllChatId(ctx context.Context, businessId string) ([]int64, error)
	}
	BusinessRepo interface {
		GetBusinessDescription(ctx context.Context, businessID string) (*entity.BusinessDescription, error)
		GetOrderStatusesByBusinessID(ctx context.Context, businessID string) ([]*entity.OrderStatus, error)
		GetBusinessByPhone(ctx context.Context, req entity.GetBussinesId) (*entity.BusinessInfo, error)
		ListBusinesses(ctx context.Context) ([]entity.Business, error)
		UpdateUserBusiness(ctx context.Context, userID string, businessID string) error
		GetAllMenusByOwnerID(ctx context.Context, ownerID string) ([]*entity.BotCommand, error)
		GetIntegrationSettingsByOwnerID(ctx context.Context, ownerID, platformid string) (*entity.IntegrationSettings, error)
	}

	RedisRepo interface {
		Get(ctx context.Context, chatID int64) (*entity.ClientState, error)
		Set(ctx context.Context, chatID int64, state *entity.ClientState) error
		Delete(ctx context.Context, chatID int64) error
		SetOrder(ctx context.Context, key string, state *entity.CreateOrder) error
		GetOrder(ctx context.Context, key string) (*entity.CreateOrder, error)
	}
)
