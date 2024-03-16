package request

import (
	"time"
)

type GetProduct struct {
	Category string `json:"category"`
}

type Product struct {
	Name         string    `json:"name" binding:"required"`
	NameEn       string    `json:"nameEn"`
	Description  string    `json:"description"`
	Price        float64   `json:"price" binding:"required"`
	CostPrice    float64   `json:"costPrice" binding:"required"`
	Unit         string    `json:"unit"`
	Quantity     int       `json:"quantity"`
	SerialNumber string    `json:"serialNumber" binding:"required"`
	Category     string    `json:"category"`
	LotNumber    string    `json:"lotNumber" binding:"required"`
	ExpireDate   time.Time `json:"expireDate" binding:"required"`
	ReceiveId    string    `json:"receiveId"`
	CreatedBy    string
	ReceiveCode  string
}

type UpdateProduct struct {
	Name         string  `json:"name" binding:"required"`
	NameEn       string  `json:"nameEn"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" binding:"required"`
	CostPrice    float64 `json:"costPrice" binding:"required"`
	Unit         string  `json:"unit"`
	Quantity     int     `json:"quantity"`
	SerialNumber string  `json:"serialNumber" binding:"required"`
	Category     string  `json:"category"`
	UpdatedBy    string
}

type ProductPrice struct {
	ProductId    string  `json:"productId" binding:"required"`
	UnitId       string  `json:"unitId" binding:"required"`
	Price        float64 `json:"price" binding:"required"`
	CustomerType string  `json:"customerType" binding:"required"`
	UpdatedBy    string
}

type CreateProductUnit struct {
	Price float64 `json:"price" binding:"required"`
	ProductUnit
}

type ProductUnit struct {
	ProductId  string  `json:"productId" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	Size       int     `json:"size" binding:"required"`
	CostPrice  float64 `json:"costPrice" binding:"required"`
	Volume     float64 `json:"volume"`
	VolumeUnit string  `json:"volumeUnit"`
	Barcode    string  `json:"barcode"`
	UpdatedBy  string
}

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

type GetExpireRange struct {
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
