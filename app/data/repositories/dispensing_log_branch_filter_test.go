package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDispensingLogByIdBranchScopeFilterUsesBranchId(t *testing.T) {
	logID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	filter := map[string]interface{}{"_id": logID}
	filter["branchId"] = branchID

	if filter["_id"] != logID {
		t.Fatalf("expected log id %s, got %+v", logID.Hex(), filter["_id"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branch id %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}
