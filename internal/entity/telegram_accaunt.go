package entity

type TelegramAccount struct {
	ID         string `json:"id"`
	Number     string `json:"number"`
	BusinessID string `json:"business_id"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type CreateTelegramAccountRequest struct {
	Number     string `json:"number"`
	BusinessID string `json:"business_id"`
}

type UpdateTelegramAccountRequest struct {
	ID         string `json:"id"`
	Number     string `json:"number"`
	Status     string `json:"status"`
}
