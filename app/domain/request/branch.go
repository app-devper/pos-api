package request

type Branch struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Code      string
	CreatedBy string
}

type UpdateBranch struct {
	Name      string `json:"name" binding:"required"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	UpdatedBy string
}

type UpdateBranchStatus struct {
	Status    string `json:"status" binding:"required"`
	UpdatedBy string
}
