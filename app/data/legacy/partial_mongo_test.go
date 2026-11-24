package legacy

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// refuseBranchIdWhere makes the server refuse to give a branchId to a document
// of coll that matches marker: a write the migration cannot finish.
func refuseBranchIdWhere(t *testing.T, database *mongo.Database, coll string, marker bson.M) {
	t.Helper()
	refuse := bson.M{"$nor": bson.A{bson.M{"$and": bson.A{marker, bson.M{"branchId": bson.M{"$exists": true}}}}}}
	if err := database.CreateCollection(context.Background(), coll, options.CreateCollection().SetValidator(refuse)); err != nil {
		t.Fatal(err)
	}
}

func TestMigrate_AFailedApplyReturnsWhatItWrote(t *testing.T) {
	// Arrange: payments, the fourth collection, refuses the backfill
	database := newTestDB(t)
	refuseBranchIdWhere(t, database, "payments", bson.M{})
	seedLegacy(t, database)
	insert(t, database, "sequences", bson.M{"field": "X"})
	insert(t, database, "sequences", bson.M{"field": "X"})

	// Act
	report, err := Migrate(context.Background(), database, true)

	// Assert
	if err == nil {
		t.Fatal("Migrate succeeded, want the payments write refused")
	}
	if report.HQ.IsZero() || !report.HQCreated {
		t.Errorf("HQ = %s created %v, want the created HQ", report.HQ.Hex(), report.HQCreated)
	}
	if _, ok := findKind(report.Findings, KindDuplicateSequence); !ok {
		t.Errorf("findings = %+v, want the duplicate sequence", report.Findings)
	}
	done := []string{"product_stocks", "orders", "order_items"}
	if len(report.Changes) != len(done) {
		t.Fatalf("changes = %+v, want only %v", report.Changes, done)
	}
	for i, name := range done {
		c := report.Changes[i]
		if c.Collection != name || c.Count != int64(legacyCollections[name]) {
			t.Errorf("change %d = %+v, want %s %d", i, c, name, legacyCollections[name])
		}
	}
}

func TestMigrate_AFailedReceiveKeepsTheReceivesDoneBeforeIt(t *testing.T) {
	// Arrange: the second Receive (by _id) is refused
	database := newTestDB(t)
	refuseBranchIdWhere(t, database, "receives", bson.M{"code": "RC_BAD"})
	good, bad := primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, database, "receives", bson.M{"_id": bad, "code": "RC_BAD"})
	insertDoc(t, database, "receives", bson.M{"_id": good, "code": "RC_GOOD"})
	if good.Hex() > bad.Hex() {
		t.Fatal("test needs the good Receive first by _id")
	}

	// Act
	report, err := Migrate(context.Background(), database, true)

	// Assert
	if err == nil {
		t.Fatal("Migrate succeeded, want the second Receive refused")
	}
	if n := changeFor(report, "receives", "branchId"); n != 1 {
		t.Errorf("receives/branchId = %d, want 1: changes %+v", n, report.Changes)
	}
	if n := changeFor(report, "receives", "status"); n != 1 {
		t.Errorf("receives/status = %d, want 1", n)
	}
	if n := changeCount(report, "product_lots"); n != 0 {
		t.Errorf("product_lots = %d, want 0 (the backfill before receives finished)", n)
	}
}
