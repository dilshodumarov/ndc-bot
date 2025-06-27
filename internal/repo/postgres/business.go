package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"
	"strings"

	"github.com/jackc/pgx"
)

type BusinessRepo struct {
	*postgres.Postgres
}

// New -.
func NewBusinessRepo(pg *postgres.Postgres) *BusinessRepo {
	return &BusinessRepo{pg}
}

func (r *BusinessRepo) GetBusinessDescription(ctx context.Context, businessID string) (*entity.BusinessDescription, error) {
	query := `
		SELECT description
		FROM business 
		WHERE guid = $1
	`

	var desc string
	err := r.Pool.QueryRow(ctx, query, businessID).Scan(&desc)
	if err != nil {
		return nil, fmt.Errorf("BusinessRepo - GetBusinessDescription - QueryRow: %w", err)
	}

	return &entity.BusinessDescription{
		Description: desc,
	}, nil
}

func (r *BusinessRepo) GetAllMenusByOwnerID(ctx context.Context, ownerID string) ([]*entity.BotCommand, error) {
	query := `
		SELECT bc.command, bc.response
		FROM bot_commands bc
		INNER JOIN integration i ON i.guid = bc.integration_id
		WHERE i.owner_id = $1 AND i.status = 'active'
	`

	rows, err := r.Pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("BusinessRepo - GetAllMenusByOwnerID - Query: %w", err)
	}
	defer rows.Close()

	var commands []*entity.BotCommand
	for rows.Next() {
		var cmd entity.BotCommand
		if err := rows.Scan(&cmd.Command, &cmd.Response); err != nil {
			return nil, fmt.Errorf("BusinessRepo - GetAllMenusByOwnerID - Scan: %w", err)
		}
		commands = append(commands, &cmd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BusinessRepo - GetAllMenusByOwnerID - Rows: %w", err)
	}

	return commands, nil
}

func (r *BusinessRepo) GetIntegrationSettingsByOwnerID(ctx context.Context, ownerID, platformID string) (*entity.IntegrationSettings, error) {
	query := `
		SELECT 
			s.prompt_text, 
			s.name,
			s.prompt_order,
			s.prompt_product,
			s.token_limit, 
			s.intelligence_level,
			s.stop_until,
			s.is_stop,
			s.error_message,
			s.chat_token,
			c.stop_until,
			c.is_stop,
			c.is_pauze,
			c.is_block
		FROM settings s
		LEFT JOIN client c ON c.bussnes_id = s.business_id AND c.platform_id = $2
		WHERE s.business_id = $1 AND s.status = true
		LIMIT 1
	`

	var (
		promptText        sql.NullString
		name              sql.NullString
		promptOrder       sql.NullString
		promptProduct     sql.NullString
		tokenLimit        sql.NullInt32
		intelligenceLevel sql.NullInt32
		stopUntil         sql.NullInt32
		stopTime          sql.NullTime
		isStop            sql.NullBool
		aiIsStop          sql.NullBool
		isPauze           sql.NullBool
		isBlocked         sql.NullBool
		errorMessage      sql.NullString
		ChatToken         sql.NullInt64
	)

	err := r.Pool.QueryRow(ctx, query, ownerID, platformID).Scan(
		&promptText,
		&name,
		&promptOrder,
		&promptProduct,
		&tokenLimit,
		&intelligenceLevel,
		&stopUntil,
		&aiIsStop,
		&errorMessage,
		&ChatToken,
		&stopTime,
		&isStop,
		&isPauze,
		&isBlocked,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("AuthRepo - GetIntegrationSettingsByOwnerID - Scan: %w", err)
	}

	settings := &entity.IntegrationSettings{}

	if promptText.Valid {
		settings.PromptText = promptText.String
	}
	if errorMessage.Valid {
		settings.ErrorMessage = errorMessage.String
	}
	if promptProduct.Valid {
		settings.PromtProdcut = promptProduct.String
	}
	if tokenLimit.Valid {
		settings.TokenLimit = int(tokenLimit.Int32)
	}
	if intelligenceLevel.Valid {
		settings.IntelligenceLevel = int(intelligenceLevel.Int32)
	}
	if stopUntil.Valid {
		settings.StopUntil = int(stopUntil.Int32)
	}
	if stopTime.Valid {
		settings.StopTime = stopTime.Time
	}
	if isStop.Valid {
		settings.IsStop = isStop.Bool
	}
	if aiIsStop.Valid {
		settings.AiIsStop = aiIsStop.Bool
	}
	if isPauze.Valid {
		settings.IsPauze = isPauze.Bool
	}
	if isBlocked.Valid {
		settings.IsBlocked = isBlocked.Bool
	}
	if ChatToken.Valid {
		settings.ChatToken = int(ChatToken.Int64)
	}

	// JSONB prompt_order ni stringga aylantirish
	if promptOrder.Valid {
		var orderMap map[string]string
		if err := json.Unmarshal([]byte(promptOrder.String), &orderMap); err != nil {
			return nil, fmt.Errorf("unmarshal prompt_order jsonb: %w", err)
		}

		keys := []string{"2", "3", "4", "5", "7"}

		var parts []string
		for _, k := range keys {
			if val, ok := orderMap[k]; ok && val != "" {
				parts = append(parts, val)
			}
		}
		settings.PromtOrder = strings.Join(parts, " ")
	}

	return settings, nil
}

func (r *BusinessRepo) GetOrderStatusesByBusinessID(ctx context.Context, businessID string) ([]*entity.OrderStatus, error) {
	query := `
		SELECT 
			os.guid,
			os.custom_name,
			os.created_at,
			ost.guid,
			ost.name,
			ost.created_at
		FROM 
			order_status os
		JOIN 
			order_status_type ost ON os.type_id = ost.guid
		WHERE 
			os.business_id = $1
		ORDER BY os.created_at DESC
	`

	rows, err := r.Pool.Query(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - GetOrderStatusesByBusinessID - Query: %w", err)
	}
	defer rows.Close()

	var statuses []*entity.OrderStatus

	for rows.Next() {
		var s entity.OrderStatus
		var t entity.OrderStatusType

		err := rows.Scan(
			&s.GUID,
			&s.CustomName,
			&s.CreatedAt,
			&t.GUID,
			&t.Name,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("OrderRepo - GetOrderStatusesByBusinessID - Scan: %w", err)
		}

		s.Type = t
		statuses = append(statuses, &s)
	}

	return statuses, nil
}

func (r *BusinessRepo) UpdateUserBusiness(ctx context.Context, userID string, businessID string) error {
	// Prepare SQL query to update the business_id for the given user
	sql, args, err := r.Builder.
		Update(`"user"`).
		Set("business_id", businessID).
		Where("guid = ?", userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("BusinessRepo - UpdateUserBusiness - ToSql: %w", err)
	}

	// Execute the query
	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BusinessRepo - UpdateUserBusiness - Exec: %w", err)
	}

	return nil
}

func (r *BusinessRepo) ListBusinesses(ctx context.Context) ([]entity.Business, error) {
	sql, args, err := r.Builder.
		Select("guid, name").
		From("business").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("BusinessRepo - ListBusinesses - ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("BusinessRepo - ListBusinesses - Query: %w", err)
	}
	defer rows.Close()

	var businesses []entity.Business
	for rows.Next() {
		var b entity.Business
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return nil, fmt.Errorf("BusinessRepo - ListBusinesses - Scan: %w", err)
		}
		businesses = append(businesses, b)
	}

	return businesses, nil
}

func (r *BusinessRepo) GetBusinessByPhone(ctx context.Context, req entity.GetBussinesId) (*entity.BusinessInfo, error) {
	var (
		query string
		arg   string
	)

	switch {
	case req.Phone != "":
		query = `
			SELECT b.guid AS business_id, b.owner_id
			FROM telegram_accaunt ta
			JOIN business b ON ta.business_id = b.guid
			WHERE ta.number = $1 AND ta.from = 'telegram' AND b.deleted_at IS NULL
		`
		arg = req.Phone

	case req.UserId != "":
		query = `
			SELECT b.guid AS business_id, b.owner_id
			FROM telegram_accaunt ta
			JOIN business b ON ta.business_id = b.guid
			WHERE ta.user_id = $1 AND ta.from = 'instagram' AND b.deleted_at IS NULL
		`
		arg = req.UserId

	default:
		return nil, fmt.Errorf("phone yoki user_id bo'lishi kerak")
	}

	var info entity.BusinessInfo
	err := r.Pool.QueryRow(ctx, query, arg).Scan(&info.BusinessID, &info.OwnerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("biznes ID topilmadi")
		}
		return nil, fmt.Errorf("OrderRepo - GetBusinessByPhone - Scan: %w", err)
	}

	return &info, nil
}
