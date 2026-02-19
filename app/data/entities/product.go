package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DrugInfo struct {
	GenericName       string `bson:"genericName" json:"genericName"`
	DrugType          string `bson:"drugType" json:"drugType"`
	DosageForm        string `bson:"dosageForm" json:"dosageForm"`
	Strength          string `bson:"strength" json:"strength"`
	Indication        string `bson:"indication" json:"indication"`
	Dosage            string `bson:"dosage" json:"dosage"`
	SideEffects       string `bson:"sideEffects" json:"sideEffects"`
	Contraindications string `bson:"contraindications" json:"contraindications"`
	StorageCondition  string `bson:"storageCondition" json:"storageCondition"`
	Manufacturer      string `bson:"manufacturer" json:"manufacturer"`
	RegistrationNo    string `bson:"registrationNo" json:"registrationNo"`
	IsControlled      bool   `bson:"isControlled" json:"isControlled"`
}

type Product struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	NameEn       string             `bson:"nameEn" json:"nameEn"`
	Description  string             `bson:"description" json:"description"`
	Price        float64            `bson:"price" json:"price"`
	CostPrice    float64            `bson:"costPrice" json:"costPrice"`
	Unit         string             `bson:"unit" json:"unit"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	SoldFirst    int                `bson:"soldFirst" json:"soldFirst"`
	SerialNumber string             `bson:"serialNumber" json:"serialNumber"`
	Category     string             `bson:"category"  json:"category"`
	Status       string             `bson:"status"  json:"status"`
	DrugInfo     *DrugInfo          `bson:"drugInfo,omitempty" json:"drugInfo,omitempty"`
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
	SoldFirst     int                `bson:"soldFirst" json:"soldFirst"`
	SerialNumber  string             `bson:"serialNumber" json:"serialNumber"`
	Category      string             `bson:"category"  json:"category"`
	Status        string             `bson:"status"  json:"status"`
	DrugInfo      *DrugInfo          `bson:"drugInfo,omitempty" json:"drugInfo,omitempty"`
	CreatedBy     string             `bson:"createdBy" json:"-"`
	CreatedDate   time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy     string             `bson:"updatedBy" json:"-"`
	UpdatedDate   time.Time          `bson:"updatedDate" json:"-"`
	ProductUnits  []ProductUnit      `bson:"units" json:"units"`
	ProductPrices []ProductPrice     `bson:"prices"  json:"prices"`
	ProductStocks []ProductStock     `bson:"stocks"  json:"stocks"`
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
	Product     Product            `bson:"product" json:"product"`
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
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId      primitive.ObjectID `bson:"unitId" json:"unitId"`
	ReceiveCode string             `bson:"receiveCode" json:"receiveCode"`
	Sequence    int                `bson:"sequence" json:"sequence"`
	LotNumber   string             `bson:"lotNumber" json:"lotNumber"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Price       float64            `bson:"price" json:"price"`
	Import      int                `bson:"import" json:"import"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	ExpireDate  time.Time          `bson:"expireDate" json:"expireDate"`
	ImportDate  time.Time          `bson:"importDate" json:"importDate"`
}

type LowStockProduct struct {
	ProductId    primitive.ObjectID `bson:"_id" json:"productId"`
	Name         string             `bson:"name" json:"name"`
	SerialNumber string             `bson:"serialNumber" json:"serialNumber"`
	Unit         string             `bson:"unit" json:"unit"`
	TotalStock   int                `bson:"totalStock" json:"totalStock"`
}

type StockReport struct {
	ProductId    primitive.ObjectID `bson:"_id" json:"productId"`
	Name         string             `bson:"name" json:"name"`
	SerialNumber string             `bson:"serialNumber" json:"serialNumber"`
	Unit         string             `bson:"unit" json:"unit"`
	TotalStock   int                `bson:"totalStock" json:"totalStock"`
	TotalCost    float64            `bson:"totalCost" json:"totalCost"`
}

type ProductHistory struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	Type        string             `bson:"type" json:"type"`
	Description string             `bson:"description" json:"description"`
	Unit        string             `bson:"unit" json:"unit"`
	Import      int                `bson:"import" json:"import"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Price       float64            `bson:"price" json:"price"`
	Balance     int                `bson:"balance" json:"balance"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
}
