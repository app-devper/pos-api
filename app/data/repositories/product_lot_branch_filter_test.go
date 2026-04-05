package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildProductLotBranchScopeFilterForHQIncludesLegacyLots(t *testing.T) {
	hqID := primitive.NewObjectID()
	filter := buildProductLotBranchScopeFilter(hqID, hqID)

	orItems, ok := filter["$or"].(bson.A)
	if !ok || len(orItems) != 3 {
		t.Fatalf("expected HQ filter to include legacy lot fallback, got %+v", filter)
	}
}

func TestBuildProductLotBranchScopeFilterForBranchUsesExactBranch(t *testing.T) {
	hqID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	filter := buildProductLotBranchScopeFilter(branchID, hqID)

	if filter["branchId"] != branchID {
		t.Fatalf("expected exact branchId filter, got %+v", filter)
	}
}
