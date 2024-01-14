package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Category struct {
	Id                   primitive.ObjectID `bson:"_id" json:"id"`
	Name                 string             `bson:"name" json:"name"`
	Value                string             `bson:"value" json:"value"`
	Description          string             `bson:"description" json:"description"`
	Default              bool               `bson:"default" json:"default"`
	RequireCustomerOrder bool               `bson:"requireCustomerOrder" json:"requireCustomerOrder"`
	CreatedBy            string             `bson:"createdBy" json:"-"`
	CreatedDate          time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy            string             `bson:"updatedBy" json:"-"`
	UpdatedDate          time.Time          `bson:"updatedDate" json:"-"`
}
