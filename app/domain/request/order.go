package request

import "time"

type Order struct {
	Items        []OrderItem `json:"items" binding:"required"`
	Amount       float64     `json:"amount" binding:"required"`
	Type         string      `json:"type" binding:"required"`
	CustomerCode string      `json:"customerCode"`
	CustomerName string      `json:"customerName"`
	Total        float64     `json:"total" binding:"required"`
	TotalCost    float64     `json:"totalCost"`
	Change       float64     `json:"change"`
	Message      string      `json:"message"`
	CreatedBy    string
	Code         string
}

type OrderItem struct {
	ProductId string           `json:"productId" binding:"required"`
	Quantity  int              `json:"quantity" binding:"required"`
	UnitId    string           `json:"unitId" binding:"required"`
	Price     float64          `json:"price" binding:"required"`
	CostPrice float64          `json:"costPrice"`
	Discount  float64          `json:"discount"`
	Stocks    []OrderItemStock `json:"stocks" binding:"required"`
}

type OrderItemStock struct {
	Quantity int    `json:"quantity" binding:"required"`
	StockId  string `json:"stockId"`
}

type GetOrderRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

type UpdateCustomerCode struct {
	CustomerCode string `json:"customerCode"`
}
