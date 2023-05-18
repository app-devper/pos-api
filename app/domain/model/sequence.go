package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"pos/app/domain/constant"
)

type Sequence struct {
	Id     primitive.ObjectID `bson:"_id" json:"id"`
	Field  string             `bson:"field" json:"field"`
	Value  int                `bson:"value" json:"value"`
	Prefix string             `bson:"prefix" json:"prefix"`
	Format int                `bson:"format" json:"format"`
	Type   string             `bson:"type" json:"type"`
	Date   string             `bson:"date" json:"date"`
}

func (data Sequence) GenerateCode() string {
	var date = ""
	if data.Type == constant.DAILY {
		date = data.Date
	} else if data.Type == constant.MONTHLY {
		date = data.Date[0:6]
	} else if data.Type == constant.YEARLY {
		date = data.Date[0:4]
	}
	return data.Prefix + date + fmt.Sprintf("%0*d", data.Format, data.Value)
}
