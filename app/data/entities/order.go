package entities

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	BranchId     primitive.ObjectID `bson:"branchId" json:"branchId"`
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
	Discount     float64            `bson:"discount" json:"discount"`
	Type         string             `bson:"type" json:"type"`
}

type OrderDetail struct {
	Id           primitive.ObjectID       `bson:"_id" json:"id"`
	BranchId     primitive.ObjectID       `bson:"branchId" json:"branchId"`
	Code         string                   `bson:"code" json:"code"`
	CustomerCode string                   `bson:"customerCode" json:"customerCode"`
	CustomerName string                   `bson:"customerName" json:"customerName"`
	Status       string                   `bson:"status" json:"status"`
	CreatedBy    string                   `bson:"createdBy" json:"-"`
	CreatedDate  time.Time                `bson:"createdDate" json:"createdDate"`
	UpdatedBy    string                   `bson:"updatedBy" json:"-"`
	UpdatedDate  time.Time                `bson:"updatedDate" json:"-"`
	Total        float64                  `bson:"total" json:"total"`
	TotalCost    float64                  `bson:"totalCost" json:"totalCost"`
	Discount     float64                  `bson:"discount" json:"discount"`
	Type         string                   `bson:"type" json:"type"`
	Items        []OrderItemProductDetail `json:"items"`
	Payment      Payment                  `json:"payment"`
}

type OrderItemStock struct {
	Quantity int    `json:"quantity" binding:"required"`
	StockId  string `json:"stockId"`
}

type OrderItem struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	OrderId     primitive.ObjectID `bson:"orderId" json:"orderId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId      primitive.ObjectID `bson:"unitId" json:"unitId"`
	Stocks      []OrderItemStock   `bson:"stocks" json:"stocks"`
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
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	OrderId     primitive.ObjectID `bson:"orderId" json:"orderId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId      primitive.ObjectID `bson:"unitId" json:"unitId"`
	Stocks      []OrderItemStock   `bson:"stocks" json:"stocks"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Price       float64            `bson:"price" json:"price"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Discount    float64            `bson:"discount" json:"discount"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
	Product     Product            `bson:"product" json:"product"`
}

type OrderItemOrderDetail struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	BranchId    primitive.ObjectID `bson:"branchId" json:"branchId"`
	OrderId     primitive.ObjectID `bson:"orderId" json:"orderId"`
	ProductId   primitive.ObjectID `bson:"productId" json:"productId"`
	UnitId      primitive.ObjectID `bson:"unitId" json:"unitId"`
	Stocks      []OrderItemStock   `bson:"stocks" json:"stocks"`
	Quantity    int                `bson:"quantity" json:"quantity"`
	Price       float64            `bson:"price" json:"price"`
	CostPrice   float64            `bson:"costPrice" json:"costPrice"`
	Discount    float64            `bson:"discount" json:"discount"`
	CreatedBy   string             `bson:"createdBy" json:"-"`
	CreatedDate time.Time          `bson:"createdDate" json:"createdDate"`
	UpdatedBy   string             `bson:"updatedBy" json:"-"`
	UpdatedDate time.Time          `bson:"updatedDate" json:"-"`
	Order       Order              `bson:"order" json:"order"`
}

type OrderSummary struct {
	TotalOrders  int     `bson:"totalOrders" json:"totalOrders"`
	TotalRevenue float64 `bson:"totalRevenue" json:"totalRevenue"`
	TotalCost    float64 `bson:"totalCost" json:"totalCost"`
	TotalProfit  float64 `bson:"totalProfit" json:"totalProfit"`
}

type OrderDailyChart struct {
	Date         string  `bson:"_id" json:"date"`
	TotalOrders  int     `bson:"totalOrders" json:"totalOrders"`
	TotalRevenue float64 `bson:"totalRevenue" json:"totalRevenue"`
	TotalCost    float64 `bson:"totalCost" json:"totalCost"`
	TotalProfit  float64 `bson:"totalProfit" json:"totalProfit"`
}

func (item OrderItemProductDetail) GetMessage() string {
	return fmt.Sprintf("%s จำนวน %d %s ราคา %.2f บาท", item.Product.Name, item.Quantity, item.Product.Unit, item.Price)
}
