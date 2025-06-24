package product

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
)

func (uc *UseCase) CreateClient(ctx context.Context, client entity.Client) (*entity.ClientResponse, error) {
	Response, err := uc.AuthRepo.CreateClient(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - createuser %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetBotIntegrations(ctx context.Context) ([]*entity.BotIntegration, error) {
	Response, err := uc.AuthRepo.GetBotIntegrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetBotIntegrations %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetBusinessDescription(ctx context.Context, businessID string) (*entity.BusinessDescription, error) {
	Response, err := uc.AuthRepo.GetBusinessDescription(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetBusinessDescription %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetAllChatId(ctx context.Context, businessId string) ([]int64, error) {
	Response, err := uc.AuthRepo.GetAllChatId(ctx, businessId)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetAllChatId %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetAllMenusByOwnerID(ctx context.Context, ownerID string) ([]*entity.BotCommand, error) {
	Response, err := uc.AuthRepo.GetAllMenusByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetAllMenusByOwnerID %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetResponseByCommand(ctx context.Context, ownerID, command string) (string, error) {
	Response, err := uc.AuthRepo.GetResponseByCommand(ctx, ownerID, command)
	if err != nil {
		return "", fmt.Errorf("TranslationUseCase - GetResponseByCommand %w", err)
	}

	return Response, nil
}

func (uc *UseCase) CreateChatHistory(ctx context.Context, chatHistory *entity.ChatHistory) error {
	err := uc.AuthRepo.CreateChatHistory(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("TranslationUseCase - CreateChatHistory %w", err)
	}

	return nil
}

func (uc *UseCase) GetChatHistory(ctx context.Context, req *entity.GetChatHistoryRequest) ([]map[string]interface{}, error) {
	Response, err := uc.AuthRepo.GetChatHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetResponseByCommand %w", err)
	}

	return Response, nil
}

func (uc *UseCase) GetIntegrationSettingsByOwnerID(ctx context.Context, ownerID, platformid string) (*entity.IntegrationSettings, error) {
	Response, err := uc.AuthRepo.GetIntegrationSettingsByOwnerID(ctx, ownerID, platformid)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetIntegrationSettingsByOwnerID %w", err)
	}

	return Response, nil
}

func (uc *UseCase) UpdateClientStatus(ctx context.Context, req *entity.UpdateClientStatusRequest) error {
	err := uc.AuthRepo.UpdateClientStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("TranslationUseCase - UpdateClientStatus %w", err)
	}

	return nil
}

func (uc *UseCase) CreateTokenUsage(ctx context.Context, usage *entity.ClientTokenUsage) error {
	err := uc.AuthRepo.CreateTokenUsage(ctx, usage)
	if err != nil {
		return fmt.Errorf("TranslationUseCase - CreateTokenUsage %w", err)
	}

	return nil
}

func (uc *UseCase) GetOrderStatusesByBusinessID(ctx context.Context, businessID string) ([]*entity.OrderStatus, error) {
	res, err := uc.AuthRepo.GetOrderStatusesByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetOrderStatusesByBusinessID %w", err)
	}

	return res, nil
}
