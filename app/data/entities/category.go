package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Value       string             `bson:"value" json:"value"`
	Default     bool               `bson:"default" json:"default"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}
