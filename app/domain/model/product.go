package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Product struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	NameEn       string             `bson:"nameEn" json:"nameEn"`
	Description  string             `bson:"description" json:"description"`
	Price        float64            `bson:"price" json:"price"`
	CostPrice    float64            `bson:"costPrice" json:"costPrice"`
	Unit         string             `bson:"unit" json:"unit"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	SerialNumber string             `bson:"serialNumber" json:"serialNumber"`
	Category     string             `bson:"category"  json:"category"`
	CreatedBy    string             `bson:"createdBy" json:"-"`
	CreatedDate  time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string             `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time          `bson:"updatedDate" json:"-"`
}

type ProductLot struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	LotNumber   string             `bson:"lotNumber" json:"lotNumber"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	ExpireDate  string             `bson:"expireDate" json:"expireDate"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type ProductPrice struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	ProductId     primitive.ObjectID `bson:"productId" json:"productId"`
	CustomerId    primitive.ObjectID `bson:"customerId" json:"customerId"`
	CustomerPrice float64            `bson:"customerPrice" json:"customerPrice"`
	CreatedBy     string             `bson:"createdBy" json:"-"`
	CreatedDate   time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy     string             `bson:"updatedBy" json:"-"`
	UpdatedDate   time.Time          `bson:"updatedDate" json:"-"`
}

type ProductPriceDetail struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	ProductId     primitive.ObjectID `bson:"productId" json:"productId"`
	CustomerId    primitive.ObjectID `bson:"customerId" json:"customerId"`
	CustomerPrice float64            `bson:"customerPrice" json:"customerPrice"`
	Customer      Customer           `bson:"customer" json:"customer"`
	Product       Product            `bson:"product" json:"product"`
	CreatedBy     string             `bson:"createdBy" json:"-"`
	CreatedDate   time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy     string             `bson:"updatedBy" json:"-"`
	UpdatedDate   time.Time          `bson:"updatedDate" json:"-"`
}
