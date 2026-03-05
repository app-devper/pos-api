package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DrugAllergy struct {
	DrugName string `bson:"drugName" json:"drugName"`
	Reaction string `bson:"reaction" json:"reaction"`
	Severity string `bson:"severity" json:"severity"`
}

type Patient struct {
	Id                 primitive.ObjectID `bson:"_id" json:"id"`
	BranchId           primitive.ObjectID `bson:"branchId" json:"branchId"`
	CustomerCode       string             `bson:"customerCode" json:"customerCode"`
	FirstName          string             `bson:"firstName" json:"firstName"`
	LastName           string             `bson:"lastName" json:"lastName"`
	IdCard             string             `bson:"idCard" json:"idCard"`
	Phone              string             `bson:"phone" json:"phone"`
	Email              string             `bson:"email" json:"email"`
	Address            string             `bson:"address" json:"address"`
	DateOfBirth        time.Time          `bson:"dateOfBirth" json:"dateOfBirth"`
	Gender             string             `bson:"gender" json:"gender"`
	BloodType          string             `bson:"bloodType" json:"bloodType"`
	Weight             float64            `bson:"weight" json:"weight"`
	Allergies          []DrugAllergy      `bson:"allergies" json:"allergies"`
	ChronicDiseases    []string           `bson:"chronicDiseases" json:"chronicDiseases"`
	CurrentMedications []string           `bson:"currentMedications" json:"currentMedications"`
	ConsentGiven       bool               `bson:"consentGiven" json:"consentGiven"`
	ConsentDate        time.Time          `bson:"consentDate,omitempty" json:"consentDate,omitempty"`
	Note               string             `bson:"note" json:"note"`
	CreatedBy          string             `bson:"createdBy" json:"-"`
	CreatedDate        time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy          string             `bson:"updatedBy" json:"-"`
	UpdatedDate        time.Time          `bson:"updatedDate" json:"-"`
}
