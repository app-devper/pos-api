package request

import "time"

type Billing struct {
	CustomerCode string   `json:"customerCode" binding:"required"`
	CustomerName string   `json:"customerName"`
	OrderIds     []string `json:"orderIds" binding:"required"`
	TotalAmount  float64  `json:"totalAmount"`
	Note         string   `json:"note"`
	DueDate      time.Time `json:"dueDate"`
	Code         string
	CreatedBy    string
	BranchId     string
}

type UpdateBilling struct {
	OrderIds    []string  `json:"orderIds"`
	TotalAmount float64   `json:"totalAmount"`
	Note        string    `json:"note"`
	DueDate     time.Time `json:"dueDate"`
	Status      string    `json:"status"`
	UpdatedBy   string
}
