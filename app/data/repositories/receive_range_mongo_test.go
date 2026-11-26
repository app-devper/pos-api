package repositories

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// insertRangeReceives stores one Receive per status in branchId, coded by its
// status label, a minute after day; nil means a legacy Receive with no status.
func insertRangeReceives(t *testing.T, pos *mongo.Database, branchId primitive.ObjectID, day time.Time) {
	t.Helper()
	statuses := map[string]any{
		"draft":     constant.ACTIVE,
		"imported":  constant.IMPORTED,
		"cancelled": constant.CANCELLED,
		"empty":     "",
		"missing":   nil,
	}
	for code, status := range statuses {
		doc := bson.M{"_id": primitive.NewObjectID(), "branchId": branchId, "code": code, "createdDate": day.Add(time.Minute)}
		if status != nil {
			doc["status"] = status
		}
		if _, err := pos.Collection("receives").InsertOne(context.Background(), doc); err != nil {
			t.Fatal(err)
		}
	}
}

func receiveRangeCodes(t *testing.T, e *receiveEntity, form request.GetReceiveRange) string {
	t.Helper()
	receives, err := e.GetReceives(form)
	if err != nil {
		t.Fatal(err)
	}
	codes := make([]string, 0, len(receives))
	for _, r := range receives {
		codes = append(codes, r.Code)
	}
	sort.Strings(codes)
	return fmt.Sprint(codes)
}

func TestGetReceivesExcludeCancelledKeepsDraftImportedAndLegacyReceipts(t *testing.T) {
	e, pos := newReceiveDocFixture(t)
	branchId := primitive.NewObjectID()
	day := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	insertRangeReceives(t, pos, branchId, day)
	form := request.GetReceiveRange{
		StartDate: request.NewFlexibleTime(day), EndDate: request.NewFlexibleTime(day.Add(time.Hour)),
		BranchId: branchId.Hex(),
	}

	all := receiveRangeCodes(t, e, form)
	form.ExcludeCancelled = true
	standing := receiveRangeCodes(t, e, form)

	if want := "[cancelled draft empty imported missing]"; all != want {
		t.Fatalf("default range %s, want %s", all, want)
	}
	if want := "[draft empty imported missing]"; standing != want {
		t.Fatalf("excludeCancelled range %s, want %s", standing, want)
	}
}
