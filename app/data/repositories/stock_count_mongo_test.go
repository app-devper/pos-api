package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newStockCountFixture(t *testing.T) (IStockCount, *mongo.Database) {
	t.Helper()
	repo, pos, _ := newCountedStockCountFixture(t)
	return repo, pos
}

// commandCounter counts the commands a client starts, by name.
type commandCounter struct {
	mu    sync.Mutex
	names map[string]int
}

func (c *commandCounter) started(_ context.Context, e *event.CommandStartedEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.names[e.CommandName]++
}

// reads is how many find, aggregate and getMore commands have started since
// the last reset, and resets the count.
func (c *commandCounter) reads() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	got := map[string]int{"find": c.names["find"], "aggregate": c.names["aggregate"], "getMore": c.names["getMore"]}
	c.names = map[string]int{}
	return got
}

func newCountedStockCountFixture(t *testing.T) (IStockCount, *mongo.Database, *commandCounter) {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}
	ctx := context.Background()
	counter := &commandCounter{names: map[string]int{}}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMonitor(&event.CommandMonitor{Started: counter.started}))
	if err != nil {
		t.Fatal(err)
	}
	pos := client.Database(fmt.Sprintf("pos_stock_count_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return NewStockCountEntity(&db.Resource{Client: client, PosDb: pos}), pos, counter
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

// countOf inserts a Count in branch, made at, of one Line per stock (each
// stock of product), and returns its id.
func countOf(t *testing.T, pos *mongo.Database, branch primitive.ObjectID, no string, at time.Time, product primitive.ObjectID, stocks ...primitive.ObjectID) primitive.ObjectID {
	t.Helper()
	lines := make([]entities.StockCountItem, len(stocks))
	for i, st := range stocks {
		lines[i] = entities.StockCountItem{ProductId: product, StockId: st, SystemQuantity: i, CountedQuantity: i}
	}
	id := primitive.NewObjectID()
	insertDoc(t, pos, "stock_counts", entities.StockCount{Id: id, BranchId: branch, CountNo: no, CreatedDate: at, Items: lines})
	return id
}

// namedStock inserts a Product with one Unit and one Stock in branch.
func namedStock(t *testing.T, pos *mongo.Database, branch primitive.ObjectID, name, unit, lot string) (primitive.ObjectID, primitive.ObjectID) {
	t.Helper()
	product, unitId, stock := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	insertDoc(t, pos, "products", entities.Product{Id: product, Name: name})
	insertDoc(t, pos, "product_units", entities.ProductUnit{Id: unitId, ProductId: product, Unit: unit})
	insertDoc(t, pos, "product_stocks", entities.ProductStock{Id: stock, BranchId: branch, ProductId: product, UnitId: unitId, LotNumber: lot})
	return product, stock
}

func TestStockCountListIsTheCallersBranchNewestFirstEachLineNamed(t *testing.T) {
	repo, pos := newStockCountFixture(t)
	mine, other := primitive.NewObjectID(), primitive.NewObjectID()
	p1, s1 := namedStock(t, pos, mine, "Paracetamol", "TAB", "L-1")
	p2, s2 := namedStock(t, pos, mine, "Syrup", "BOT", "L-2")
	now := time.Now()
	countOf(t, pos, mine, "SC-1", now.Add(-2*time.Hour), p1, s1)
	countOf(t, pos, other, "SC-2", now.Add(-time.Hour), p1, s1)
	countOf(t, pos, mine, "SC-3", now, p2, s2)

	got, err := repo.GetStockCounts(mine.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].CountNo != "SC-3" || got[1].CountNo != "SC-1" {
		t.Fatalf("got %+v, want SC-3 then SC-1", got)
	}
	if fmt.Sprint(lineNames(got[0].Items)) != fmt.Sprint([]countLineNames{{"Syrup", "BOT", "L-2"}}) ||
		fmt.Sprint(lineNames(got[1].Items)) != fmt.Sprint([]countLineNames{{"Paracetamol", "TAB", "L-1"}}) {
		t.Fatalf("lines named %v and %v", lineNames(got[0].Items), lineNames(got[1].Items))
	}
}

func TestStockCountNamesEveryLineOfAStockListedTwice(t *testing.T) {
	repo, pos := newStockCountFixture(t)
	branch := primitive.NewObjectID()
	p1, s1 := namedStock(t, pos, branch, "Paracetamol", "TAB", "L-1")
	_, s2 := namedStock(t, pos, branch, "Syrup", "BOT", "L-2")
	id := countOf(t, pos, branch, "SC-1", time.Now(), p1, s1, s2, s1)

	got, err := repo.GetStockCountById(id.Hex(), branch.Hex())
	if err != nil {
		t.Fatal(err)
	}
	want := []countLineNames{{"Paracetamol", "TAB", "L-1"}, {"Paracetamol", "BOT", "L-2"}, {"Paracetamol", "TAB", "L-1"}}
	if fmt.Sprint(lineNames(got.Items)) != fmt.Sprint(want) || got.Items[1].StockId != s2 || got.Items[2].SystemQuantity != 2 {
		t.Fatalf("lines %+v, want named %v in stored order", got.Items, want)
	}
}

// Naming costs one query per collection for the whole response, however
// many Counts it holds: never one lookup per Count.
func TestStockCountReadsCostTheSameQueriesForOneCountOrMany(t *testing.T) {
	repo, pos, counter := newCountedStockCountFixture(t)
	branch := primitive.NewObjectID()
	p1, s1 := namedStock(t, pos, branch, "Paracetamol", "TAB", "L-1")
	byCount := map[int]map[string]int{}
	for _, n := range []int{1, 8} {
		for len(mustList(t, repo, branch)) < n {
			countOf(t, pos, branch, "SC", time.Now(), p1, s1)
		}
		counter.reads()
		if got := mustList(t, repo, branch); len(got) != n || got[0].Items[0].ProductName != "Paracetamol" {
			t.Fatalf("listed %+v, want %d named Counts", got, n)
		}
		byCount[n] = counter.reads()
	}
	want := map[string]int{"find": 4, "aggregate": 0, "getMore": 0}
	for n, got := range byCount {
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("listing %d Counts issued %v, want %v (stock_counts, products, product_stocks, product_units)", n, got, want)
		}
	}
	id := countOf(t, pos, branch, "SC-X", time.Now(), p1, s1)
	counter.reads()
	if _, err := repo.GetStockCountById(id.Hex(), branch.Hex()); err != nil {
		t.Fatal(err)
	}
	if got := counter.reads(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("one Count by id issued %v, want %v", got, want)
	}
}

func mustList(t *testing.T, repo IStockCount, branch primitive.ObjectID) []entities.StockCountDetail {
	t.Helper()
	got, err := repo.GetStockCounts(branch.Hex())
	if err != nil {
		t.Fatal(err)
	}
	return got
}
