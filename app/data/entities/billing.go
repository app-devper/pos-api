package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Billing struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	BranchId     primitive.ObjectID `bson:"branchId" json:"branchId"`
	CustomerCode string             `bson:"customerCode" json:"customerCode"`
	CustomerName string             `bson:"customerName" json:"customerName"`
	Code         string             `bson:"code" json:"code"`
	OrderIds     []primitive.ObjectID `bson:"orderIds" json:"orderIds"`
	TotalAmount  float64            `bson:"totalAmount" json:"totalAmount"`
	Note         string             `bson:"note" json:"note"`
	Status       string             `bson:"status" json:"status"`
	DueDate      time.Time          `bson:"dueDate" json:"dueDate"`
	CreatedBy    string             `bson:"createdBy" json:"-"`
	CreatedDate  time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string             `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time          `bson:"updatedDate" json:"-"`
}
