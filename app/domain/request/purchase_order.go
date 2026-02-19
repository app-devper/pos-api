package request

import "time"

type PurchaseOrder struct {
	SupplierId string              `json:"supplierId" binding:"required"`
	Reference  string              `json:"reference"`
	Items      []PurchaseOrderItem `json:"items" binding:"required"`
	TotalCost  float64             `json:"totalCost"`
	Note       string              `json:"note"`
	DueDate    time.Time           `json:"dueDate"`
	Code       string
	CreatedBy  string
	BranchId   string
}

type PurchaseOrderItem struct {
	ProductId string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	CostPrice float64 `json:"costPrice" binding:"required"`
	Total     float64 `json:"total"`
}

type UpdatePurchaseOrder struct {
	SupplierId string              `json:"supplierId" binding:"required"`
	Reference  string              `json:"reference"`
	Items      []PurchaseOrderItem `json:"items" binding:"required"`
	TotalCost  float64             `json:"totalCost"`
	Note       string              `json:"note"`
	DueDate    time.Time           `json:"dueDate"`
	Status     string              `json:"status"`
	UpdatedBy  string
}
