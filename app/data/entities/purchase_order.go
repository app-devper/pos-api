package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseOrder struct {
	Id          primitive.ObjectID  `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID  `bson:"branchId" json:"branchId"`
	SupplierId  primitive.ObjectID  `bson:"supplierId" json:"supplierId"`
	Code        string              `bson:"code" json:"code"`
	Reference   string              `bson:"reference" json:"reference"`
	Items       []PurchaseOrderItem `bson:"items" json:"items"`
	TotalCost   float64             `bson:"totalCost" json:"totalCost"`
	Note        string              `bson:"note" json:"note"`
	Status      string              `bson:"status" json:"status"`
	DueDate     time.Time           `bson:"dueDate" json:"dueDate"`
	CreatedBy   string              `bson:"createdBy" json:"-"`
	CreatedDate time.Time           `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string              `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time           `bson:"updatedDate" json:"-"`
}

type PurchaseOrderItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	CostPrice float64            `bson:"costPrice" json:"costPrice"`
	Total     float64            `bson:"total" json:"total"`
}
