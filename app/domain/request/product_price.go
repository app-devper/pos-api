package request

type ProductPrice struct {
	ProductId    string  `json:"productId" binding:"required"`
	UnitId       string  `json:"unitId" binding:"required"`
	Price        float64 `json:"price" binding:"required"`
	CustomerType string  `json:"customerType" binding:"required"`
	UpdatedBy    string
}
