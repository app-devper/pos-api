package request

type Customer struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Code      string
	CreatedBy string
}

type UpdateCustomer struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	UpdatedBy string
}

type UpdateCustomerStatus struct {
	Status    string `json:"status"`
	UpdatedBy string
}
