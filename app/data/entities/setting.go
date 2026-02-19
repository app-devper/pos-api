package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Setting struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`
	BranchId       primitive.ObjectID `bson:"branchId" json:"branchId"`
	ReceiptFooter  string             `bson:"receiptFooter" json:"receiptFooter"`
	CompanyName    string             `bson:"companyName" json:"companyName"`
	CompanyAddress string             `bson:"companyAddress" json:"companyAddress"`
	CompanyPhone   string             `bson:"companyPhone" json:"companyPhone"`
	CompanyTaxId   string             `bson:"companyTaxId" json:"companyTaxId"`
	LogoUrl        string             `bson:"logoUrl" json:"logoUrl"`
	ShowCredit     bool               `bson:"showCredit" json:"showCredit"`
	PromptPayId    string             `bson:"promptPayId" json:"promptPayId"`
	UpdatedBy      string             `bson:"updatedBy" json:"-"`
	UpdatedDate    time.Time          `bson:"updatedDate" json:"-"`
}
