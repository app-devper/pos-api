package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreditNote struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	OrderId     primitive.ObjectID `bson:"orderId" json:"orderId"`
	Code        string             `bson:"code" json:"code"`
	Reason      string             `bson:"reason" json:"reason"`
	Items       []CreditNoteItem   `bson:"items" json:"items"`
	TotalRefund float64            `bson:"totalRefund" json:"totalRefund"`
	Note        string             `bson:"note" json:"note"`
	Status      string             `bson:"status" json:"status"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type CreditNoteItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     float64            `bson:"price" json:"price"`
	StockId   string             `bson:"stockId" json:"stockId"`
}
