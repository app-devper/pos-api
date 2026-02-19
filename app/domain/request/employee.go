package request

type Employee struct {
	BranchId  string `json:"branchId" binding:"required"`
	UserId    string `json:"userId" binding:"required"`
	Role      string `json:"role" binding:"required"`
	CreatedBy string
}

type UpdateEmployee struct {
	BranchId  string `json:"branchId" binding:"required"`
	Role      string `json:"role" binding:"required"`
	UpdatedBy string
}
