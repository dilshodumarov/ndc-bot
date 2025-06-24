package persistent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"

	"github.com/jackc/pgx/v5"
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

func (r *AuthRepo) GetBusinessDescription(ctx context.Context, businessID string) (*entity.BusinessDescription, error) {
	query := `
		SELECT description
		FROM business 
		WHERE guid = $1
	`

	var desc string
	err := r.Pool.QueryRow(ctx, query, businessID).Scan(&desc)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - GetBusinessDescription - QueryRow: %w", err)
	}

	return &entity.BusinessDescription{
		Description: desc,
	}, nil
}

func (r *AuthRepo) GetAllChatId(ctx context.Context, businessId string) ([]int64, error) {
	query := `
		SELECT chat_id
		FROM client
		WHERE bussnes_id = $1
	`

	rows, err := r.Pool.Query(ctx, query, businessId)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - GetAllChatId - Query: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("AuthRepo - GetAllChatId - Scan: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AuthRepo - GetAllChatId - Rows: %w", err)
	}

	return chatIDs, nil
}

func (r *AuthRepo) GetAllMenusByOwnerID(ctx context.Context, ownerID string) ([]*entity.BotCommand, error) {
	query := `
		SELECT bc.command, bc.response
		FROM bot_commands bc
		INNER JOIN integration i ON i.guid = bc.integration_id
		WHERE i.owner_id = $1 AND i.status = 'active'
	`

	rows, err := r.Pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("AuthRepo - GetAllMenusByOwnerID - Query: %w", err)
	}
	defer rows.Close()

	var commands []*entity.BotCommand
	for rows.Next() {
		var cmd entity.BotCommand
		if err := rows.Scan(&cmd.Command, &cmd.Response); err != nil {
			return nil, fmt.Errorf("AuthRepo - GetAllMenusByOwnerID - Scan: %w", err)
		}
		commands = append(commands, &cmd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AuthRepo - GetAllMenusByOwnerID - Rows: %w", err)
	}

	return commands, nil
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

func (r *AuthRepo) GetChatHistory(ctx context.Context, req *entity.GetChatHistoryRequest) ([]map[string]interface{}, error) {
	var (
		query string
		rows  pgx.Rows
		err   error
	)

	if req.Phone != "" {
		query = `
			SELECT ch.message_id, ch.message, ch.ai_response, ch.reply_to_message_id, ch.created_at, r.message AS reply_message
			FROM chat_history ch
			LEFT JOIN chat_history r ON ch.reply_to_message_id = r.message_id
			WHERE ch.phone = $1
			ORDER BY ch.created_at DESC
		`
		rows, err = r.Pool.Query(ctx, query, req.Phone)
	} else {
		query = `
			SELECT ch.message_id, ch.message, ch.ai_response, ch.reply_to_message_id, ch.created_at, r.message AS reply_message
			FROM chat_history ch
			LEFT JOIN chat_history r ON ch.reply_to_message_id = r.message_id
			WHERE ch.business_id = $1 AND ch.chat_id = $2
			ORDER BY ch.created_at DESC
		`
		rows, err = r.Pool.Query(ctx, query, req.BusinessID, req.ChatID)
	}

	if err != nil {
		return nil, fmt.Errorf("AuthRepo - GetChatHistory - Query: %w", err)
	}
	defer rows.Close()

	var (
		chatHistory  []map[string]interface{}
		tokenBudget  = req.TokenLimit // <-- yangi: token limiti
		tokenCounter int
	)

	for rows.Next() {
		var (
			message, aiResponse, replyMessage sql.NullString
			replyToMessageId, messageID       sql.NullInt64
			createdAt                         sql.NullTime
		)

		if err := rows.Scan(&messageID, &message, &aiResponse, &replyToMessageId, &createdAt, &replyMessage); err != nil {
			return nil, fmt.Errorf("AuthRepo - GetChatHistory - Scan: %w", err)
		}

		// AI javobi
		if aiResponse.Valid && aiResponse.String != "" {
			tokens := len([]rune(aiResponse.String))
			if tokenCounter+tokens > tokenBudget {
				break
			}
			tokenCounter += tokens
			chatHistory = append(chatHistory, map[string]interface{}{
				"role":       "assistant",
				"content":    aiResponse.String,
				"message_id": messageID.Int64,
				"created_at": createdAt.Time,
			})
		}

		// User xabari
		if message.Valid && message.String != "" {
			tokens := len([]rune(message.String))
			if tokenCounter+tokens > tokenBudget {
				break
			}
			tokenCounter += tokens
			chatHistory = append(chatHistory, map[string]interface{}{
				"role":       "user",
				"content":    message.String,
				"message_id": messageID.Int64,
				"created_at": createdAt.Time,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("AuthRepo - GetChatHistory - Rows: %w", err)
	}

	// Chatni eng so‘nggi tartibda olish (oldingi bo‘lsa oxirgi yozilganlar oxirida bo‘lardi)
	for i, j := 0, len(chatHistory)-1; i < j; i, j = i+1, j-1 {
		chatHistory[i], chatHistory[j] = chatHistory[j], chatHistory[i]
	}

	return chatHistory, nil
}

func (r *AuthRepo) CreateChatHistory(ctx context.Context, chatHistory *entity.ChatHistory) error {
	query := `
		INSERT INTO chat_history (message_id,business_id,phone, platform_id, chat_id, message, ai_response, reply_to_message_id)
		VALUES ($1, $2, $3, $4, $5, $6,$7,$8)
	`

	_, err := r.Pool.Exec(ctx, query, chatHistory.MessageId, chatHistory.BusinessId, chatHistory.Phone, chatHistory.PlatformID, chatHistory.ChatID, chatHistory.Message, chatHistory.AIResponse, chatHistory.ReplyToMessageID)
	if err != nil {
		return fmt.Errorf("AuthRepo - CreateChatHistory - Exec: %w", err)
	}

	return nil
}

func (r *AuthRepo) GetIntegrationSettingsByOwnerID(ctx context.Context, ownerID, platformID string) (*entity.IntegrationSettings, error) {
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

func (r *AuthRepo) GetOrderStatusesByBusinessID(ctx context.Context, businessID string) ([]*entity.OrderStatus, error) {
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

