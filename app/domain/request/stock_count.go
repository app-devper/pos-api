package request

type StockCountItem struct {
	ProductId string `json:"productId" binding:"required"`
	StockId   string `json:"stockId" binding:"required"`
	Counted   int    `json:"counted"`
}

type StockCount struct {
	Note      string           `json:"note"`
	Items     []StockCountItem `json:"items" binding:"required"`
	BranchId  string
	CreatedBy string
}
