package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Receive struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	SupplierId  primitive.ObjectID `bson:"supplierId" json:"supplierId"`
	Code        string             `bson:"code" json:"code"`
	Reference   string             `bson:"reference" json:"reference"`
	TotalCost   float64            `bson:"totalCost" json:"totalCost"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type ReceiveItem struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	ReceiveId   primitive.ObjectID `bson:"receiveId" json:"receiveId"`
	LotId       primitive.ObjectID `bson:"lotId" json:"lotId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}
