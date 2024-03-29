package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Customer struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Code         string             `bson:"code" json:"code"`
	CustomerType string             `bson:"customerType" json:"customerType"`
	Name         string             `bson:"name" json:"name"`
	Address      string             `bson:"address" json:"address"`
	Phone        string             `bson:"phone" json:"phone"`
	Email        string             `bson:"email" json:"email"`
	Status       string             `bson:"status" json:"status"`
	CreatedBy    string             `bson:"createdBy" json:"-"`
	CreatedDate  time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string             `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time          `bson:"updatedDate" json:"-"`
}
