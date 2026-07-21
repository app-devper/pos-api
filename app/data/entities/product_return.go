package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductReturnItem struct {
	OrderItemId primitive.ObjectID `bson:"orderItemId" json:"orderItemId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Price       float64            `bson:"price" json:"price"`
	Refund      float64            `bson:"refund" json:"refund"`
}

type ProductReturn struct {
	Id           primitive.ObjectID  `bson:"_id" json:"id"`
	BranchId     primitive.ObjectID  `bson:"branchId" json:"branchId"`
	ReturnNo     string              `bson:"returnNo" json:"returnNo"`
	OrderId      primitive.ObjectID  `bson:"orderId" json:"orderId"`
	CustomerCode string              `bson:"customerCode,omitempty" json:"customerCode,omitempty"`
	Reason       string              `bson:"reason" json:"reason"`
	Note         string              `bson:"note,omitempty" json:"note,omitempty"`
	Items        []ProductReturnItem `bson:"items" json:"items"`
	TotalRefund  float64             `bson:"totalRefund" json:"totalRefund"`
	CreatedBy    string              `bson:"createdBy" json:"-"`
	CreatedDate  time.Time           `bson:"createdDate" json:"createdDate"`
}
