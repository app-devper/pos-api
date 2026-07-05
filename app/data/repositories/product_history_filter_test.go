package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildProductHistoryFilterIncludesBranchId(t *testing.T) {
	productID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	filter, err := buildProductHistoryFilter(productID, branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filter["productId"] != productID {
		t.Fatalf("expected productId %s, got %+v", productID.Hex(), filter["productId"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}

func TestBuildProductHistoryFilterRejectsInvalidBranchId(t *testing.T) {
	productID := primitive.NewObjectID()

	if _, err := buildProductHistoryFilter(productID, "invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branchId error")
	}
}
