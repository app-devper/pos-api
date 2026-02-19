package request

import "time"

type PatientDrugAllergy struct {
	DrugName string `json:"drugName" binding:"required"`
	Reaction string `json:"reaction"`
	Severity string `json:"severity"`
}

type Patient struct {
	CustomerCode       string               `json:"customerCode" binding:"required"`
	IdCard             string               `json:"idCard"`
	DateOfBirth        time.Time            `json:"dateOfBirth"`
	Gender             string               `json:"gender"`
	BloodType          string               `json:"bloodType"`
	Weight             float64              `json:"weight"`
	Allergies          []PatientDrugAllergy `json:"allergies"`
	ChronicDiseases    []string             `json:"chronicDiseases"`
	CurrentMedications []string             `json:"currentMedications"`
	Note               string               `json:"note"`
	CreatedBy          string
	BranchId           string
}

type UpdatePatient struct {
	IdCard             string               `json:"idCard"`
	DateOfBirth        time.Time            `json:"dateOfBirth"`
	Gender             string               `json:"gender"`
	BloodType          string               `json:"bloodType"`
	Weight             float64              `json:"weight"`
	Allergies          []PatientDrugAllergy `json:"allergies"`
	ChronicDiseases    []string             `json:"chronicDiseases"`
	CurrentMedications []string             `json:"currentMedications"`
	Note               string               `json:"note"`
	UpdatedBy          string
}

type AllergyCheck struct {
	ProductIds []string `json:"productIds" binding:"required"`
}

type AllergyCheckResult struct {
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	DrugName    string `json:"drugName"`
	Reaction    string `json:"reaction"`
	Severity    string `json:"severity"`
}
