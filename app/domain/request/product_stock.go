package request

import "time"

type ProductStock struct {
	ProductId   string    `json:"productId" binding:"required"`
	UnitId      string    `json:"unitId" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required"`
	ReceiveCode string    `json:"receiveCode"`
	LotNumber   string    `json:"lotNumber"`
	CostPrice   float64   `json:"costPrice"`
	Price       float64   `json:"price"`
	ExpireDate  time.Time `json:"expireDate" binding:"required"`
	ImportDate  time.Time `json:"importDate" binding:"required"`
	UpdatedBy   string
}

type UpdateProductStock struct {
	ProductId  string    `json:"productId" binding:"required"`
	UnitId     string    `json:"unitId" binding:"required"`
	LotNumber  string    `json:"lotNumber"`
	CostPrice  float64   `json:"costPrice"`
	Price      float64   `json:"price"`
	ExpireDate time.Time `json:"expireDate" binding:"required"`
	ImportDate time.Time `json:"importDate" binding:"required"`
	UpdatedBy  string
}

type UpdateProductStockQuantity struct {
	Quantity  int `json:"quantity"`
	UpdatedBy string
}

type UpdateProductStockSequence struct {
	Stocks []ProductStockSequence `json:"stocks" binding:"required"`
}

type ProductStockSequence struct {
	StockId  string `json:"stockId" binding:"required"`
	Sequence int    `json:"sequence" binding:"required"`
}
