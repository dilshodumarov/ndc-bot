package postgres

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
)

func (uc *UseCase) GetAllChatId(ctx context.Context, businessId string) ([]int64, error) {
	Response, err := uc.Chat.GetAllChatId(ctx, businessId)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetAllChatId %w", err)
	}

	return Response, nil
}

func (uc *UseCase) CreateChatHistory(ctx context.Context, chatHistory *entity.ChatHistory) error {
	err := uc.Chat.CreateChatHistory(ctx, chatHistory)
	if err != nil {
		return fmt.Errorf("TranslationUseCase - CreateChatHistory %w", err)
	}

	return nil
}

func (uc *UseCase) GetChatHistory(ctx context.Context, req *entity.GetChatHistoryRequest) ([]map[string]interface{}, error) {
	Response, err := uc.Chat.GetChatHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - GetResponseByCommand %w", err)
	}

	return Response, nil
}
