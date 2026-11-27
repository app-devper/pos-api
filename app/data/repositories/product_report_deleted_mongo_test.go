package repositories

import (
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// A Product deleted before deletes zeroed its Stock still holds goods; the
// stock reports must not list it.
func TestStockReportsSkipADeletedProductThatStillHoldsStock(t *testing.T) {
	f := newSaleFixture(t)
	f.stock(1, 4, 3)
	gone, goneUnit := primitive.NewObjectID(), primitive.NewObjectID()
	f.insert("products", bson.M{"_id": gone, "name": "Gone", "deletedDate": time.Now()})
	f.insert("product_units", entities.ProductUnit{Id: goneUnit, ProductId: gone, Unit: "BOX"})
	f.insert("product_stocks", entities.ProductStock{Id: primitive.NewObjectID(), BranchId: f.branchId, ProductId: gone, UnitId: goneUnit, Sequence: 1, Quantity: 2, CostPrice: 5})
	stocks := newProductStockEntity(&db.Resource{Client: f.pos.Client(), PosDb: f.pos})

	report, err := stocks.GetStockReport(f.branchId.Hex())
	if err != nil {
		t.Fatal(err)
	}
	low, err := stocks.GetLowStockProducts(100, f.branchId.Hex())
	if err != nil {
		t.Fatal(err)
	}

	if len(report) != 1 || report[0].ProductId != f.product {
		t.Fatalf("stock report %+v, want only the live product", report)
	}
	if len(low) != 1 || low[0].ProductId != f.product {
		t.Fatalf("low stock %+v, want only the live product", low)
	}
}
