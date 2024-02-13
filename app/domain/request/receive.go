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
	TotalCode  float64 `json:"totalCode" binding:"required"`
	UpdatedBy  string
}

type UpdateReceiveTotalCode struct {
	TotalCode float64 `json:"totalCode" binding:"required"`
}
