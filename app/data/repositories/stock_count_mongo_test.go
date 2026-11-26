package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newStockCountFixture(t *testing.T) (IStockCount, *mongo.Database) {
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
	pos := client.Database(fmt.Sprintf("pos_stock_count_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return NewStockCountEntity(&db.Resource{Client: client, PosDb: pos}), pos
}

func insertDoc(t *testing.T, pos *mongo.Database, collection string, doc any) {
	t.Helper()
	if _, err := pos.Collection(collection).InsertOne(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
}

type countLineNames struct{ productName, unit, lotNumber string }

func lineNames(lines []entities.StockCountLine) []countLineNames {
	got := make([]countLineNames, len(lines))
	for i, l := range lines {
		got[i] = countLineNames{l.ProductName, l.Unit, l.LotNumber}
	}
	return got
}

// A Count stores only ids; its Lines are named when read, so a Count made
// before names were returned is named too, and a Line whose Product or Stock
// has gone since is still returned, with what is unknown left empty.
func TestStockCountLinesAreNamedWhenRead(t *testing.T) {
	repo, pos := newStockCountFixture(t)
	branch := primitive.NewObjectID()
	live, retired, gone := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	tab, box := primitive.NewObjectID(), primitive.NewObjectID()
	liveStock, retiredStock, goneStock := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, pos, "products", entities.Product{Id: live, Name: "Paracetamol"})
	deleted := time.Now()
	insertDoc(t, pos, "products", entities.Product{Id: retired, Name: "Old syrup", DeletedDate: &deleted})
	insertDoc(t, pos, "product_units", entities.ProductUnit{Id: tab, ProductId: live, Unit: "TAB"})
	insertDoc(t, pos, "product_units", entities.ProductUnit{Id: box, ProductId: retired, Unit: "BOX"})
	insertDoc(t, pos, "product_stocks", entities.ProductStock{Id: liveStock, BranchId: branch, ProductId: live, UnitId: tab, LotNumber: "L-1"})
	insertDoc(t, pos, "product_stocks", entities.ProductStock{Id: retiredStock, BranchId: branch, ProductId: retired, UnitId: box, LotNumber: "L-2"})
	count := entities.StockCount{Id: primitive.NewObjectID(), BranchId: branch, CountNo: "SC-1", CreatedDate: time.Now(), Items: []entities.StockCountItem{
		{ProductId: live, StockId: liveStock, SystemQuantity: 5, CountedQuantity: 3, Delta: -2},
		{ProductId: retired, StockId: retiredStock, SystemQuantity: 1, CountedQuantity: 1},
		{ProductId: gone, StockId: goneStock, SystemQuantity: 2, CountedQuantity: 0, Delta: -2},
	}}
	insertDoc(t, pos, "stock_counts", count)
	want := []countLineNames{{"Paracetamol", "TAB", "L-1"}, {"Old syrup", "BOX", "L-2"}, {"", "", ""}}

	one, err := repo.GetStockCountById(count.Id.Hex(), branch.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got := lineNames(one.Items); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("by id: lines named %v, want %v", got, want)
	}
	if one.CountNo != "SC-1" || one.Items[0].ProductId != live || one.Items[0].StockId != liveStock || one.Items[0].Delta != -2 || one.Items[2].SystemQuantity != 2 {
		t.Fatalf("by id: stored fields changed: %+v", one)
	}
	list, err := repo.GetStockCounts(branch.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || fmt.Sprint(lineNames(list[0].Items)) != fmt.Sprint(want) {
		t.Fatalf("list: %+v, want one Count with lines named %v", list, want)
	}
	body, err := json.Marshal(one.Items[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"productId"`, `"stockId"`, `"systemQuantity"`, `"countedQuantity"`, `"delta"`, `"productName":"Paracetamol"`, `"unit":"TAB"`, `"lotNumber":"L-1"`} {
		if !strings.Contains(string(body), key) {
			t.Fatalf("line JSON %s lacks %s", body, key)
		}
	}
}

// A Count with no Lines (never written by the ledger, but possible in old
// data) is still returned, with an empty list of Lines.
func TestStockCountWithoutLinesReadsAsEmpty(t *testing.T) {
	repo, pos := newStockCountFixture(t)
	branch := primitive.NewObjectID()
	id := primitive.NewObjectID()
	insertDoc(t, pos, "stock_counts", map[string]any{"_id": id, "branchId": branch, "countNo": "SC-0", "createdDate": time.Now()})

	got, err := repo.GetStockCountById(id.Hex(), branch.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got.CountNo != "SC-0" || len(got.Items) != 0 {
		t.Fatalf("got %+v, want SC-0 with no Lines", got)
	}
}
