package legacy

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestInsert_LeavesTheGivenDocumentAlone(t *testing.T) {
	// Arrange
	db := newTestDB(t)
	doc := bson.M{"type": "CASH"}
	// Act
	id := insert(t, db, "payments", doc)
	// Assert
	if _, ok := doc["_id"]; ok || len(doc) != 1 {
		t.Errorf("insert changed its argument: %v", doc)
	}
	if n := countDocs(t, db, "payments", bson.M{"_id": id, "type": "CASH"}); n != 1 {
		t.Errorf("inserted documents with id %s = %d, want 1", id.Hex(), n)
	}
}
