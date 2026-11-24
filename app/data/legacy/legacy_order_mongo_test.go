package legacy

import (
	"context"
	"testing"
	"time"

	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	legacyStockBefore = 7 // 10 imported, 3 sold on the 2024 till
	legacyLineQty     = 3
)

// legacyOrder is a 2024 sale as that build stored it: no branchId anywhere,
// an ACTIVE Order, a Line with no status naming the Stock it drew from in
// stocks[] (fields stockid / quantity, the driver's default names), and an
// ACTIVE payment. Shapes from b165b4d app/data/entities.
type legacyOrder struct {
	order, line, stock primitive.ObjectID
}

func seedLegacyOrder(t *testing.T, database *mongo.Database) legacyOrder {
	t.Helper()
	product, unit := seedProduct(t, database)
	o := legacyOrder{order: primitive.NewObjectID(), line: primitive.NewObjectID(), stock: primitive.NewObjectID()}
	now := time.Now()
	insertDoc(t, database, "product_stocks", bson.M{"_id": o.stock, "productId": product, "unitId": unit,
		"receiveCode": "RC_0001", "sequence": 1, "costPrice": 3.0, "price": 10.0, "import": 10, "quantity": legacyStockBefore, "importDate": now})
	insertDoc(t, database, "orders", bson.M{"_id": o.order, "code": "OD_202401150001", "status": constant.ACTIVE,
		"total": 30.0, "totalCost": 9.0, "type": "CASH", "createdBy": "cashier", "createdDate": now})
	insertDoc(t, database, "order_items", bson.M{"_id": o.line, "orderId": o.order, "productId": product, "unitId": unit,
		"stocks":   bson.A{bson.M{"quantity": legacyLineQty, "stockid": o.stock.Hex()}},
		"quantity": legacyLineQty, "price": 30.0, "costPrice": 9.0, "discount": 0.0, "createdBy": "cashier", "createdDate": now})
	insertDoc(t, database, "payments", bson.M{"_id": primitive.NewObjectID(), "orderId": o.order, "status": constant.ACTIVE,
		"amount": 30.0, "total": 30.0, "change": 0.0, "type": "CASH", "createdBy": "cashier", "createdDate": now})
	return o
}

// migrateToHQ applies the migration and returns the HQ Branch id.
func migrateToHQ(t *testing.T, database *mongo.Database) primitive.ObjectID {
	t.Helper()
	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}
	return report.HQ
}

func orderStatus(t *testing.T, database *mongo.Database, id primitive.ObjectID) string {
	t.Helper()
	var got struct {
		Status string `bson:"status"`
	}
	if err := database.Collection("orders").FindOne(context.Background(), bson.M{"_id": id}).Decode(&got); err != nil {
		t.Fatal(err)
	}
	return got.Status
}

func TestMigrate_AMigrated2024OrderCanBeCancelled(t *testing.T) {
	database := newTestDB(t)
	o := seedLegacyOrder(t, database)
	hq := migrateToHQ(t, database)

	if err := newLedger(database).CancelOrder(context.Background(), o.order.Hex(), hq.Hex(), "manager", "test"); err != nil {
		t.Fatalf("cancel migrated 2024 order: %v", err)
	}

	if q := stockQuantity(t, database, o.stock); q != legacyStockBefore+legacyLineQty {
		t.Errorf("stock quantity = %d, want %d", q, legacyStockBefore+legacyLineQty)
	}
	if s := orderStatus(t, database, o.order); s != constant.CANCELLED {
		t.Errorf("order status = %q, want %q", s, constant.CANCELLED)
	}
}

func TestMigrate_AMigrated2024OrderTakesAPartReturn(t *testing.T) {
	database := newTestDB(t)
	o := seedLegacyOrder(t, database)
	hq := migrateToHQ(t, database)
	const returned = 2

	_, err := newLedger(database).Return(context.Background(), request.ProductReturn{
		OrderId: o.order.Hex(), BranchId: hq.Hex(), CreatedBy: "manager", Reason: "test",
		Items: []request.ProductReturnItem{{OrderItemId: o.line.Hex(), Quantity: returned, Refund: 20}},
	})
	if err != nil {
		t.Fatalf("return part of migrated 2024 line: %v", err)
	}

	if q := stockQuantity(t, database, o.stock); q != legacyStockBefore+returned {
		t.Errorf("stock quantity = %d, want %d", q, legacyStockBefore+returned)
	}
}
