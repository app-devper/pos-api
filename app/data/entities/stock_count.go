package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockCountItem struct {
	ProductId       primitive.ObjectID `bson:"productId" json:"productId"`
	StockId         primitive.ObjectID `bson:"stockId" json:"stockId"`
	SystemQuantity  int                `bson:"systemQuantity" json:"systemQuantity"`
	CountedQuantity int                `bson:"countedQuantity" json:"countedQuantity"`
	Delta           int                `bson:"delta" json:"delta"`
}

type StockCount struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	CountNo     string             `bson:"countNo" json:"countNo"`
	Note        string             `bson:"note,omitempty" json:"note,omitempty"`
	Items       []StockCountItem   `bson:"items" json:"items"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
}

// StockCountLine is a Count Line as read back: the stored Line plus what its
// Product, Unit and Stock are called at read time. A name that cannot be
// resolved any more (its Product or Stock is gone) is empty.
type StockCountLine struct {
	StockCountItem `bson:",inline"`
	ProductName    string `bson:"productName" json:"productName"`
	Unit           string `bson:"unit" json:"unit"`
	LotNumber      string `bson:"lotNumber" json:"lotNumber"`
}

// StockCountDetail is a StockCount as read back, its Lines named.
type StockCountDetail struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	CountNo     string             `bson:"countNo" json:"countNo"`
	Note        string             `bson:"note,omitempty" json:"note,omitempty"`
	Items       []StockCountLine   `bson:"items" json:"items"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
}
