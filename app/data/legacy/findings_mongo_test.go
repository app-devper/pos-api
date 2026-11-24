package legacy

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// insert writes a copy of doc with a new _id; doc itself is not changed.
func insert(t *testing.T, db *mongo.Database, coll string, doc bson.M) primitive.ObjectID {
	t.Helper()
	id := primitive.NewObjectID()
	withId := bson.M{"_id": id}
	for k, v := range doc {
		if k != "_id" {
			withId[k] = v
		}
	}
	if _, err := db.Collection(coll).InsertOne(context.Background(), withId); err != nil {
		t.Fatal(err)
	}
	return id
}

func runFindings(t *testing.T, db *mongo.Database) []Finding {
	t.Helper()
	got, err := findings(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func findKind(all []Finding, kind string) (Finding, bool) {
	for _, f := range all {
		if f.Kind == kind {
			return f, true
		}
	}
	return Finding{}, false
}

func wantFinding(t *testing.T, all []Finding, kind string, total int64, ids ...primitive.ObjectID) {
	t.Helper()
	f, ok := findKind(all, kind)
	if !ok {
		t.Fatalf("no %s finding in %+v", kind, all)
	}
	if f.Total != total {
		t.Errorf("%s total = %d, want %d", kind, f.Total, total)
	}
	want := map[string]bool{}
	for _, id := range ids {
		want[id.Hex()] = true
	}
	if len(f.Ids) != len(want) {
		t.Errorf("%s ids = %v, want %d ids", kind, f.Ids, len(want))
	}
	for _, id := range f.Ids {
		if !want[id] {
			t.Errorf("%s unexpected id %s", kind, id)
		}
	}
}

func wantNone(t *testing.T, all []Finding, kind string) {
	t.Helper()
	if f, ok := findKind(all, kind); ok {
		t.Errorf("unexpected %s finding: %+v", kind, f)
	}
}

// seedFindingsProduct inserts a product with its main unit and returns both ids.
func seedFindingsProduct(t *testing.T, db *mongo.Database, serial string) (primitive.ObjectID, primitive.ObjectID) {
	t.Helper()
	pid := insert(t, db, "products", bson.M{"serialNumber": serial, "unit": "ชิ้น"})
	uid := insert(t, db, "product_units", bson.M{"productId": pid, "unit": "ชิ้น"})
	return pid, uid
}

func TestFindings_CleanDatabaseHasNone(t *testing.T) {
	db := newTestDB(t)
	pid, uid := seedFindingsProduct(t, db, "A1")
	insert(t, db, "product_stocks", bson.M{"productId": pid, "unitId": uid, "quantity": 3})
	insert(t, db, "sequences", bson.M{"field": "ORDER"})
	insert(t, db, "payments", bson.M{"type": "CASH"})
	if got := runFindings(t, db); len(got) != 0 {
		t.Fatalf("findings = %+v, want none", got)
	}
}

func TestFindings_NegativeStock(t *testing.T) {
	db := newTestDB(t)
	pid, uid := seedFindingsProduct(t, db, "A1")
	insert(t, db, "product_stocks", bson.M{"productId": pid, "unitId": uid, "quantity": 0})
	bad := insert(t, db, "product_stocks", bson.M{"productId": pid, "unitId": uid, "quantity": -2})
	wantFinding(t, runFindings(t, db), KindNegativeStock, 1, bad)
}

func TestFindings_DuplicateSerial(t *testing.T) {
	db := newTestDB(t)
	a, _ := seedFindingsProduct(t, db, "A1")
	b, _ := seedFindingsProduct(t, db, "A1")
	seedFindingsProduct(t, db, "B2")
	c, _ := seedFindingsProduct(t, db, "")
	d, _ := seedFindingsProduct(t, db, "")
	wantFinding(t, runFindings(t, db), KindDuplicateSerial, 4, a, b, c, d)
}

func TestFindings_DuplicateSerialCountsSoftDeleted(t *testing.T) {
	db := newTestDB(t)
	a, _ := seedFindingsProduct(t, db, "A1")
	b := insert(t, db, "products", bson.M{"serialNumber": "A1", "unit": "ชิ้น", "deletedDate": primitive.NewDateTimeFromTime(time.Now())})
	insert(t, db, "product_units", bson.M{"productId": b, "unit": "ชิ้น"})
	wantFinding(t, runFindings(t, db), KindDuplicateSerial, 2, a, b)
}

func TestFindings_DuplicateSequence(t *testing.T) {
	db := newTestDB(t)
	a := insert(t, db, "sequences", bson.M{"field": "ORDER"})
	b := insert(t, db, "sequences", bson.M{"field": "ORDER"})
	insert(t, db, "sequences", bson.M{"field": "RECEIVE"})
	wantFinding(t, runFindings(t, db), KindDuplicateSequence, 2, a, b)
}

func TestFindings_OrphanStock(t *testing.T) {
	db := newTestDB(t)
	pid, uid := seedFindingsProduct(t, db, "A1")
	insert(t, db, "product_stocks", bson.M{"productId": pid, "unitId": uid, "quantity": 1})
	noProduct := insert(t, db, "product_stocks", bson.M{"productId": primitive.NewObjectID(), "unitId": uid, "quantity": 1})
	noUnit := insert(t, db, "product_stocks", bson.M{"productId": pid, "unitId": primitive.NewObjectID(), "quantity": 1})
	wantFinding(t, runFindings(t, db), KindOrphanStock, 2, noProduct, noUnit)
}

func TestFindings_NoMainUnit(t *testing.T) {
	db := newTestDB(t)
	seedFindingsProduct(t, db, "A1")
	bad := insert(t, db, "products", bson.M{"serialNumber": "B2", "unit": "กล่อง"})
	insert(t, db, "product_units", bson.M{"productId": bad, "unit": "ชิ้น"})
	insert(t, db, "products", bson.M{"serialNumber": "C3", "unit": "กล่อง", "deletedDate": primitive.NewDateTimeFromTime(time.Now())})
	wantFinding(t, runFindings(t, db), KindNoMainUnit, 1, bad)
}

func TestFindings_DiscountConvention(t *testing.T) {
	db := newTestDB(t)
	order := func(total float64, discount float64) primitive.ObjectID {
		id := insert(t, db, "orders", bson.M{"total": total})
		for i := 0; i < 2; i++ {
			insert(t, db, "order_items", bson.M{"orderId": id, "price": 100.0, "quantity": 2, "discount": discount})
		}
		return id
	}
	// two lines of price 100, qty 2, discount 5: per unit 2*(100-10)=180, per line 2*(100-5)=190
	order(180, 5)
	perLine := order(190, 5)
	unknown := order(160, 5)
	order(200, 0) // no discounted line: never examined
	order(1, 0)
	all := runFindings(t, db)
	wantFinding(t, all, KindDiscountPerLine, 1, perLine)
	wantFinding(t, all, KindDiscountUnknown, 1, unknown)
}

func TestFindings_DiscountBothFormulasIsFine(t *testing.T) {
	db := newTestDB(t)
	id := insert(t, db, "orders", bson.M{"total": 95.0})
	insert(t, db, "order_items", bson.M{"orderId": id, "price": 100.0, "quantity": 1, "discount": 5.0})
	if got := runFindings(t, db); len(got) != 0 {
		t.Fatalf("findings = %+v, want none", got)
	}
}

func TestFindings_UnknownPaymentType(t *testing.T) {
	db := newTestDB(t)
	for _, typ := range []string{"CASH", "CREDIT", "PROMPTPAY", "TRANSFER"} {
		insert(t, db, "payments", bson.M{"type": typ})
	}
	bad := insert(t, db, "payments", bson.M{"type": "CHEQUE"})
	wantFinding(t, runFindings(t, db), KindUnknownPaymentType, 1, bad)
}

func TestFindings_IdsCappedAndSorted(t *testing.T) {
	db := newTestDB(t)
	for i := 0; i < 25; i++ {
		insert(t, db, "payments", bson.M{"type": "CHEQUE"})
	}
	f, _ := findKind(runFindings(t, db), KindUnknownPaymentType)
	if f.Total != 25 || len(f.Ids) != maxFindingIds {
		t.Fatalf("total %d ids %d", f.Total, len(f.Ids))
	}
	for i := 1; i < len(f.Ids); i++ {
		if f.Ids[i-1] > f.Ids[i] {
			t.Fatal("ids not sorted")
		}
	}
}

func TestFindings_OrderIsFixed(t *testing.T) {
	db := newTestDB(t)
	insert(t, db, "payments", bson.M{"type": "CHEQUE"})
	insert(t, db, "sequences", bson.M{"field": "X"})
	insert(t, db, "sequences", bson.M{"field": "X"})
	got := runFindings(t, db)
	if len(got) != 2 || got[0].Kind != KindDuplicateSequence || got[1].Kind != KindUnknownPaymentType {
		t.Fatalf("order = %+v", got)
	}
}

func TestMigrate_ReportsFindingsInBothModes(t *testing.T) {
	for _, apply := range []bool{false, true} {
		db := newTestDB(t)
		insert(t, db, "payments", bson.M{"type": "CHEQUE"})
		report, err := Migrate(context.Background(), db, apply)
		if err != nil {
			t.Fatal(err)
		}
		f, ok := findKind(report.Findings, KindUnknownPaymentType)
		if !ok || f.Total != 1 || !report.HasFindings() {
			t.Fatalf("apply=%v findings = %+v", apply, report.Findings)
		}
	}
}

func TestIsReplicaSet(t *testing.T) {
	cases := []struct {
		name  string
		reply bson.M
		want  bool
	}{
		{"replica set", bson.M{"setName": "rs0"}, true},
		{"standalone", bson.M{"ok": 1}, false},
		{"empty name", bson.M{"setName": ""}, false},
		{"wrong type", bson.M{"setName": 3}, false},
		{"mongos", bson.M{"msg": "isdbgrid"}, true},
		{"other msg", bson.M{"msg": "hello"}, false},
	}
	for _, c := range cases {
		if got := isReplicaSet(c.reply); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
}
