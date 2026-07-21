package request

type StockAdjustment struct {
	ProductId string `json:"productId" binding:"required"`
	StockId   string `json:"stockId" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
	Note      string `json:"note"`
	Delta     int    `json:"delta" binding:"required"`
	BranchId  string
	CreatedBy string
}
