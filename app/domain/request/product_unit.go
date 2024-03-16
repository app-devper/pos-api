package request

type CreateProductUnit struct {
	ProductId  string  `json:"productId" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	Size       int     `json:"size" binding:"required"`
	CostPrice  float64 `json:"costPrice" binding:"required"`
	Price      float64 `json:"price" binding:"required"`
	Volume     float64 `json:"volume"`
	VolumeUnit string  `json:"volumeUnit"`
	Barcode    string  `json:"barcode"`
	UpdatedBy  string
}

type ProductUnit struct {
	ProductId  string  `json:"productId" binding:"required"`
	Unit       string  `json:"unit" binding:"required"`
	Size       int     `json:"size" binding:"required"`
	CostPrice  float64 `json:"costPrice" binding:"required"`
	Volume     float64 `json:"volume"`
	VolumeUnit string  `json:"volumeUnit"`
	Barcode    string  `json:"barcode"`
	UpdatedBy  string
}
