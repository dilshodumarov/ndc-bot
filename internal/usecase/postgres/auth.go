package postgres
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






func (uc *UseCase) GetResponseByCommand(ctx context.Context, ownerID, command string) (string, error) {
	Response, err := uc.AuthRepo.GetResponseByCommand(ctx, ownerID, command)
	if err != nil {
		return "", fmt.Errorf("TranslationUseCase - GetResponseByCommand %w", err)
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


