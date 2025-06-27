package postgres

import (
	"context"
	"fmt"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"
)

type TelegramAccountRepo struct {
	*postgres.Postgres
}

func NewTelegramAccountRepo(pg *postgres.Postgres) *TelegramAccountRepo {
	return &TelegramAccountRepo{pg}
}

// Create
func (r *TelegramAccountRepo) Create(ctx context.Context, req entity.CreateTelegramAccountRequest) (string, error) {
	query := `
		INSERT INTO telegram_accaunt (number, business_id)
		VALUES ($1, $2)
		RETURNING guid
	`
	var id string
	err := r.Pool.QueryRow(ctx, query, req.Number, req.BusinessID).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("TelegramAccountRepo - Create: %w", err)
	}
	return id, nil
}

// Get By ID
func (r *TelegramAccountRepo) Get(ctx context.Context, id string) (*entity.TelegramAccount, error) {
	query := `
		SELECT guid, number, business_id, status, created_at, updated_at
		FROM telegram_accaunt
		WHERE guid = $1
	`
	var acc entity.TelegramAccount
	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.Number, &acc.BusinessID, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("TelegramAccountRepo - Get: %w", err)
	}
	return &acc, nil
}

// Update
func (r *TelegramAccountRepo) Update(ctx context.Context, req entity.UpdateTelegramAccountRequest) error {
	query := `
		UPDATE telegram_accaunt
		SET number = $1, status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE guid = $3
	`
	_, err := r.Pool.Exec(ctx, query, req.Number, req.Status, req.ID)
	if err != nil {
		return fmt.Errorf("TelegramAccountRepo - Update: %w", err)
	}
	return nil
}

// Delete
func (r *TelegramAccountRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM telegram_accaunt WHERE guid = $1`
	_, err := r.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("TelegramAccountRepo - Delete: %w", err)
	}
	return nil
}

// List (optional: business_id orqali filter bilan)
func (r *TelegramAccountRepo) List(ctx context.Context, businessID string) ([]entity.TelegramAccount, error) {
	query := `
		SELECT guid, number, business_id, status, created_at, updated_at
		FROM telegram_accaunt
		WHERE business_id = $1
	`
	rows, err := r.Pool.Query(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("TelegramAccountRepo - List: %w", err)
	}
	defer rows.Close()

	var accounts []entity.TelegramAccount
	for rows.Next() {
		var acc entity.TelegramAccount
		if err := rows.Scan(
			&acc.ID, &acc.Number, &acc.BusinessID, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("TelegramAccountRepo - List Scan: %w", err)
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}



