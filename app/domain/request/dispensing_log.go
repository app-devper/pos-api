package request

type DispensingItem struct {
	ProductId   string `json:"productId" binding:"required"`
	ProductName string `json:"productName"`
	GenericName string `json:"genericName"`
	Quantity    int    `json:"quantity" binding:"required"`
	Unit        string `json:"unit"`
	Dosage      string `json:"dosage"`
	LotNumber   string `json:"lotNumber"`
}

type DispensingLog struct {
	OrderId        string           `json:"orderId" binding:"required"`
	PatientId      string           `json:"patientId" binding:"required"`
	Items          []DispensingItem `json:"items" binding:"required"`
	PharmacistName string           `json:"pharmacistName" binding:"required"`
	LicenseNo      string           `json:"licenseNo" binding:"required"`
	Note           string           `json:"note"`
	CreatedBy      string
	BranchId       string
}
