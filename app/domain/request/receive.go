package request

import "time"

type GetReceiveRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
	BranchId  string
}

type Receive struct {
	SupplierId string `json:"supplierId" binding:"required"`
	Reference  string `json:"reference"`
	Code       string
	UpdatedBy  string
	BranchId   string
}

type UpdateReceive struct {
	SupplierId   string        `json:"supplierId" binding:"required"`
	Reference    string        `json:"reference"`
	TotalCost    float64       `json:"totalCost"`
	ReceiveItems []ReceiveItem `json:"items"`
	UpdatedBy    string
}

type UpdateReceiveTotalCode struct {
	TotalCost float64 `json:"totalCost"`
}

type ReceiveItem struct {
	ProductId string  `json:"productId" binding:"required"`
	CostPrice float64 `json:"costPrice" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
}

type UpdateReceiveItems struct {
	ReceiveItems []ReceiveItem `json:"items"`
	UpdatedBy    string
}
