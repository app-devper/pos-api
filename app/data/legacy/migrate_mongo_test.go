package legacy

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// legacyCollections are the collections seeded by seedLegacy, with how many
// documents each gets.
var legacyCollections = map[string]int{
	"product_stocks":    2,
	"orders":            1,
	"order_items":       2,
	"payments":          1,
	"product_histories": 1,
}

// newTestDB connects to MONGO_TEST_URI and returns a database of its own,
// dropped when the test ends; the test is skipped without a URI.
func newTestDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	database := client.Database(fmt.Sprintf("pos_legacy_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = database.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return database
}

// seedLegacy inserts documents shaped as the 2024 build wrote them: no branchId.
func seedLegacy(t *testing.T, database *mongo.Database) {
	t.Helper()
	for name, count := range legacyCollections {
		for i := 0; i < count; i++ {
			doc := bson.M{"_id": primitive.NewObjectID(), "createdDate": time.Now(), "note": name}
			if _, err := database.Collection(name).InsertOne(context.Background(), doc); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func countDocs(t *testing.T, database *mongo.Database, name string, filter bson.M) int64 {
	t.Helper()
	n, err := database.Collection(name).CountDocuments(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func changeCount(report Report, collection string) int64 {
	for _, c := range report.Changes {
		if c.Collection == collection && c.Field == "branchId" {
			return c.Count
		}
	}
	return -1
}

func TestMigrate_ReportModeWritesNothing(t *testing.T) {
	database := newTestDB(t)
	seedLegacy(t, database)

	report, err := Migrate(context.Background(), database, false)
	if err != nil {
		t.Fatal(err)
	}

	for name, want := range legacyCollections {
		if got := changeCount(report, name); got != int64(want) {
			t.Errorf("%s change count = %d, want %d", name, got, want)
		}
		if n := countDocs(t, database, name, bson.M{"branchId": bson.M{"$exists": true}}); n != 0 {
			t.Errorf("%s has %d documents with branchId after a report run", name, n)
		}
	}
	if report.HQCreated {
		t.Error("HQCreated = true in report mode")
	}
	if n := countDocs(t, database, "branches", bson.M{}); n != 0 {
		t.Errorf("branches count = %d, want 0", n)
	}
}

func TestMigrate_ApplyGivesLegacyDocumentsTheHQBranch(t *testing.T) {
	database := newTestDB(t)
	seedLegacy(t, database)

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	if !report.HQCreated {
		t.Error("HQCreated = false, want true")
	}
	if n := countDocs(t, database, "branches", bson.M{"code": "HQ", "status": "ACTIVE", "name": "สำนักงานใหญ่", "createdBy": "system"}); n != 1 {
		t.Fatalf("HQ branches = %d, want 1", n)
	}
	if report.HQ.IsZero() {
		t.Fatal("report.HQ is zero")
	}
	for name, want := range legacyCollections {
		if n := countDocs(t, database, name, bson.M{"branchId": report.HQ}); n != int64(want) {
			t.Errorf("%s with HQ branchId = %d, want %d", name, n, want)
		}
	}
}

func TestMigrate_UsesAnExistingHQBranch(t *testing.T) {
	database := newTestDB(t)
	seedLegacy(t, database)
	existing := primitive.NewObjectID()
	if _, err := database.Collection("branches").InsertOne(context.Background(), bson.M{"_id": existing, "code": "HQ", "name": "HQ", "status": "ACTIVE"}); err != nil {
		t.Fatal(err)
	}

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	if n := countDocs(t, database, "branches", bson.M{}); n != 1 {
		t.Errorf("branches count = %d, want 1", n)
	}
	if report.HQCreated {
		t.Error("HQCreated = true, want false")
	}
	if report.HQ != existing {
		t.Errorf("report.HQ = %s, want %s", report.HQ.Hex(), existing.Hex())
	}
}

func TestMigrate_LeavesCurrentDocumentsAlone(t *testing.T) {
	database := newTestDB(t)
	seedLegacy(t, database)
	other := primitive.NewObjectID()
	current := bson.M{"_id": primitive.NewObjectID(), "branchId": other}
	if _, err := database.Collection("orders").InsertOne(context.Background(), current); err != nil {
		t.Fatal(err)
	}

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	if report.HQ == other {
		t.Fatal("test needs a branch other than HQ")
	}
	if n := countDocs(t, database, "orders", bson.M{"_id": current["_id"], "branchId": other}); n != 1 {
		t.Errorf("current order lost its branchId, matches = %d", n)
	}
}

func TestMigrate_SecondRunChangesNothing(t *testing.T) {
	database := newTestDB(t)
	seedLegacy(t, database)
	if _, err := Migrate(context.Background(), database, true); err != nil {
		t.Fatal(err)
	}

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range report.Changes {
		if c.Count != 0 {
			t.Errorf("%s/%s count = %d on second run, want 0", c.Collection, c.Field, c.Count)
		}
	}
	if report.HQCreated {
		t.Error("HQCreated = true on second run")
	}
	if n := countDocs(t, database, "branches", bson.M{}); n != 1 {
		t.Errorf("branches count = %d, want 1", n)
	}
}
