package request

type Setting struct {
	BranchId       string `json:"branchId"`
	ReceiptFooter  string `json:"receiptFooter"`
	CompanyName    string `json:"companyName"`
	CompanyAddress string `json:"companyAddress"`
	CompanyPhone   string `json:"companyPhone"`
	CompanyTaxId   string `json:"companyTaxId"`
	LogoUrl        string `json:"logoUrl"`
	ShowCredit     bool   `json:"showCredit"`
	PromptPayId    string `json:"promptPayId"`
	UpdatedBy      string
}
