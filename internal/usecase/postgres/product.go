package postgres

import (
	"context"
	"fmt"

	"ndc/ai_bot/internal/entity"
)

func (uc *UseCase) GetProduct(ctx context.Context, req entity.GetProductRequest) ([]entity.Product, error) {
	products, err := uc.Repo.GetProduct(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.Repo.GetHistory: %w", err)
	}

	return products, nil
}

func (uc *UseCase) GetProductsByAlternatives(ctx context.Context, name string, businessID string) ([]entity.Product, error) {
	// bitta stringni slice holatida uzatamiz
	names := []string{name}

	products, err := uc.Repo.GetProductsByAlternatives(ctx, names, businessID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductsByAlternatives: %w", err)
	}

	return products, nil
}
func (uc *UseCase) GetProductInfoForNotification(ctx context.Context, productID string) (*entity.ProductNotificationInfo, error) {
	products, err := uc.Repo.GetProductInfoForNotification(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductInfoForNotification: %w", err)
	}

	return products, nil
}
func (uc *UseCase) GetProductById(ctx context.Context, req entity.GetProductByIDRequest) (*entity.Product, error) {
	products, err := uc.Repo.GetProductById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductInfoForNotification: %w", err)
	}

	return products, nil
}

func (uc *UseCase) ListProductsForAI(ctx context.Context, businessID string) ([]entity.ProductAI, error) {
	Products, err := uc.Repo.ListProductsForAI(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetListBussness: %w", err)
	}

	return Products, nil
}

func (uc *UseCase) CheckProductCount(ctx context.Context, businessID string, products []entity.ProductOrder) (*entity.ProductCheckResponse, error) {
	Products, err := uc.Repo.CheckProductCount(ctx, businessID, products)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetLCheckProductCountistBussness: %w", err)
	}

	return Products, nil
}

// Translate -.
// func (uc *UseCase) Translate(ctx context.Context, t entity.Translation) (entity.Translation, error) {
// 	translation, err := uc.webAPI.Translate(t)
// 	if err != nil {
// 		return entity.Translation{}, fmt.Errorf("TranslationUseCase - Translate - s.webAPI.Translate: %w", err)
// 	}

// 	err = uc.Repo.Store(ctx, translation)
// 	if err != nil {
// 		return entity.Translation{}, fmt.Errorf("TranslationUseCase - Translate - s.Repo.Store: %w", err)
// 	}

// 	return translation, nil
// }
