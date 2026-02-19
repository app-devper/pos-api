package request

type CreditNote struct {
	OrderId     string           `json:"orderId" binding:"required"`
	Reason      string           `json:"reason" binding:"required"`
	Items       []CreditNoteItem `json:"items" binding:"required"`
	TotalRefund float64          `json:"totalRefund"`
	Note        string           `json:"note"`
	Code        string
	CreatedBy   string
	BranchId    string
}

type CreditNoteItem struct {
	ProductId string  `json:"productId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	StockId   string  `json:"stockId"`
}

type UpdateCreditNote struct {
	Reason      string           `json:"reason"`
	Items       []CreditNoteItem `json:"items"`
	TotalRefund float64          `json:"totalRefund"`
	Note        string           `json:"note"`
	Status      string           `json:"status"`
	UpdatedBy   string
}
