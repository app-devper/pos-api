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

type ProductDetail struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	Name          string             `bson:"name" json:"name"`
	NameEn        string             `bson:"nameEn" json:"nameEn"`
	Description   string             `bson:"description" json:"description"`
	Price         float64            `bson:"price" json:"price"`
	CostPrice     float64            `bson:"costPrice" json:"costPrice"`
	Unit          string             `bson:"unit" json:"unit"`
	Quantity      int                `bson:"quantity" json:"quantity"`
	SerialNumber  string             `bson:"serialNumber" json:"serialNumber"`
	Category      string             `bson:"category"  json:"category"`
	CreatedDate   time.Time          `bson:"createdDate" json:"createdDate"`
	ProductUnits  []ProductUnit      `bson:"units" json:"units"`
	ProductPrices []ProductPrice     `bson:"prices"  json:"prices"`
}

type ProductLot struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	LotNumber   string             `bson:"lotNumber" json:"lotNumber"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	ExpireDate  time.Time          `bson:"expireDate" json:"expireDate"`
	Notify      bool               `bson:"notify" json:"notify"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type ProductLotDetail struct {
	Product Product `bson:"product" json:"product"`
	ProductLot
}

type ProductUnit struct {
	Id         primitive.ObjectID `bson:"_id" json:"id"`
	ProductId  primitive.ObjectID `bson:"productId" json:"productId"`
	Unit       string             `bson:"unit" json:"unit"`
	Size       int                `bson:"size" json:"size"`
	CostPrice  float64            `bson:"costPrice" json:"costPrice"`
	Volume     float64            `bson:"volume" json:"volume"`
	VolumeUnit string             `bson:"volumeUnit" json:"volumeUnit"`
	Barcode    string             `bson:"barcode" json:"barcode"`
}

type ProductPrice struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	ProductId    primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId       primitive.ObjectID `bson:"unitId" json:"unitId"`
	CustomerType string             `bson:"customerType" json:"customerType"`
	Price        float64            `bson:"price" json:"price"`
}

type ProductStock struct {
	Id         primitive.ObjectID `bson:"_id" json:"id"`
	ProductId  primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId     primitive.ObjectID `bson:"unitId" json:"unitId"`
	Sequence   int                `bson:"sequence" json:"sequence"`
	LotNumber  string             `bson:"lotNumber" json:"lotNumber"`
	CostPrice  float64            `bson:"costPrice" json:"costPrice"`
	Price      float64            `bson:"price" json:"price"`
	Import     int                `bson:"import" json:"import"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	ExpireDate time.Time          `bson:"expireDate" json:"expireDate"`
	ImportDate time.Time          `bson:"importDate" json:"importDate"`
}
