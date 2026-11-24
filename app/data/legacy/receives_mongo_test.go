package legacy

import (
	"context"
	"reflect"
	"testing"
	"time"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func insertDoc(t *testing.T, database *mongo.Database, collection string, doc bson.M) {
	t.Helper()
	if _, err := database.Collection(collection).InsertOne(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
}

func findReceive(t *testing.T, database *mongo.Database, id primitive.ObjectID) bson.M {
	t.Helper()
	var got bson.M
	if err := database.Collection("receives").FindOne(context.Background(), bson.M{"_id": id}).Decode(&got); err != nil {
		t.Fatal(err)
	}
	return got
}

func changeFor(report Report, collection, field string) int64 {
	for _, c := range report.Changes {
		if c.Collection == collection && c.Field == field {
			return c.Count
		}
	}
	return -1
}

func TestMigrate_LegacyReceiveBecomesImportedWithItsItems(t *testing.T) {
	database := newTestDB(t)
	receiveID := primitive.NewObjectID()
	productA, productB := primitive.NewObjectID(), primitive.NewObjectID()
	now := time.Now()
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "code": "RC_0001", "totalCost": 100.0})
	// Inserted newest first: the items must come back in createdDate order.
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": productB, "costPrice": 5.0, "quantity": 20, "createdBy": "u", "createdDate": now.Add(time.Minute)})
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": productA, "costPrice": 4.0, "quantity": 10, "createdBy": "u", "createdDate": now})

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	got := findReceive(t, database, receiveID)
	if got["branchId"] != report.HQ {
		t.Errorf("branchId = %v, want HQ %s", got["branchId"], report.HQ.Hex())
	}
	if got["status"] != constant.IMPORTED {
		t.Errorf("status = %v, want %s", got["status"], constant.IMPORTED)
	}
	want := bson.A{
		bson.M{"productId": productA, "costPrice": 4.0, "quantity": int32(10)},
		bson.M{"productId": productB, "costPrice": 5.0, "quantity": int32(20)},
	}
	items, _ := got["items"].(bson.A)
	if len(items) != len(want) {
		t.Fatalf("items = %v, want %v", got["items"], want)
	}
	for i := range want {
		item, _ := items[i].(bson.M)
		w := want[i].(bson.M)
		if len(item) != 3 || item["productId"] != w["productId"] || item["costPrice"] != w["costPrice"] || item["quantity"] != w["quantity"] {
			t.Errorf("items[%d] = %v, want %v", i, item, w)
		}
	}
	for _, field := range []string{"branchId", "status", "items"} {
		if n := changeFor(report, "receives", field); n != 1 {
			t.Errorf("receives/%s change = %d, want 1", field, n)
		}
	}
	if n := countDocs(t, database, "receive_items", bson.M{}); n != 2 {
		t.Errorf("receive_items = %d, want 2", n)
	}
}

func TestMigrate_CurrentDraftReceiveStaysADraft(t *testing.T) {
	database := newTestDB(t)
	receiveID, branch := primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "branchId": branch, "code": "RC_0002", "status": "",
		"items": bson.A{bson.M{"productId": primitive.NewObjectID(), "costPrice": 1.0, "quantity": 3}}})
	before := findReceive(t, database, receiveID)

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	after := findReceive(t, database, receiveID)
	if len(after) != len(before) || after["status"] != "" || after["branchId"] != branch {
		t.Errorf("draft changed: before %v, after %v", before, after)
	}
	if items, _ := after["items"].(bson.A); len(items) != 1 {
		t.Errorf("items = %v, want the one it had", after["items"])
	}
	for _, field := range []string{"branchId", "status", "items"} {
		if n := changeFor(report, "receives", field); n != 0 {
			t.Errorf("receives/%s change = %d, want 0", field, n)
		}
	}
}

func TestMigrate_LegacyReceiveWithItsOwnItemsKeepsThem(t *testing.T) {
	database := newTestDB(t)
	receiveID, product := primitive.NewObjectID(), primitive.NewObjectID()
	own := bson.A{bson.M{"productId": product, "costPrice": 7.0, "quantity": int32(4)}}
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "code": "RC_0003", "items": own})
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": primitive.NewObjectID(), "costPrice": 1.0, "quantity": 99, "createdDate": time.Now()})

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	got := findReceive(t, database, receiveID)
	items, _ := got["items"].(bson.A)
	if len(items) != 1 {
		t.Fatalf("items = %v, want its own single item", got["items"])
	}
	item, _ := items[0].(bson.M)
	if item["productId"] != product || item["costPrice"] != 7.0 || item["quantity"] != int32(4) {
		t.Errorf("item = %v, want %v", item, own[0])
	}
	if got["branchId"] != report.HQ || got["status"] != constant.IMPORTED {
		t.Errorf("receive not migrated: %v", got)
	}
	if n := changeFor(report, "receives", "items"); n != 0 {
		t.Errorf("receives/items change = %d, want 0", n)
	}
}

func TestMigrate_ReportModeLeavesAReceiveAsSeeded(t *testing.T) {
	database := newTestDB(t)
	receiveID := primitive.NewObjectID()
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "code": "RC_0004", "totalCost": 10.0})
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": primitive.NewObjectID(), "costPrice": 1.0, "quantity": 1, "createdDate": time.Now()})
	before := findReceive(t, database, receiveID)

	report, err := Migrate(context.Background(), database, false)
	if err != nil {
		t.Fatal(err)
	}

	after := findReceive(t, database, receiveID)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("report mode changed the receive: before %v, after %v", before, after)
	}
	for _, field := range []string{"branchId", "status", "items"} {
		if n := changeFor(report, "receives", field); n != 1 {
			t.Errorf("receives/%s change = %d, want 1", field, n)
		}
	}
}

func TestMigrate_SecondRunChangesNoReceives(t *testing.T) {
	database := newTestDB(t)
	receiveID := primitive.NewObjectID()
	insertDoc(t, database, "receives", bson.M{"_id": receiveID, "code": "RC_0005", "totalCost": 10.0})
	insertDoc(t, database, "receive_items", bson.M{"_id": primitive.NewObjectID(), "receiveId": receiveID, "productId": primitive.NewObjectID(), "costPrice": 1.0, "quantity": 1, "createdDate": time.Now()})
	if _, err := Migrate(context.Background(), database, true); err != nil {
		t.Fatal(err)
	}
	between := findReceive(t, database, receiveID)

	report, err := Migrate(context.Background(), database, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, field := range []string{"branchId", "status", "items"} {
		if n := changeFor(report, "receives", field); n != 0 {
			t.Errorf("receives/%s change = %d on second run, want 0", field, n)
		}
	}
	if after := findReceive(t, database, receiveID); !reflect.DeepEqual(between, after) {
		t.Errorf("second run changed the receive: %v -> %v", between, after)
	}
}
