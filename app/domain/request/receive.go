package request

import "time"

type GetReceiveRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
	BranchId  string
}

type Receive struct {
	SupplierId string        `json:"supplierId" binding:"required"`
	Reference  string        `json:"reference"`
	Items      []ReceiveItem `json:"items"`
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
	ProductId    string  `json:"productId" binding:"required"`
	CostPrice    float64 `json:"costPrice" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required"`
	LotNumber    string  `json:"lotNumber"`
	ExpireDate   string  `json:"expireDate"`
	UnitId       string  `json:"unitId"`
	BaseQuantity int     `json:"baseQuantity"`
}

type UpdateReceiveItems struct {
	ReceiveItems []ReceiveItem `json:"items"`
	UpdatedBy    string
}
