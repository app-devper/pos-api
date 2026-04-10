package repositories

import (
	"testing"

	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildStockTransferReservationFilterIncludesBranchAndProductScope(t *testing.T) {
	branchID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	filter, err := buildStockTransferReservationFilter(branchID.Hex(), request.StockTransferItem{
		ProductId: productID.Hex(),
		StockId:   stockID.Hex(),
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("expected filter to build, got %v", err)
	}

	if got := filter["_id"]; got != stockID {
		t.Fatalf("expected stock id %v, got %v", stockID, got)
	}
	if got := filter["branchId"]; got != branchID {
		t.Fatalf("expected branch id %v, got %v", branchID, got)
	}
	if got := filter["productId"]; got != productID {
		t.Fatalf("expected product id %v, got %v", productID, got)
	}
	quantityFilter, ok := filter["quantity"].(bson.M)
	if !ok {
		t.Fatalf("expected quantity filter bson.M, got %T", filter["quantity"])
	}
	if got := quantityFilter["$gte"]; got != 3 {
		t.Fatalf("expected quantity threshold 3, got %v", got)
	}
}

func TestBuildStockTransferReservationFilterRejectsInvalidIDs(t *testing.T) {
	if _, err := buildStockTransferReservationFilter("invalid-branch", request.StockTransferItem{}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
	if _, err := buildStockTransferReservationFilter(primitive.NewObjectID().Hex(), request.StockTransferItem{
		ProductId: "invalid-product",
		StockId:   primitive.NewObjectID().Hex(),
		Quantity:  1,
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
	if _, err := buildStockTransferReservationFilter(primitive.NewObjectID().Hex(), request.StockTransferItem{
		ProductId: primitive.NewObjectID().Hex(),
		StockId:   "invalid-stock",
		Quantity:  1,
	}); err == nil {
		t.Fatal("expected invalid stock id error")
	}
}
