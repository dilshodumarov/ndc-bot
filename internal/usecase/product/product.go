package product

import (
	"context"
	"fmt"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/repo"
)

// UseCase -.
type UseCase struct {
	repo      repo.ProductRepo
	repoOrder repo.OrderRepo
	AuthRepo  repo.AuthRepo
}

// New -.
func New(r repo.ProductRepo, or repo.OrderRepo, au repo.AuthRepo) *UseCase {
	return &UseCase{
		repo:      r,
		repoOrder: or,
		AuthRepo:  au,
	}
}

// History - getting translate history from store.
func (uc *UseCase) History(ctx context.Context) ([]entity.Product, error) {
	translations, err := uc.repo.GetHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return translations, nil
}

func (uc *UseCase)  GetProduct(ctx context.Context, req entity.GetProductRequest) ([]entity.Product, error) {
	products, err := uc.repo.GetProduct(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return products, nil
}

func (uc *UseCase) GetProductsByAlternatives(ctx context.Context, name string, businessID string) ([]entity.Product, error) {
	// bitta stringni slice holatida uzatamiz
	names := []string{name}

	products, err := uc.repo.GetProductsByAlternatives(ctx, names, businessID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductsByAlternatives: %w", err)
	}

	return products, nil
}
func (uc *UseCase) GetProductInfoForNotification(ctx context.Context, productID string) (*entity.ProductNotificationInfo, error) {
	products, err := uc.repo.GetProductInfoForNotification(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductInfoForNotification: %w", err)
	}

	return products, nil
}
func (uc *UseCase)GetProductById(ctx context.Context, req entity.GetProductByIDRequest) (*entity.Product, error) {
	products, err := uc.repo.GetProductById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetProductInfoForNotification: %w", err)
	}

	return products, nil
}

func (uc *UseCase) ListBusinesses(ctx context.Context) ([]entity.Business, error) {
	Business, err := uc.repo.ListBusinesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetListBussness: %w", err)
	}

	return Business, nil
}

func (uc *UseCase) UpdateUserBusiness(ctx context.Context, userID string, businessID string) error {
	err := uc.repo.UpdateUserBusiness(ctx, userID, businessID)
	if err != nil {
		return fmt.Errorf("UseCase - UpdateUserBusiness: %w", err)
	}

	return nil
}

func (uc *UseCase) ListProductsForAI(ctx context.Context, businessID string) ([]entity.ProductAI, error) {
	Products, err := uc.repo.ListProductsForAI(ctx,businessID)
	if err != nil {
		return nil, fmt.Errorf("UseCase - GetListBussness: %w", err)
	}

	return Products, nil
}

func (uc *UseCase) CheckProductCount(ctx context.Context, businessID string, products []entity.ProductOrder) (*entity.ProductCheckResponse, error){
	Products, err := uc.repo.CheckProductCount(ctx,businessID,products)
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

// 	err = uc.repo.Store(ctx, translation)
// 	if err != nil {
// 		return entity.Translation{}, fmt.Errorf("TranslationUseCase - Translate - s.repo.Store: %w", err)
// 	}

// 	return translation, nil
// }
