package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerHistory struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	CustomerCode string            `bson:"customerCode" json:"customerCode"`
	Type        string             `bson:"type" json:"type"`
	Description string             `bson:"description" json:"description"`
	Reference   string             `bson:"reference" json:"reference"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
}
