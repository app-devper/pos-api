package request

type CustomerHistory struct {
	CustomerCode string `json:"customerCode" binding:"required"`
	Type         string `json:"type" binding:"required"`
	Description  string `json:"description"`
	Reference    string `json:"reference"`
	CreatedBy    string
	BranchId     string
}
