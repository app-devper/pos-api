package request

type Customer struct {
	CustomerType string `json:"customerType"`
	Name         string `json:"name" binding:"required"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Code         string
	CreatedBy    string
}

type UpdateCustomer struct {
	CustomerType string `json:"customerType"  binding:"required"`
	Name         string `json:"name" binding:"required"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	UpdatedBy    string
}

type UpdateCustomerStatus struct {
	Status    string `json:"status"`
	UpdatedBy string
}
