package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A Receive is a draft until it is imported or cancelled. Every change to it
// checks that in the same write, so nothing slips in after an import commits.
func newReceiveDocFixture(t *testing.T) (*receiveEntity, *mongo.Database) {
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
	pos := client.Database(fmt.Sprintf("pos_receive_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return NewReceiveEntity(&db.Resource{Client: client, PosDb: pos}).(*receiveEntity), pos
}

func insertReceive(t *testing.T, pos *mongo.Database, status string) primitive.ObjectID {
	t.Helper()
	id := primitive.NewObjectID()
	doc := bson.M{"_id": id, "branchId": primitive.NewObjectID(), "code": "RC1", "status": status, "totalCost": 5.0}
	if _, err := pos.Collection("receives").InsertOne(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
	if _, err := pos.Collection("receive_items").InsertOne(context.Background(),
		bson.M{"receiveId": id, "productId": primitive.NewObjectID(), "quantity": 1, "costPrice": 5.0}); err != nil {
		t.Fatal(err)
	}
	return id
}

func receiveNow(t *testing.T, pos *mongo.Database, id primitive.ObjectID) (entities.Receive, int64) {
	t.Helper()
	var r entities.Receive
	if err := pos.Collection("receives").FindOne(context.Background(), bson.M{"_id": id}).Decode(&r); err != nil {
		t.Fatal(err)
	}
	n, err := pos.Collection("receive_items").CountDocuments(context.Background(), bson.M{"receiveId": id})
	if err != nil {
		t.Fatal(err)
	}
	return r, n
}

func TestAReceiveThatIsNoLongerADraftRefusesEveryChange(t *testing.T) {
	item := request.ReceiveItem{ProductId: primitive.NewObjectID().Hex(), Quantity: 3, CostPrice: 2}
	changes := map[string]func(e *receiveEntity, id string) error{
		"edit": func(e *receiveEntity, id string) error {
			_, err := e.UpdateReceiveById(id, request.UpdateReceive{SupplierId: primitive.NewObjectID().Hex(), ReceiveItems: []request.ReceiveItem{item}})
			return err
		},
		"edit items": func(e *receiveEntity, id string) error {
			_, err := e.UpdateReceiveItemsById(id, request.UpdateReceiveItems{ReceiveItems: []request.ReceiveItem{item}})
			return err
		},
		"set total cost": func(e *receiveEntity, id string) error {
			_, err := e.UpdateReceiveTotalCostById(id, 99)
			return err
		},
		"cancel": func(e *receiveEntity, id string) error {
			_, err := e.CancelReceiveById(id, "u1")
			return err
		},
	}
	for _, status := range []string{constant.IMPORTED, constant.CANCELLED} {
		for name, change := range changes {
			t.Run(status+"/"+name, func(t *testing.T) {
				e, pos := newReceiveDocFixture(t)
				id := insertReceive(t, pos, status)

				err := change(e, id.Hex())

				if !errors.Is(err, ErrReceiveLocked) {
					t.Fatalf("got %v, want ErrReceiveLocked", err)
				}
				r, items := receiveNow(t, pos, id)
				if r.Status != status || r.TotalCost != 5 || items != 1 {
					t.Fatalf("receive changed: status %q totalCost %v items %d", r.Status, r.TotalCost, items)
				}
			})
		}
	}
}

func TestADraftReceiveCanBeEditedAndCancelled(t *testing.T) {
	e, pos := newReceiveDocFixture(t)
	id := insertReceive(t, pos, "")
	item := request.ReceiveItem{ProductId: primitive.NewObjectID().Hex(), Quantity: 3, CostPrice: 2}

	if _, err := e.UpdateReceiveItemsById(id.Hex(), request.UpdateReceiveItems{ReceiveItems: []request.ReceiveItem{item, item}}); err != nil {
		t.Fatal(err)
	}
	r, items := receiveNow(t, pos, id)
	if items != 2 || r.TotalCost != 12 {
		t.Fatalf("items %d totalCost %v, want 2 and 12", items, r.TotalCost)
	}

	if _, err := e.UpdateReceiveById(id.Hex(), request.UpdateReceive{SupplierId: primitive.NewObjectID().Hex(),
		ReceiveItems: []request.ReceiveItem{item}}); err != nil {
		t.Fatal(err)
	}
	if r, items := receiveNow(t, pos, id); items != 1 || r.TotalCost != 6 {
		t.Fatalf("items %d totalCost %v, want 1 and 6", items, r.TotalCost)
	}

	if _, err := e.CancelReceiveById(id.Hex(), "u1"); err != nil {
		t.Fatal(err)
	}
	if r, _ := receiveNow(t, pos, id); r.Status != constant.CANCELLED {
		t.Fatalf("status %q, want CANCELLED", r.Status)
	}
}

func TestAMissingReceiveIsNotFoundRatherThanLocked(t *testing.T) {
	e, _ := newReceiveDocFixture(t)

	_, err := e.CancelReceiveById(primitive.NewObjectID().Hex(), "u1")

	if err == nil || errors.Is(err, ErrReceiveLocked) {
		t.Fatalf("got %v, want a not-found error", err)
	}
}
