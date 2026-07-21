package request

type ProductReturnItem struct {
	OrderItemId string  `json:"orderItemId" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	Refund      float64 `json:"refund"`
}

type ProductReturn struct {
	OrderId   string              `json:"orderId" binding:"required"`
	Reason    string              `json:"reason"`
	Note      string              `json:"note"`
	Items     []ProductReturnItem `json:"items" binding:"required"`
	BranchId  string
	CreatedBy string
}
