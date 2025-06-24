package entity

import "time"

// Category struct corresponds to the "category" table
type Category struct {
	ID        string    `json:"id" db:"id"` // UUID primary key
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Attribute struct corresponds to the "attribute" table
type Attribute struct {
	ID         string    `json:"id" db:"id"` // UUID primary key
	Name       string    `json:"name" db:"name"`
	CategoryID string    `json:"category_id" db:"category_id"` // Foreign key
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Product struct corresponds to the "product" table
type Product struct {
	ID           string    `json:"id" db:"id"` // UUID primary key
	Name         string    `json:"name" db:"name"`
	CategoryID   string    `json:"category_id" db:"category_id"` // Foreign key
	ShortInfo    string    `json:"short_info" db:"short_info"`
	Description  string    `json:"description" db:"description"`
	Cost         int       `json:"cost" db:"cost"`
	Count        int       `json:"count" db:"count"`
	DiscountCost int       `json:"discount_cost" db:"discount_cost"`
	Discount     int       `json:"discount" db:"discount"`
	ProductId    int64
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Image_urls    []string
}

type ProductNotificationInfo struct {
	Name         string
	ShortInfo    string
	Description  string
	Cost         int
	Discount     int
	DiscountCost int
}



type Business struct {
	ID   string
	Name string
}


type GetProductRequest struct {
	Name        string `json:"name"`
	BusinessID  string `json:"business_id,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
}


type GetProductByIDRequest struct {
	ProductId   int64
	BusinessID  string
	PhoneNumber string
}


type ProductAI struct {
	ID   int64 `json:"id"`
	Name string `json:"name"`
	Cost int    `json:"cost"`
	Description string `json:"description"`
	Count  int 
}


// ProductCheckResponse mahsulot sonini tekshirish javobi
type ProductCheckResponse struct {
	ProductID int `json:"product_id"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
}