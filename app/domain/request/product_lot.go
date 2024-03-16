package request

import (
	"time"
)

type ProductLot struct {
	ProductId  string    `json:"productId" binding:"required"`
	Quantity   int       `json:"quantity" binding:"required"`
	LotNumber  string    `json:"lotNumber" binding:"required"`
	ExpireDate time.Time `json:"expireDate" binding:"required"`
	CostPrice  float64   `json:"costPrice" binding:"required"`
	UpdatedBy  string
}

type UpdateProductLot struct {
	Quantity   int       `json:"quantity"`
	LotNumber  string    `json:"lotNumber" binding:"required"`
	ExpireDate time.Time `json:"expireDate" binding:"required"`
	CostPrice  float64   `json:"costPrice" binding:"required"`
	UpdatedBy  string
}

type GetProductLotsExpireRange struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

type UpdateProductLotQuantity struct {
	Quantity  int `json:"quantity"`
	UpdatedBy string
}

type UpdateProductLotNotify struct {
	Notify    bool `json:"notify" binding:"required"`
	UpdatedBy string
}
