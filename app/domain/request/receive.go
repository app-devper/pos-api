package request

import "time"

type GetReceiveRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

type Receive struct {
	SupplierId string `json:"supplierId" binding:"required"`
	Reference  string `json:"reference"`
	Code       string
	UpdatedBy  string
}

type UpdateReceive struct {
	SupplierId string  `json:"supplierId" binding:"required"`
	Reference  string  `json:"reference"`
	TotalCost  float64 `json:"totalCost"`
	UpdatedBy  string
}

type UpdateReceiveTotalCode struct {
	TotalCost float64 `json:"totalCost"`
}
