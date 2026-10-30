package repositories

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAReadByIdOnlyFindsTheCallersBranch(t *testing.T) {
	e, pos := newReceiveDocFixture(t)
	id := insertReceive(t, pos, "")
	var doc struct {
		BranchId primitive.ObjectID `bson:"branchId"`
	}
	if err := pos.Collection("receives").FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc); err != nil {
		t.Fatal(err)
	}

	if _, err := e.GetReceiveById(id.Hex(), doc.BranchId.Hex()); err != nil {
		t.Fatalf("own branch: %v", err)
	}
	if _, err := e.GetReceiveById(id.Hex(), primitive.NewObjectID().Hex()); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("another branch: got %v, want mongo.ErrNoDocuments", err)
	}
}
