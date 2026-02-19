package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeliveryOrder struct {
	Id          primitive.ObjectID   `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID   `bson:"branchId" json:"branchId"`
	OrderId     primitive.ObjectID   `bson:"orderId" json:"orderId"`
	Code        string               `bson:"code" json:"code"`
	CustomerCode string              `bson:"customerCode" json:"customerCode"`
	CustomerName string              `bson:"customerName" json:"customerName"`
	Address     string               `bson:"address" json:"address"`
	Items       []DeliveryOrderItem  `bson:"items" json:"items"`
	Note        string               `bson:"note" json:"note"`
	Status      string               `bson:"status" json:"status"`
	DeliveryDate time.Time           `bson:"deliveryDate" json:"deliveryDate"`
	CreatedBy   string               `bson:"createdBy" json:"-"`
	CreatedDate time.Time            `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string               `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time            `bson:"updatedDate" json:"-"`
}

type DeliveryOrderItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     float64            `bson:"price" json:"price"`
}
