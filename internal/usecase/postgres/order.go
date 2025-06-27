package postgres

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
)

func (uc *UseCase) CreateOrder(ctx context.Context, order entity.Order) (*entity.CreateOrderResponse, error) {
	Response, err := uc.RepoOrder.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetClientOrders(ctx context.Context, platformId, bussnesid string) ([]*entity.OrderResponseByOrderId, error) {
	Response, err := uc.RepoOrder.GetClientOrders(ctx, platformId, bussnesid)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetUsersByLastOrder(ctx context.Context) ([]*entity.LastOrders, error) {
	Response, err := uc.RepoOrder.GetUsersByLastOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetOrderByID(ctx context.Context, order entity.GetOrderByID) (*entity.OrderResponseByOrderId, error) {
	Response, err := uc.RepoOrder.GetOrderByID(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetOrderByID: %w", err)
	}

	return Response, nil
}

func (uc *UseCase) UpdateOrderStatus(ctx context.Context, req *entity.UpdateOrderRequest) (*entity.UpdateOrderResponse, error) {
	message, err := uc.RepoOrder.UpdateOrderStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.UpdateOrderStatus: %w", err)
	}

	return message, nil
}

func (uc *UseCase) RestoreProductCounts(ctx context.Context, req entity.CanseledOrder) (*entity.UpdateOrderResponse, error) {
	res, err := uc.RepoOrder.RestoreProductCounts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.RestoreProductCounts: %w", err)
	}

	return res, nil
}
