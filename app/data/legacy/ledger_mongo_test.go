package legacy

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// newLedger builds the Stock ledger over the test database, with sequential codes.
func newLedger(database *mongo.Database) *ledger.Ledger {
	var seq atomic.Int64
	codes := func(_ context.Context, _ string, prefix string) (string, error) {
		return fmt.Sprintf("%s%04d", prefix, seq.Add(1)), nil
	}
	return ledger.New(database.Client(), database, codes)
}

// seedProduct inserts a Product with its main Unit and returns their ids.
func seedProduct(t *testing.T, database *mongo.Database) (product, unit primitive.ObjectID) {
	t.Helper()
	product, unit = primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, database, "products", bson.M{"_id": product, "name": "Paracetamol", "unit": "TAB", "soldFirst": 0})
	insertDoc(t, database, "product_units", bson.M{"_id": unit, "productId": product, "unit": "TAB", "costPrice": 3.0, "isMain": true})
	return product, unit
}

func stockQuantity(t *testing.T, database *mongo.Database, id primitive.ObjectID) int {
	t.Helper()
	var got entities.ProductStock
	if err := database.Collection("product_stocks").FindOne(context.Background(), bson.M{"_id": id}).Decode(&got); err != nil {
		t.Fatal(err)
	}
	return got.Quantity
}

func TestMigrate_AMigratedReceiveCannotBeImportedAgain(t *testing.T) {
	database := newTestDB(t)
	product, unit := seedProduct(t, database)
	receiveID, stockID := primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "code": "RC_0001", "totalCost": 30.0})
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": product, "costPrice": 3.0, "quantity": 10})
	insertDoc(t, database, "product_stocks", bson.M{"_id": stockID, "productId": product, "unitId": unit, "receiveCode": "RC_0001", "import": 10, "quantity": 10})

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	if got := findReceive(t, database, receiveID); got["branchId"] != report.HQ || got["status"] != constant.IMPORTED {
		t.Fatalf("receive not migrated to HQ as IMPORTED: %v", got)
	}
	_, err = newLedger(database).ImportReceive(context.Background(), receiveID.Hex(), report.HQ.Hex(), "tester")
	var rejected *ledger.Rejected
	if !errors.As(err, &rejected) || rejected.Reason != "receive already imported" {
		t.Fatalf("import error = %v, want rejection %q", err, "receive already imported")
	}
	if q := stockQuantity(t, database, stockID); q != 10 {
		t.Errorf("stock quantity = %d, want 10", q)
	}
}

func TestMigrate_ASaleDrawsMigratedStock(t *testing.T) {
	database := newTestDB(t)
	product, unit := seedProduct(t, database)
	stockID := primitive.NewObjectID()
	insertDoc(t, database, "product_prices", bson.M{"_id": primitive.NewObjectID(), "productId": product, "unitId": unit, "customerType": "General", "price": 10.0})
	insertDoc(t, database, "product_stocks", bson.M{"_id": stockID, "productId": product, "unitId": unit, "sequence": 1, "import": 10, "quantity": 10, "costPrice": 3.0})

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	sold, err := newLedger(database).Sell(context.Background(), request.Sale{
		SaleId: "legacy-sale", Type: "CASH", BranchId: report.HQ.Hex(), CreatedBy: "cashier",
		Items:    []request.SaleLine{{ProductId: product.Hex(), UnitId: unit.Hex(), Quantity: 3, PriceType: "General"}},
		Payments: []request.OrderPayment{{Amount: 30, Type: "CASH"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if q := stockQuantity(t, database, stockID); q != 7 {
		t.Errorf("stock quantity = %d, want 7", q)
	}
	var p entities.Product
	if err := database.Collection("products").FindOne(context.Background(), bson.M{"_id": product}).Decode(&p); err != nil {
		t.Fatal(err)
	}
	if p.SoldFirst != 0 {
		t.Errorf("soldFirst = %d, want 0", p.SoldFirst)
	}
	if len(sold.Lines) != 1 || len(sold.Lines[0].Stocks) != 1 || sold.Lines[0].Stocks[0].StockId != stockID.Hex() || sold.Lines[0].Stocks[0].Quantity != 3 {
		t.Errorf("line stocks = %+v, want 3 from %s", sold.Lines, stockID.Hex())
	}
}
