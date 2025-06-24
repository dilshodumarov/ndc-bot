package entity

import (
	"time"
)

type Client struct {
	PlatformID string `json:"platform_id"`
	FirstName  string `json:"first_name"`
	Phone      string `json:"phone"`
	UserName   string
	From       string
	ChatId     int64
	BusinessId string
}

type UpdateClientStatusRequest struct {
	PlatformID  string `json:"platform_id"`
	From        string
	BusinessId  string
	Goal        string
	OrderStatus string
	Location    string
	StopStatus  *bool
	IsPauzse    *bool
	StopTime    *time.Time
	LocationText string
}

type ClientResponse struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

type ClientState struct {
	State         string
	ClientDetails Client
}

type OrderState struct {
	State      string  `json:"state"`
	ProductIDs []int64 `json:"product_ids,omitempty"`
}

type BotIntegration struct {
	Token      string `json:"token"`
	BusinessID string `json:"business_id"`
	UserID     string `json:"user_id"`
}

type IntegrationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Message_id int
}

type IntegrationResponseInsta struct {
	MessageID   string `json:"message_id"`   // ✅ To‘g‘ri type
	RecipientID string `json:"recipient_id"` // OK
	Code        int    `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
}


type BotNotification struct {
	Guid      string `json:"Guid"`
	ProductId string
}
type BusinessDescription struct {
	Description string
}

type BotCommand struct {
	Command  string
	Response string
}

type ChatHistory struct {
	MessageId        int
	GUID             string `json:"guid"`
	Message          string `json:"message"`
	BusinessId       string
	ChatID           int64     `json:"chat_id"`
	PlatformID       string    `json:"platform_id"`
	AIResponse       string    `json:"ai_response"`
	ReplyToMessageID int       `json:"reply_to_message_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Phone            string
	Platform         string
}

type SendMessageResponse struct {
	Type          string        `json:"type"`
	Notifications *Notification `json:"notifications,omitempty"`
	ChatMessage   *SendMessage  `json:"chat_message,omitempty"`
}

type Notification struct {
	UserId    string `json:"user_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type SendMessage struct {
	Message          string `json:"message"`
	AIResponse       string `json:"ai_response"`
	UserId           string `json:"user_id"`
	BusinessId       string `json:"business_id"`
	From             string `json:"from"`
	Platform         string `json:"platform"`
	Timestamp        string `json:"timestamp"`
	MessageId        int    `json:"message_id"`
	ReplyToMessageID int    `json:"reply_to_message_id"`
	Chatid           int64  `json:"chatid"`
}

type GetChatHistoryRequest struct {
	BusinessID string
	ChatID     int64
	TokenLimit int
	Phone      string
}

type SendMessageModel struct {
	ChatID           int64
	Message          string
	ReplyToMessageID *int
}

type ImageMessageRequest struct {
	Phone     string   `json:"phone"`
	UserID    int      `json:"user_id"`
	Message   string   `json:"message"`
	ImageURLs []string `json:"image_urls"`
}
