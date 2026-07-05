package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildProductStockSequenceBranchFilterIncludesBranchId(t *testing.T) {
	branchID := primitive.NewObjectID()

	filter, err := buildProductStockSequenceBranchFilter(branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}

func TestBuildProductStockSequenceBranchFilterRejectsInvalidBranchId(t *testing.T) {
	if _, err := buildProductStockSequenceBranchFilter("invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}
