package repositories

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetProductsByIdsExcludesLogicallyDeletedProducts(t *testing.T) {
	filter, err := buildGetProductsByIDsFilter([]string{
		"507f1f77bcf86cd799439011",
		"507f1f77bcf86cd799439012",
	})
	if err != nil {
		t.Fatalf("expected filter build to succeed, got %v", err)
	}

	deletedDate, ok := filter["deletedDate"].(bson.M)
	if !ok {
		t.Fatalf("expected deletedDate filter to be present, got %+v", filter)
	}
	if deletedDate["$exists"] != false {
		t.Fatalf("expected deletedDate exists=false filter, got %+v", deletedDate)
	}

	idFilter, ok := filter["_id"].(bson.M)
	if !ok {
		t.Fatalf("expected _id filter, got %+v", filter["_id"])
	}
	objIDs, ok := idFilter["$in"].([]primitive.ObjectID)
	if !ok {
		t.Fatalf("expected object id slice in $in filter, got %+v", idFilter["$in"])
	}
	if len(objIDs) != 2 {
		t.Fatalf("expected 2 object ids in filter, got %d", len(objIDs))
	}
}
