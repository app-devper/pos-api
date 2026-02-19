package request

import "time"

type DeliveryOrder struct {
	OrderId      string              `json:"orderId" binding:"required"`
	CustomerCode string              `json:"customerCode"`
	CustomerName string              `json:"customerName"`
	Address      string              `json:"address"`
	Items        []DeliveryOrderItem `json:"items" binding:"required"`
	Note         string              `json:"note"`
	DeliveryDate time.Time           `json:"deliveryDate"`
	Code         string
	CreatedBy    string
	BranchId     string
}

type DeliveryOrderItem struct {
	ProductId string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

type UpdateDeliveryOrder struct {
	Address      string              `json:"address"`
	Items        []DeliveryOrderItem `json:"items" binding:"required"`
	Note         string              `json:"note"`
	DeliveryDate time.Time           `json:"deliveryDate"`
	Status       string              `json:"status"`
	UpdatedBy    string
}
