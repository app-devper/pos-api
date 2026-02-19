package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Receive struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	SupplierId  primitive.ObjectID `bson:"supplierId" json:"supplierId"`
	Code        string             `bson:"code" json:"code"`
	Reference   string             `bson:"reference" json:"reference"`
	TotalCost   float64            `bson:"totalCost" json:"totalCost"`
	Items       []ReceiveItem      `bson:"items" json:"items"`
	Status      string             `bson:"status" json:"status"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type ReceiveItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	CostPrice float64            `bson:"costPrice" json:"costPrice"`
	Quantity  int                `bson:"quantity" json:"quantity"`
}
