package entities

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Order struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	Code         string             `bson:"code" json:"code"`
	CustomerCode string             `bson:"customerCode" json:"customerCode"`
	CustomerName string             `bson:"customerName" json:"customerName"`
	Status       string             `bson:"status" json:"status"`
	CreatedBy    string             `bson:"createdBy" json:"-"`
	CreatedDate  time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string             `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time          `bson:"updatedDate" json:"-"`
	Total        float64            `bson:"total" json:"total"`
	TotalCost    float64            `bson:"totalCost" json:"totalCost"`
	Type         string             `bson:"type" json:"type"`
}

type OrderDetail struct {
	Items   []OrderItemProductDetail `json:"items"`
	Payment Payment                  `json:"payment"`
	Order
}

type OrderItem struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	OrderId     primitive.ObjectID `bson:"orderId" json:"orderId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	StockId     primitive.ObjectID `bson:"stockId" json:"stockId"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Price       float64            `bson:"price" json:"price"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Discount    float64            `bson:"discount" json:"discount"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
}

type OrderItemProductDetail struct {
	Product Product `bson:"product" json:"product"`
	OrderItem
}

type OrderItemOrderDetail struct {
	Order Order `bson:"order" json:"order"`
	OrderItem
}

func (item OrderItemProductDetail) GetMessage() string {
	return fmt.Sprintf("%s จำนวน %d %s ราคา %.2f บาท", item.Product.Name, item.Quantity, item.Product.Unit, item.Price)
}
