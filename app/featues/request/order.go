package request

import "time"

type Order struct {
	Items     []OrderItem `json:"items" binding:"required"`
	Amount    float64     `json:"amount" binding:"required"`
	Type      string      `json:"type" binding:"required"`
	Total     float64     `json:"total"`
	TotalCost float64     `json:"totalCost"`
	Change    float64     `json:"change"`
	Message   string      `json:"message"`
	CreatedBy string
}

type OrderItem struct {
	ProductId string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	CostPrice float64 `json:"costPrice"`
	Discount  float64 `json:"discount"`
}

type GetOrderRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}
