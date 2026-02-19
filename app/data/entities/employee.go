package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	UserId      string             `bson:"userId" json:"userId"`
	Role        string             `bson:"role" json:"role"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type EmployeeDetail struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	UserId      string             `bson:"userId" json:"userId"`
	Role        string             `bson:"role" json:"role"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
	Branch      Branch             `bson:"branch" json:"branch"`
}
