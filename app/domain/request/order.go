package request

type OrderPayment struct {
	Amount float64 `json:"amount" binding:"required"`
	Type   string  `json:"type" binding:"required"`
}

type OrderItem struct {
	ProductId     string  `json:"productId" binding:"required"`
	Quantity      int     `json:"quantity" binding:"required"`
	UnitId        string  `json:"unitId" binding:"required"`
	Price         float64 `json:"price" binding:"required"`
	CostPrice     float64 `json:"costPrice"`
	Discount      float64 `json:"discount"`
	AllowOversell bool    `json:"allowOversell"`
}

type GetOrderRange struct {
	StartDate FlexibleTime `form:"startDate" binding:"required"`
	EndDate   FlexibleTime `form:"endDate" binding:"required"`
	BranchId  string
}

type UpdateCustomerCode struct {
	CustomerCode string `json:"customerCode"`
}

type CancelOrderAction struct {
	Reason string `json:"reason"`
}
