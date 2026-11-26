package repositories

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newCustomerFixture(t *testing.T) (ICustomer, *mongo.Database) {
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
	pos := client.Database(fmt.Sprintf("pos_customer_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	return NewCustomerEntity(&db.Resource{Client: client, PosDb: pos}), pos
}

// insertStatusCustomers stores one Customer per status, coded by its status
// label; nil means a legacy Customer with no status field.
func insertStatusCustomers(t *testing.T, pos *mongo.Database) {
	t.Helper()
	statuses := map[string]any{
		"active":   constant.ACTIVE,
		"inactive": constant.INACTIVE,
		"draft":    "DRAFT",
		"archived": constant.ARCHIVED,
		"empty":    "",
		"missing":  nil,
	}
	for code, status := range statuses {
		doc := bson.M{"_id": primitive.NewObjectID(), "code": code, "name": code}
		if status != nil {
			doc["status"] = status
		}
		if _, err := pos.Collection("customers").InsertOne(context.Background(), doc); err != nil {
			t.Fatal(err)
		}
	}
}

func customerCodes(t *testing.T, customers ICustomer, form request.GetCustomers) string {
	t.Helper()
	items, err := customers.GetCustomerAll(form)
	if err != nil {
		t.Fatal(err)
	}
	codes := make([]string, 0, len(items))
	for _, c := range items {
		codes = append(codes, c.Code)
	}
	sort.Strings(codes)
	return fmt.Sprint(codes)
}

func TestGetCustomerAllExcludeArchivedKeepsEveryOtherCustomer(t *testing.T) {
	customers, pos := newCustomerFixture(t)
	insertStatusCustomers(t, pos)

	all := customerCodes(t, customers, request.GetCustomers{})
	kept := customerCodes(t, customers, request.GetCustomers{ExcludeArchived: true})

	if want := "[active archived draft empty inactive missing]"; all != want {
		t.Fatalf("default list %s, want %s", all, want)
	}
	if want := "[active draft empty inactive missing]"; kept != want {
		t.Fatalf("excludeArchived list %s, want %s", kept, want)
	}
}

func TestGetCustomerByCodeStillResolvesAnArchivedCustomer(t *testing.T) {
	customers, pos := newCustomerFixture(t)
	insertStatusCustomers(t, pos)

	found, err := customers.GetCustomerByCode("archived")

	if err != nil || found.Status != constant.ARCHIVED {
		t.Fatalf("archived customer by code: %+v, %v", found, err)
	}
}
