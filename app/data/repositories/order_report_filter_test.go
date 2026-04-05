package repositories

import (
	"testing"
	"time"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildActiveOrderAnalyticsMatchFilterIncludesActiveStatusAndBranch(t *testing.T) {
	branchID := primitive.NewObjectID()
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	filter, err := buildActiveOrderAnalyticsMatchFilter(startDate, endDate, branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filter["status"] != constant.ACTIVE {
		t.Fatalf("expected ACTIVE status filter, got %+v", filter["status"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), filter["branchId"])
	}

	dateFilter, ok := filter["createdDate"].(bson.M)
	if !ok {
		t.Fatalf("expected createdDate bson.M, got %+v", filter["createdDate"])
	}
	if dateFilter["$gt"] != startDate || dateFilter["$lt"] != endDate {
		t.Fatalf("unexpected date filter: %+v", dateFilter)
	}
}

func TestBuildMonthlyActiveOrderAnalyticsMatchFilterIncludesActiveStatusAndBranch(t *testing.T) {
	branchID := primitive.NewObjectID()
	startDate := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)

	filter, err := buildMonthlyActiveOrderAnalyticsMatchFilter(startDate, branchID.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filter["status"] != constant.ACTIVE {
		t.Fatalf("expected ACTIVE status filter, got %+v", filter["status"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branchId %s, got %+v", branchID.Hex(), filter["branchId"])
	}

	dateFilter, ok := filter["createdDate"].(bson.M)
	if !ok {
		t.Fatalf("expected createdDate bson.M, got %+v", filter["createdDate"])
	}
	if dateFilter["$gte"] != startDate {
		t.Fatalf("unexpected date filter: %+v", dateFilter)
	}
}
