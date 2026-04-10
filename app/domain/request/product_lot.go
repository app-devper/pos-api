package request

type ProductLot struct {
	ProductId  string    `json:"productId" binding:"required"`
	Quantity   int       `json:"quantity" binding:"required"`
	LotNumber  string    `json:"lotNumber" binding:"required"`
	ExpireDate FlexibleTime `json:"expireDate" binding:"required"`
	CostPrice  float64   `json:"costPrice" binding:"required"`
	BranchId   string
	UpdatedBy  string
}

type UpdateProductLot struct {
	Quantity   int       `json:"quantity"`
	LotNumber  string    `json:"lotNumber" binding:"required"`
	ExpireDate FlexibleTime `json:"expireDate" binding:"required"`
	CostPrice  float64   `json:"costPrice" binding:"required"`
	BranchId   string
	UpdatedBy  string
}

type GetProductLotsExpireRange struct {
	StartDate FlexibleTime `form:"startDate" binding:"required"`
	EndDate   FlexibleTime `form:"endDate" binding:"required"`
}

type UpdateProductLotQuantity struct {
	Quantity  int `json:"quantity"`
	UpdatedBy string
}

type UpdateProductLotNotify struct {
	Notify    bool `json:"notify" binding:"required"`
	UpdatedBy string
}
