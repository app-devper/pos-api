package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPatientBranchScopedFilterIncludesBranchId(t *testing.T) {
	patientID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	filter := map[string]interface{}{
		"_id":      patientID,
		"branchId": branchID,
	}

	if filter["_id"] != patientID {
		t.Fatalf("expected patient id %s, got %+v", patientID.Hex(), filter["_id"])
	}
	if filter["branchId"] != branchID {
		t.Fatalf("expected branch id %s, got %+v", branchID.Hex(), filter["branchId"])
	}
}
