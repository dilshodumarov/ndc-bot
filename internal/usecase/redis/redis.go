package redisrepo

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/internal/repo"
)

type Uscase struct {
	RedisRepo repo.RedisRepo
}

func NewRedisRepo(redisrepo repo.RedisRepo) *Uscase {
	return &Uscase{RedisRepo: redisrepo}
}

func (r *Uscase) Get(ctx context.Context, chatID int64) (*entity.ClientState, error) {
	res, err := r.RedisRepo.Get(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("redisUseCase - History - r.repo.Get: %w", err)
	}

	return res, nil

}

func (r *Uscase) GetOrder(ctx context.Context, key string) (*entity.CreateOrder, error){
	res, err := r.RedisRepo.GetOrder(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("redisUseCase - History - r.repo.GetOrder: %w", err)
	}

	return res, nil

}

func (r *Uscase) Set(ctx context.Context, chatID int64, state *entity.ClientState) error {
	err := r.RedisRepo.Set(ctx, chatID, state)
	if err != nil {
		return fmt.Errorf("redisUseCase - History - r.repo.Set: %w", err)
	}

	return nil
}

func (r *Uscase) SetOrder(ctx context.Context, key string, state *entity.CreateOrder) error {
	err := r.RedisRepo.SetOrder(ctx, key, state)
	if err != nil {
		return fmt.Errorf("redisUseCase - History - r.repo.SetOrder: %w", err)
	}

	return nil
}

func (r *Uscase) Delete(ctx context.Context, chatID int64) error {
	err := r.RedisRepo.Delete(ctx, chatID)
	if err != nil {
		return fmt.Errorf("redisUseCase - History - r.repo.Get: %w", err)
	}

	return nil

}
