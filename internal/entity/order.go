package entity

import "time"

type Order struct {
	ID            string
	UserID        int64
	Status        string
	TgUserName    string
	BusinessId    string
	Location      string   
	ProductOrders  []CreateProductOrder
	TotalPrice   int
	StatusId     *string
	Platform     string
	StatusNumber int
}

type CreateProductOrder struct {
    Id string `json:"product_id"`
    Count     int `json:"count"`
	Price    int
	TotalPrice int
}

type OrderProduct struct {
	ProductName string    `json:"product_name"`
	Count       int       `json:"count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LastOrders struct {
	ChatId  int64
	BotGuid string
}

type CreateOrderResponse struct {
	Message string
	Id  string
	TgUserName string
	OrderIdSerial int64
}


type CreateOrder struct {
	ID            []ProductOrder
	ChatId        int64
	UserId        int64  
	Location      string
	StatusId      *string
}


type UpdateOrderRequest struct {
	OrderID       int64    `json:"order_id"`
	NewStatus     string  `json:"new_status,omitempty"`
	PaymentMethod string  `json:"payment_method,omitempty"`
	LocationUrl   string
	ImageUrl      string
	AdminStatus   *string
	StatusNumber  int
	Location      string
	UserNote      string
}

type UpdateOrderResponse struct {
	Message  string
	Isupdate bool
}


type CanseledOrder struct {
	OrderID   int64  `json:"order_id"`
	NewStatus string `json:"new_status,omitempty"` 
	Reason    string `json:"reason"`
	StatusId  *string
}


type AddProductRequest struct {
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
	Count     int    `json:"count"`
}


type OrderResponseByOrderId struct {
	OrderID           int64    `json:"order_id"`
	Status            string    `json:"status"`
	StatusChangedTime time.Time `json:"status_changed_time"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Products          []*OrderProduct `json:"products"`
}


type GetOrderByID struct {
	OrderID  string
	PlatformId string
	BussnesId string
}


type OrderStatus struct {
	GUID       string            `json:"guid"`
	CustomName string            `json:"custom_name"`
	CreatedAt  time.Time         `json:"created_at"`
	Type       OrderStatusType   `json:"type"`
}

type OrderStatusType struct {
	GUID       string    `json:"guid"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}
