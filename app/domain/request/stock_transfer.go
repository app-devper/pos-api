package request

type StockTransferItem struct {
	ProductId string `json:"productId" binding:"required"`
	StockId   string `json:"stockId"`
	Quantity  int    `json:"quantity" binding:"required"`
}

type StockTransfer struct {
	ToBranchId string              `json:"toBranchId" binding:"required"`
	Items      []StockTransferItem `json:"items" binding:"required"`
	Note       string              `json:"note"`
	Code       string
	CreatedBy  string
	FromBranchId string
}

type UpdateStockTransfer struct {
	Status    string `json:"status" binding:"required"`
	UpdatedBy string
}
