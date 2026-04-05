package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildProductStockSequenceFilterIncludesBranchId(t *testing.T) {
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	filter, err := buildProductStockSequenceFilter(productID.Hex(), unitID.Hex(), branchID.Hex())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if filter["productId"] != productID {
		t.Fatalf("expected productId %s, got %+v", productID.Hex(), filter["productId"])
	}
	if filter["unitId"] != unitID {
		t.Fatalf("expected unitId %s, got %+v", unitID.Hex(), filter["unitId"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}

func TestBuildProductStockSequenceFilterRejectsInvalidIDs(t *testing.T) {
	if _, err := buildProductStockSequenceFilter("invalid-id", primitive.NewObjectID().Hex(), ""); err == nil {
		t.Fatal("expected invalid product id error")
	}
	if _, err := buildProductStockSequenceFilter(primitive.NewObjectID().Hex(), "invalid-id", ""); err == nil {
		t.Fatal("expected invalid unit id error")
	}
	if _, err := buildProductStockSequenceFilter(primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex(), "invalid-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestBuildReceiveProductStockSequenceMatchIncludesBranchId(t *testing.T) {
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	match := buildReceiveProductStockSequenceMatch(productID, unitID, branchID)

	if match["productId"] != productID {
		t.Fatalf("expected productId %s, got %+v", productID.Hex(), match["productId"])
	}
	if match["unitId"] != unitID {
		t.Fatalf("expected unitId %s, got %+v", unitID.Hex(), match["unitId"])
	}
	if match["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), match["branchId"])
	}
}
