package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Supplier struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	ClientId    string             `bson:"clientId" json:"clientId"`
	Name        string             `bson:"name" json:"name"`
	Address     string             `bson:"address" json:"address"`
	Phone       string             `bson:"phone" json:"phone"`
	TaxId       string             `bson:"taxId" json:"taxId"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}
