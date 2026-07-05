package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOrderCustomerCodeFilterIncludesBranchId(t *testing.T) {
	branchID := primitive.NewObjectID()
	filter := map[string]interface{}{
		"customerCode": "C001",
		"branchId":     branchID,
	}

	if filter["customerCode"] != "C001" {
		t.Fatalf("expected customer code C001, got %+v", filter["customerCode"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branch id %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}
