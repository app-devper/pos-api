package request

import (
	"time"
)

type GetProduct struct {
	Category string `json:"category"`
}

type Product struct {
	Name         string           `json:"name" binding:"required"`
	NameEn       string           `json:"nameEn"`
	Description  string           `json:"description"`
	Price        float64          `json:"price" binding:"required"`
	CostPrice    float64          `json:"costPrice" binding:"required"`
	Unit         string           `json:"unit"`
	Quantity     int              `json:"quantity"`
	SerialNumber string           `json:"serialNumber" binding:"required"`
	Category     string           `json:"category"`
	LotNumber    string           `json:"lotNumber" binding:"required"`
	ExpireDate   time.Time        `json:"expireDate" binding:"required"`
	ReceiveId    string           `json:"receiveId"`
	Status       string           `json:"status"`
	DrugInfo     *RequestDrugInfo `json:"drugInfo"`
	ReceiveCode  string
	CreatedBy    string
	BranchId     string
}

type CreateProduct struct {
	Name         string           `json:"name" binding:"required"`
	NameEn       string           `json:"nameEn"`
	Description  string           `json:"description"`
	Price        float64          `json:"price" binding:"required"`
	CostPrice    float64          `json:"costPrice" binding:"required"`
	Unit         string           `json:"unit" binding:"required"`
	SerialNumber string           `json:"serialNumber" binding:"required"`
	Category     string           `json:"category"`
	Status       string           `json:"status"`
	DrugInfo     *RequestDrugInfo `json:"drugInfo"`
	CreatedBy    string
}

type UpdateProduct struct {
	Name        string           `json:"name" binding:"required"`
	NameEn      string           `json:"nameEn"`
	Description string           `json:"description"`
	Category    string           `json:"category"`
	Status      string           `json:"status"`
	DrugInfo    *RequestDrugInfo `json:"drugInfo"`
	UpdatedBy   string
}

type RequestDrugInfo struct {
	GenericName       string `json:"genericName"`
	DrugType          string `json:"drugType"`
	DosageForm        string `json:"dosageForm"`
	Strength          string `json:"strength"`
	Indication        string `json:"indication"`
	Dosage            string `json:"dosage"`
	SideEffects       string `json:"sideEffects"`
	Contraindications string `json:"contraindications"`
	StorageCondition  string `json:"storageCondition"`
	Manufacturer      string `json:"manufacturer"`
	RegistrationNo    string `json:"registrationNo"`
	IsControlled      bool   `json:"isControlled"`
}
