package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildProductStockSequenceFilterForListIncludesBranchId(t *testing.T) {
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
