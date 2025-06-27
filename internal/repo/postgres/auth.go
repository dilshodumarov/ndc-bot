package postgres

import (
	"context"
	"fmt"
	"strings"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"
)

// AuthRepo -.
type AuthRepo struct {
	*postgres.Postgres
}

// New -.
func NewAuthRepo(pg *postgres.Postgres) *AuthRepo {
	return &AuthRepo{pg}
}

func (r *AuthRepo) CreateClient(ctx context.Context, client entity.Client) (*entity.ClientResponse, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - CreateClient - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	queryClient := `
		INSERT INTO "client" (first_name, user_name,from_chanel,bussnes_id,  phone, platform_id, chat_id,created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5,$6,$7,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING guid
	`
	var clientID string
	err = tx.QueryRow(ctx, queryClient, client.FirstName, client.UserName, client.From, client.BusinessId, client.Phone, client.PlatformID, client.ChatId).Scan(&clientID)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - CreateClient - QueryRow (client): %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("AuthRepo - CreateClient - Commit: %w", err)
	}

	return &entity.ClientResponse{
		Id:    clientID,
		Phone: client.Phone,
	}, nil
}

func (r *AuthRepo) UpdateClientStatus(ctx context.Context, req *entity.UpdateClientStatusRequest) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("UpdateClientStatus - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	setParts := []string{}
	args := []interface{}{}
	argPos := 1

	if req.OrderStatus != "" {
		setParts = append(setParts, fmt.Sprintf(`order_status = $%d`, argPos))
		args = append(args, req.OrderStatus)
		argPos++
	}

	if req.StopStatus != nil {

		setParts = append(setParts, fmt.Sprintf(`is_stop = $%d`, argPos))
		args = append(args, *req.StopStatus)
		argPos++
	}
	if req.LocationText != "" {
		setParts = append(setParts, fmt.Sprintf(`location_text = $%d`, argPos))
		args = append(args, req.LocationText)
		argPos++
	}
	if req.IsPauzse != nil {
		setParts = append(setParts, fmt.Sprintf(`is_pauze = $%d`, argPos))
		args = append(args, *req.IsPauzse)
		argPos++
	}
	if req.StopTime != nil {
		fmt.Println(888)
		setParts = append(setParts, fmt.Sprintf(`stop_until = $%d`, argPos))
		args = append(args, req.StopTime)
		argPos++
	}
	if req.Goal != "" {
		setParts = append(setParts, fmt.Sprintf(`goal = $%d`, argPos))
		args = append(args, req.Goal)
		argPos++
	}
	if req.Location != "" {
		setParts = append(setParts, fmt.Sprintf(`location = $%d`, argPos))
		args = append(args, req.Location)
		argPos++
	}
	setParts = append(setParts, `updated_at = CURRENT_TIMESTAMP`)

	if len(setParts) == 1 { // Faqat updated_at bo‘lsa
		return fmt.Errorf("Yangilanish uchun hech qanday ma'lumot kiritilmadi.")
	}

	query := fmt.Sprintf(`UPDATE "client" SET %s WHERE from_chanel = $%d AND platform_id = $%d AND bussnes_id = $%d`, strings.Join(setParts, ", "), argPos, argPos+1, argPos+2)
	args = append(args, req.From, req.PlatformID, req.BusinessId)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateClientStatus - update: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("UpdateClientStatus - commit: %w", err)
	}

	return nil
}

func (r *AuthRepo) GetBotIntegrations(ctx context.Context) ([]*entity.BotIntegration, error) {
	query := `
		SELECT 
			i.integration_token, 
			i.owner_id AS business_id,
			b.owner_id AS user_id
		FROM integration i
		JOIN business b ON b.guid = i.owner_id
		WHERE i.deleted_at IS NULL 
		  AND i.status = 'active'
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - GetBotIntegrations - Query: %w", err)
	}
	defer rows.Close()

	var bots []*entity.BotIntegration

	for rows.Next() {
		var bot entity.BotIntegration
		if err := rows.Scan(&bot.Token, &bot.BusinessID, &bot.UserID); err != nil {
			return nil, fmt.Errorf("AuthRepo - GetBotIntegrations - Scan: %w", err)
		}
		bots = append(bots, &bot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AuthRepo - GetBotIntegrations - Rows: %w", err)
	}

	return bots, nil
}

func (r *AuthRepo) GetResponseByCommand(ctx context.Context, ownerID, command string) (string, error) {
	query := `
		SELECT bc.response
		FROM bot_commands bc
		INNER JOIN integration i ON i.guid = bc.integration_id
		WHERE i.owner_id = $1 AND bc.command = $2 AND i.status = 'active'
		LIMIT 1
	`

	var response string
	err := r.Pool.QueryRow(ctx, query, ownerID, command).Scan(&response)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", nil
		}
		return "", fmt.Errorf("AuthRepo - GetResponseByCommand - Scan: %w", err)
	}

	return response, nil
}

func (r *AuthRepo) CreateTokenUsage(ctx context.Context, usage *entity.ClientTokenUsage) error {
	query := `
		INSERT INTO client_token_usage (business_id, source_type, used_for, request_tokens, response_tokens)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.Pool.Exec(ctx, query,
		usage.BusinessID,
		usage.SourceType,
		usage.UsedFor,
		usage.RequestTokens,
		usage.ResponseTokens,
	)
	if err != nil {
		return fmt.Errorf("AuthRepo - CreateTokenUsage - Exec: %w", err)
	}

	return nil
}
