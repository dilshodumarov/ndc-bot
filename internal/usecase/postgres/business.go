package postgres

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
)

func (uc *UseCase) GetBusinessDescription(ctx context.Context, businessID string) (*entity.BusinessDescription, error) {
	Response, err := uc.Business.GetBusinessDescription(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetBusinessDescription %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetAllMenusByOwnerID(ctx context.Context, ownerID string) ([]*entity.BotCommand, error) {
	Response, err := uc.Business.GetAllMenusByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetAllMenusByOwnerID %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetIntegrationSettingsByOwnerID(ctx context.Context, ownerID, platformid string) (*entity.IntegrationSettings, error) {
	Response, err := uc.Business.GetIntegrationSettingsByOwnerID(ctx, ownerID, platformid)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetIntegrationSettingsByOwnerID %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetOrderStatusesByBusinessID(ctx context.Context, businessID string) ([]*entity.OrderStatus, error) {
	res, err := uc.Business.GetOrderStatusesByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetOrderStatusesByBusinessID %w", err)
	}

	return res, nil
}

func (uc *UseCase) GetBusinessByPhone(ctx context.Context, req entity.GetBussinesId) (*entity.BusinessInfo, error) {
	Response, err := uc.Business.GetBusinessByPhone(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetBusinessByPhone: %w", err)
	}

	return Response, nil
}

func (uc *UseCase) ListBusinesses(ctx context.Context) ([]entity.Business, error) {
	Business, err := uc.Business.ListBusinesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetListBussness: %w", err)
	}

	return Business, nil
}

func (uc *UseCase) UpdateUserBusiness(ctx context.Context, userID string, businessID string) error {
	err := uc.Business.UpdateUserBusiness(ctx, userID, businessID)
	if err != nil {
		return fmt.Errorf("UseCase - UpdateUserBusiness: %w", err)
	}

	return nil
}
