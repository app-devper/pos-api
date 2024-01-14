package request

type Supplier struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address" binding:"required"`
	Phone     string `json:"phone"`
	TaxId     string `json:"taxId"`
	CreatedBy string
}
