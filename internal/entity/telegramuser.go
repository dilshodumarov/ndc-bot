package entity

import "time"

type PhoneNumber struct {
	Phone string `json:"phone"`
}

type CodeInput struct {
	Phone    string `json:"phone"`
	Code     string `json:"code"`
	Password string `json:"password,omitempty"`
}

type MessageRequest struct {
	Phone  string `json:"phone"`
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

type MessageResponse struct {
	Phone            string `json:"phone"`
	FromID           string `json:"fromid"`              // sender ID
	Text             string `json:"text"`                // matn
	MessageID        int    `json:"message_id"`          // xabar ID
	ReplyToMessageID *int   `json:"reply_to_message_id"` // reply bo‘lsa ID, bo‘lmasa null
	ReplyText        string `json:"reply_text"`          // reply xabarning matni
	Code             int    `json:"code"`                // kod (0 yoki 1)
	Message          string `json:"message"`             // tushunarli xabar
}

type IntegrationSettings struct {
	PromptText        string
	AiIsStop          bool
	Name              string
	ErrorMessage      string
	TokenLimit        int
	IntelligenceLevel int
	IsBlocked         bool
	StopUntil         int
	StopTime          time.Time
	IsStop            bool
	IsPauze           bool
	PromtOrder        string
	PromtProdcut      string
	ChatToken         int
}

type ClientTokenUsage struct {
	BusinessID     string `json:"business_id"`
	SourceType     string `json:"source_type"` // e.g. "bot", "telegram"
	UsedFor        string `json:"used_for"`    // e.g. "generate_text"
	RequestTokens  int    `json:"request_tokens"`
	ResponseTokens int    `json:"response_tokens"`
}

type BusinessInfo struct {
	BusinessID string `json:"business_id"`
	OwnerID    string `json:"owner_id"`
}

type GetBussinesId struct {
	Phone  string `json:"phone"`
	UserId string `json:"user_id"`
}
