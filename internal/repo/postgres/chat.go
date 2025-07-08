package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"

	"github.com/jackc/pgx/v5"
)

type ChatRepo struct {
	*postgres.Postgres
}

// New -.
func NewChatepo(pg *postgres.Postgres) *ChatRepo {
	return &ChatRepo{pg}
}

func (r *ChatRepo) GetAllChatId(ctx context.Context, businessId string) ([]int64, error) {
	query := `
		SELECT chat_id
		FROM client
		WHERE bussnes_id = $1
	`

	rows, err := r.Pool.Query(ctx, query, businessId)
	if err != nil {
		return nil, fmt.Errorf("ChatRepo - GetAllChatId - Query: %w", err)
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("ChatRepo - GetAllChatId - Scan: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ChatRepo - GetAllChatId - Rows: %w", err)
	}

	return chatIDs, nil
}

func (r *ChatRepo) GetChatHistory(ctx context.Context, req *entity.GetChatHistoryRequest) ([]map[string]interface{}, error) {
	var (
		query      string
		rows       pgx.Rows
		err        error
		dateFilter string
		args       []interface{}
	)

	// Date filtering
	if req.Days > 0 {
		dateFilter = "AND ch.created_at >= NOW() - INTERVAL '%d days'"
		dateFilter = fmt.Sprintf(dateFilter, req.Days)
	}

	if req.Phone != "" {
		query = fmt.Sprintf(`
			SELECT ch.message_id, ch.message, ch.ai_response, ch.reply_to_message_id, ch.created_at, r.message AS reply_message
			FROM chat_history ch
			LEFT JOIN chat_history r ON ch.reply_to_message_id = r.message_id
			WHERE ch.phone = $1 %s
			ORDER BY ch.created_at DESC
		`, dateFilter)
		args = append(args, req.Phone)
	} else {
		query = fmt.Sprintf(`
			SELECT ch.message_id, ch.message, ch.ai_response, ch.reply_to_message_id, ch.created_at, r.message AS reply_message
			FROM chat_history ch
			LEFT JOIN chat_history r ON ch.reply_to_message_id = r.message_id
			WHERE ch.business_id = $1 AND ch.chat_id = $2 %s
			ORDER BY ch.created_at DESC
		`, dateFilter)
		args = append(args, req.BusinessID, req.ChatID)
	}

	rows, err = r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ChatRepo - GetChatHistory - Query: %w", err)
	}
	defer rows.Close()

	var (
		chatHistory  []map[string]interface{}
		tokenBudget  = req.TokenLimit
		tokenCounter int
	)

	for rows.Next() {
		var (
			message, aiResponse, replyMessage sql.NullString
			replyToMessageId, messageID       sql.NullInt64
			createdAt                         sql.NullTime
		)

		if err := rows.Scan(&messageID, &message, &aiResponse, &replyToMessageId, &createdAt, &replyMessage); err != nil {
			return nil, fmt.Errorf("ChatRepo - GetChatHistory - Scan: %w", err)
		}

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
		return nil, fmt.Errorf("ChatRepo - GetChatHistory - Rows: %w", err)
	}

	// Reverse chat order
	for i, j := 0, len(chatHistory)-1; i < j; i, j = i+1, j-1 {
		chatHistory[i], chatHistory[j] = chatHistory[j], chatHistory[i]
	}

	return chatHistory, nil
}


func (r *ChatRepo) CreateChatHistory(ctx context.Context, chatHistory *entity.ChatHistory) error {
	query := `
		INSERT INTO chat_history (message_id,business_id,phone, platform_id, chat_id, message, ai_response, reply_to_message_id)
		VALUES ($1, $2, $3, $4, $5, $6,$7,$8)
	`
	_, err := r.Pool.Exec(ctx, query, chatHistory.MessageId, chatHistory.BusinessId, chatHistory.Phone, chatHistory.PlatformID, chatHistory.ChatID, chatHistory.Message, chatHistory.AIResponse, chatHistory.ReplyToMessageID)
	if err != nil {
		return fmt.Errorf("ChatRepo - CreateChatHistory - Exec: %w", err)
	}

	return nil
}





//