package request

import "time"

type Promotion struct {
	Code        string    `json:"code" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Type        string    `json:"type" binding:"required"`
	Value       float64   `json:"value" binding:"required"`
	MinPurchase float64   `json:"minPurchase"`
	MaxDiscount float64   `json:"maxDiscount"`
	ProductIds  []string  `json:"productIds"`
	StartDate   time.Time `json:"startDate" binding:"required"`
	EndDate     time.Time `json:"endDate" binding:"required"`
	CreatedBy   string
	BranchId    string
}

type UpdatePromotion struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Value       float64   `json:"value"`
	MinPurchase float64   `json:"minPurchase"`
	MaxDiscount float64   `json:"maxDiscount"`
	ProductIds  []string  `json:"productIds"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Status      string    `json:"status"`
	UpdatedBy   string
}

type ApplyPromotion struct {
	PromotionCode string  `json:"promotionCode" binding:"required"`
	OrderTotal    float64 `json:"orderTotal" binding:"required"`
	ProductIds    []string `json:"productIds"`
}

type ApplyPromotionResult struct {
	PromotionId string  `json:"promotionId"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Discount    float64 `json:"discount"`
}
