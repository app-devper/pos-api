package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Promotion struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	Code        string             `bson:"code" json:"code"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Type        string             `bson:"type" json:"type"`
	Value       float64            `bson:"value" json:"value"`
	MinPurchase float64            `bson:"minPurchase" json:"minPurchase"`
	MaxDiscount float64            `bson:"maxDiscount" json:"maxDiscount"`
	ProductIds  []primitive.ObjectID `bson:"productIds" json:"productIds"`
	StartDate   time.Time          `bson:"startDate" json:"startDate"`
	EndDate     time.Time          `bson:"endDate" json:"endDate"`
	Status      string             `bson:"status" json:"status"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}
