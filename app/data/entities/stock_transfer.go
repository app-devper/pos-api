package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockTransfer struct {
	Id             primitive.ObjectID  `bson:"_id" json:"id"`
	FromBranchId   primitive.ObjectID  `bson:"fromBranchId" json:"fromBranchId"`
	ToBranchId     primitive.ObjectID  `bson:"toBranchId" json:"toBranchId"`
	Code           string              `bson:"code" json:"code"`
	Items          []StockTransferItem `bson:"items" json:"items"`
	Note           string              `bson:"note" json:"note"`
	Status         string              `bson:"status" json:"status"`
	CreatedBy      string              `bson:"createdBy" json:"-"`
	CreatedDate    time.Time           `bson:"createdDate" json:"createdDate"`
	UpdatedBy      string              `bson:"updatedBy" json:"-"`
	UpdatedDate    time.Time           `bson:"updatedDate" json:"-"`
}

type StockTransferItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	StockId   string             `bson:"stockId" json:"stockId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
}
