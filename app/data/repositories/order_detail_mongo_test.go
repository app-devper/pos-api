package repositories

import (
	"testing"
	"time"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func insertDetailLine(f *saleFixture, order primitive.ObjectID, product primitive.ObjectID, price float64, status any) primitive.ObjectID {
	id := primitive.NewObjectID()
	doc := bson.M{
		"_id": id, "branchId": f.branchId, "orderId": order, "productId": product, "unitId": f.unit,
		"quantity": 1, "price": price, "costPrice": price / 2, "discount": 0, "createdDate": time.Now(),
	}
	if status != nil {
		doc["status"] = status
	}
	f.insert("order_items", doc)
	return id
}

func TestGetOrderDetailKeepsALineWhoseProductIsGone(t *testing.T) {
	f := newSaleFixture(t)
	order := primitive.NewObjectID()
	f.insert("orders", bson.M{"_id": order, "branchId": f.branchId, "code": "OD_000000000001", "status": constant.CONFIRMED, "total": 30})
	kept := insertDetailLine(f, order, f.product, 10, constant.CONFIRMED)
	gone := insertDetailLine(f, order, primitive.NewObjectID(), 20, constant.CONFIRMED)

	detail, err := f.orders.GetOrderDetailById(order.Hex())
	if err != nil {
		t.Fatal(err)
	}

	byId := map[primitive.ObjectID]float64{}
	for _, line := range detail.Items {
		byId[line.Id] = line.Price
	}
	if len(detail.Items) != 2 || byId[kept] != 10 || byId[gone] != 20 {
		t.Fatalf("lines %+v", detail.Items)
	}
}

func TestGetOrderDetailGivesEveryLineAStatus(t *testing.T) {
	f := newSaleFixture(t)
	order := primitive.NewObjectID()
	f.insert("orders", bson.M{"_id": order, "branchId": f.branchId, "code": "OD_000000000001", "status": constant.CONFIRMED})
	missing := insertDetailLine(f, order, f.product, 10, nil)
	empty := insertDetailLine(f, order, f.product, 10, "")
	active := insertDetailLine(f, order, f.product, 10, constant.ACTIVE)
	cancelled := insertDetailLine(f, order, f.product, 10, constant.CANCELLED)

	detail, err := f.orders.GetOrderDetailById(order.Hex())
	if err != nil {
		t.Fatal(err)
	}

	got := map[primitive.ObjectID]string{}
	for _, line := range detail.Items {
		got[line.Id] = line.Status
	}
	want := map[primitive.ObjectID]string{
		missing: constant.CONFIRMED, empty: constant.CONFIRMED, active: constant.ACTIVE, cancelled: constant.CANCELLED,
	}
	for id, status := range want {
		if got[id] != status {
			t.Fatalf("line %s status %q, want %q (all %v)", id.Hex(), got[id], status, got)
		}
	}
}
