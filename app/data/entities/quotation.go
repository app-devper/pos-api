package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Quotation struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	BranchId     primitive.ObjectID `bson:"branchId" json:"branchId"`
	CustomerCode string             `bson:"customerCode" json:"customerCode"`
	CustomerName string             `bson:"customerName" json:"customerName"`
	Code         string             `bson:"code" json:"code"`
	Items        []QuotationItem    `bson:"items" json:"items"`
	TotalAmount  float64            `bson:"totalAmount" json:"totalAmount"`
	Discount     float64            `bson:"discount" json:"discount"`
	Note         string             `bson:"note" json:"note"`
	Status       string             `bson:"status" json:"status"`
	ValidUntil   time.Time          `bson:"validUntil" json:"validUntil"`
	CreatedBy    string             `bson:"createdBy" json:"-"`
	CreatedDate  time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string             `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time          `bson:"updatedDate" json:"-"`
}

type QuotationItem struct {
	ProductId primitive.ObjectID `bson:"productId" json:"productId"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     float64            `bson:"price" json:"price"`
	Total     float64            `bson:"total" json:"total"`
}
