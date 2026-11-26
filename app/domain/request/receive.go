package request

type GetReceiveRange struct {
	StartDate FlexibleTime `form:"startDate" binding:"required"`
	EndDate   FlexibleTime `form:"endDate" binding:"required"`
	BranchId  string
	// ExcludeCancelled leaves out cancelled Receives; the handler sets it from
	// the opt-in `excludeCancelled` query parameter.
	ExcludeCancelled bool `form:"-"`
}

type Receive struct {
	SupplierId string        `json:"supplierId" binding:"required"`
	Reference  string        `json:"reference"`
	Items      []ReceiveItem `json:"items"`
	Code       string
	UpdatedBy  string
	BranchId   string
}

type UpdateReceive struct {
	SupplierId   string        `json:"supplierId" binding:"required"`
	Reference    string        `json:"reference"`
	ReceiveItems []ReceiveItem `json:"items"`
	UpdatedBy    string
}

type UpdateReceiveTotalCost struct {
	TotalCost float64 `json:"totalCost"`
}

type ReceiveItem struct {
	ProductId    string  `json:"productId" binding:"required"`
	CostPrice    float64 `json:"costPrice" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required"`
	LotNumber    string  `json:"lotNumber"`
	ExpireDate   string  `json:"expireDate"`
	UnitId       string  `json:"unitId"`
	BaseQuantity int     `json:"baseQuantity"`
}

type UpdateReceiveItems struct {
	ReceiveItems []ReceiveItem `json:"items"`
	UpdatedBy    string
}
