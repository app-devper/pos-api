package repositories

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// inBranch matches the document with id only if it belongs to branchId. A
// read by id takes the caller's branch, so another branch's document is
// mongo.ErrNoDocuments, the same as one that does not exist: no handler has
// to remember to compare branches afterwards.
func inBranch(id string, branchId string) (bson.M, error) {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	branch, err := primitive.ObjectIDFromHex(branchId)
	if err != nil {
		return nil, err
	}
	return bson.M{"_id": objId, "branchId": branch}, nil
}
