// Package entity defines main entities for business logic (services), data base mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

// Translation -.
type Translation struct {
	Source      string `json:"source"       example:"auto"`
	Destination string `json:"destination"  example:"en"`
	Original    string `json:"original"     example:"текст для перевода"`
	Translation string `json:"translation"  example:"text for translation"`
}

type ProductQuery struct {
	Products     []ListProdutAi `json:"products"`
	AiResponse   string         `json:"AiResponse,omitempty"`
	IsAiResponse bool           `json:"IsAiResponse,omitempty"`
}

type ListProdutAi struct {
	UserMessage string `json:"user_message"`
	Id          int    `json:"id"`
}

type ActionResponse struct {
	AiResponse           string         `json:"AiResponse,omitempty"`
	IsAiResponse         bool           `json:"IsAiResponse,omitempty"`
	Action               string         `json:"action,omitempty"`
	OrderID              string         `json:"order_id,omitempty"`
	UserMessage          string         `json:"user_message,omitempty"`
	MessageID            int            `json:"message_id,omitempty"`
	Products             []ProductOrder `json:"products,omitempty"`
	Method               string         `json:"method,omitempty"`
	PaymentScreenshotURL string         `json:"payment_screenshot_url,omitempty"`
	IsProductSearch      bool           `json:"is_product_search,omitempty"`
	LocationUrl          string         `json:"location_url"`
	Reason               string         `json:"reason"`
	UserMessageID        int            `json:"user_message_id,omitempty"`
	ID                   int            `json:"id"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	Cost                 int            `json:"cost"`
	Count                int            `json:"count"`
	Title                string         `json:"title"`
	Message              string         `json:"message"`
	Location             string         `json:"location"`
	UserNote             string         `json:"user_note"`
}
type ProductOrder struct {
	ProductID int `json:"product_id"`
	Count     int `json:"count"`
}
