package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DispensingItem struct {
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	ProductName string             `bson:"productName" json:"productName"`
	GenericName string             `bson:"genericName" json:"genericName"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Unit        string             `bson:"unit" json:"unit"`
	Dosage      string             `bson:"dosage" json:"dosage"`
	LotNumber   string             `bson:"lotNumber" json:"lotNumber"`
}

type DispensingLog struct {
	Id              primitive.ObjectID `bson:"_id" json:"id"`
	BranchId        primitive.ObjectID `bson:"branchId" json:"branchId"`
	OrderId         primitive.ObjectID `bson:"orderId" json:"orderId"`
	PatientId       primitive.ObjectID `bson:"patientId" json:"patientId"`
	Items           []DispensingItem   `bson:"items" json:"items"`
	PharmacistName  string             `bson:"pharmacistName" json:"pharmacistName"`
	LicenseNo       string             `bson:"licenseNo" json:"licenseNo"`
	Note            string             `bson:"note" json:"note"`
	CreatedBy       string             `bson:"createdBy" json:"-"`
	CreatedDate     time.Time          `bson:"createdDate" json:"createdDate"`
}
