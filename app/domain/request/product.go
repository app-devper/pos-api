package request

type Product struct {
	Name         string  `json:"name" binding:"required"`
	NameEn       string  `json:"nameEn"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" binding:"required"`
	CostPrice    float64 `json:"costPrice" binding:"required"`
	Unit         string  `json:"unit"`
	Quantity     int     `json:"quantity"`
	SerialNumber string  `json:"serialNumber" binding:"required"`
	Category     string  `json:"category"`
	LotNumber    string  `json:"lotNumber"`
	ExpireDate   string  `json:"expireDate"`
	CreatedBy    string
}

type UpdateProduct struct {
	Name        string  `json:"name" binding:"required"`
	NameEn      string  `json:"nameEn"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
	CostPrice   float64 `json:"costPrice" binding:"required"`
	Unit        string  `json:"unit"`
	Quantity    int     `json:"quantity"`
	Category    string  `json:"category"`
	UpdatedBy   string
}

type ProductLot struct {
	Quantity   int     `json:"quantity" binding:"required"`
	LotNumber  string  `json:"lotNumber" binding:"required"`
	ExpireDate string  `json:"expireDate" binding:"required"`
	CostPrice  float64 `json:"costPrice"  binding:"required"`
}

type ProductPrice struct {
	ProductId     string  `json:"productId" binding:"required"`
	CustomerId    string  `json:"customerId" binding:"required"`
	CustomerPrice float64 `json:"customerPrice"  binding:"required"`
	CreatedBy     string
}
