package repositories

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// insertRangeOrders stores one Order per status (nil means a legacy Order
// with no status field), coded by its status label, a minute into the hour starting at day.
func insertRangeOrders(f *saleFixture, day time.Time) {
	statuses := map[string]any{
		"confirmed": constant.CONFIRMED,
		"active":    constant.ACTIVE,
		"cancelled": constant.CANCELLED,
		"empty":     "",
		"missing":   nil,
	}
	for code, status := range statuses {
		doc := bson.M{"_id": primitive.NewObjectID(), "branchId": f.branchId, "code": code, "createdDate": day.Add(time.Minute)}
		if status != nil {
			doc["status"] = status
		}
		f.insert("orders", doc)
	}
}

func rangeCodes(t *testing.T, f *saleFixture, form request.GetOrderRange) []string {
	t.Helper()
	orders, err := f.orders.GetOrderRange(form)
	if err != nil {
		t.Fatal(err)
	}
	codes := make([]string, 0, len(orders))
	for _, o := range orders {
		codes = append(codes, o.Code)
	}
	sort.Strings(codes)
	return codes
}

func TestGetOrderRangeExcludeCancelledKeepsStandingAndLegacyOrders(t *testing.T) {
	f := newSaleFixture(t)
	day := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	insertRangeOrders(f, day)
	form := request.GetOrderRange{
		StartDate: request.NewFlexibleTime(day), EndDate: request.NewFlexibleTime(day.Add(time.Hour)),
		BranchId: f.branchId.Hex(),
	}

	all := rangeCodes(t, f, form)
	form.ExcludeCancelled = true
	standing := rangeCodes(t, f, form)

	if want := "[active cancelled confirmed empty missing]"; fmt.Sprint(all) != want {
		t.Fatalf("default range %v, want %s", all, want)
	}
	if want := "[active confirmed empty missing]"; fmt.Sprint(standing) != want {
		t.Fatalf("excludeCancelled range %v, want %s", standing, want)
	}
}
