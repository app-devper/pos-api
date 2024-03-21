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
	ReceiveCode  string
	CreatedBy    string
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
