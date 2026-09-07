package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockAdjustment struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	Code        string             `bson:"code" json:"code"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	StockId     primitive.ObjectID `bson:"stockId" json:"stockId"`
	Reason      string             `bson:"reason" json:"reason"`
	Note        string             `bson:"note,omitempty" json:"note,omitempty"`
	Delta       int                `bson:"delta" json:"delta"`
	Before      int                `bson:"before" json:"before"`
	After       int                `bson:"after" json:"after"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
}
