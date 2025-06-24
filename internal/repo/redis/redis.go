// internal/storage/redis/client_state.go
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClientStateRepo struct {
	Redis *redis.Client
}

func NewClientStateRepo(redis *redis.Client) *ClientStateRepo {
	return &ClientStateRepo{Redis: redis}
}

func (r *ClientStateRepo) Get(ctx context.Context, chatID int64) (*entity.ClientState, error) {
	val, err := r.Redis.Get(ctx, clientKey(chatID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state entity.ClientState
	err = json.Unmarshal([]byte(val), &state)
	return &state, err
}

func (r *ClientStateRepo) Set(ctx context.Context, chatID int64, state  *entity.ClientState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.Redis.Set(ctx, clientKey(chatID), data, 30*time.Minute).Err()
}

func (r *ClientStateRepo) SetOrder(ctx context.Context, key string, state *entity.CreateOrder) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.Redis.Set(ctx, key, data, 30*time.Minute).Err()
}

func (r *ClientStateRepo) Delete(ctx context.Context, chatID int64) error {
	return r.Redis.Del(ctx, clientKey(chatID)).Err()
}

func clientKey(chatID int64) string {
	return fmt.Sprintf("client_state:%d", chatID)
}


func (r *ClientStateRepo) GetOrder(ctx context.Context, key string) (*entity.CreateOrder, error) {
	val, err := r.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var state entity.CreateOrder
	err = json.Unmarshal([]byte(val), &state)
	return &state, err
}