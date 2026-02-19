package request

import "time"

type Quotation struct {
	CustomerCode string          `json:"customerCode"`
	CustomerName string          `json:"customerName"`
	Items        []QuotationItem `json:"items" binding:"required"`
	TotalAmount  float64         `json:"totalAmount"`
	Discount     float64         `json:"discount"`
	Note         string          `json:"note"`
	ValidUntil   time.Time       `json:"validUntil"`
	Code         string
	CreatedBy    string
	BranchId     string
}

type QuotationItem struct {
	ProductId string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Total     float64 `json:"total"`
}

type UpdateQuotation struct {
	CustomerCode string          `json:"customerCode"`
	CustomerName string          `json:"customerName"`
	Items        []QuotationItem `json:"items"`
	TotalAmount  float64         `json:"totalAmount"`
	Discount     float64         `json:"discount"`
	Note         string          `json:"note"`
	ValidUntil   time.Time       `json:"validUntil"`
	Status       string          `json:"status"`
	UpdatedBy    string
}
