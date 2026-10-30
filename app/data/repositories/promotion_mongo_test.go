package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/promotion"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newPromotionFixture(t *testing.T) (*promotionEntity, *mongo.Database, primitive.ObjectID, primitive.ObjectID) {
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
	pos := client.Database(fmt.Sprintf("pos_promotion_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() { _ = pos.Drop(ctx); _ = client.Disconnect(ctx) })
	id, branch := primitive.NewObjectID(), primitive.NewObjectID()
	start, end := time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour)
	if _, err := pos.Collection("promotions").InsertOne(ctx, entities.Promotion{Id: id, BranchId: branch, Code: "P10",
		Name: "Ten off", Type: promotion.Percentage, Value: 10, StartDate: start, EndDate: end, Status: "ACTIVE"}); err != nil {
		t.Fatal(err)
	}
	return NewPromotionEntity(&db.Resource{Client: client, PosDb: pos}).(*promotionEntity), pos, id, branch
}

func TestAPartialUpdateChangesOnlyWhatWasSent(t *testing.T) {
	e, _, id, branch := newPromotionFixture(t)
	status := "INACTIVE"

	got, err := e.UpdatePromotionById(id.Hex(), branch.Hex(), request.UpdatePromotion{Status: status})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != status || got.Name != "Ten off" || got.Value != 10 || got.Type != promotion.Percentage || got.EndDate.IsZero() {
		t.Fatalf("a status-only update changed other fields: %+v", got)
	}
}

func TestAnUpdateThatBreaksTheTermsIsRefused(t *testing.T) {
	e, pos, id, branch := newPromotionFixture(t)
	value := 150.0

	_, err := e.UpdatePromotionById(id.Hex(), branch.Hex(), request.UpdatePromotion{Value: &value})

	if !errors.Is(err, promotion.ErrValue) {
		t.Fatalf("got %v, want promotion.ErrValue", err)
	}
	var stored entities.Promotion
	_ = pos.Collection("promotions").FindOne(context.Background(), bson.M{"_id": id}).Decode(&stored)
	if stored.Value != 10 {
		t.Fatalf("value changed to %v", stored.Value)
	}
}

func TestAPromotionIsCreatedOnlyWithValidTerms(t *testing.T) {
	e, _, _, branch := newPromotionFixture(t)
	form := request.Promotion{Code: "BAD", Name: "Bad", Type: "BOGO", Value: 1, BranchId: branch.Hex(),
		StartDate: request.NewFlexibleTime(time.Now()), EndDate: request.NewFlexibleTime(time.Now().Add(time.Hour))}

	if _, err := e.CreatePromotion(form); !errors.Is(err, promotion.ErrUnknownType) {
		t.Fatalf("got %v, want ErrUnknownType", err)
	}
	form.Type, form.Value = "percentage", 5
	created, err := e.CreatePromotion(form)
	if err != nil {
		t.Fatal(err)
	}
	if created.Type != promotion.Percentage {
		t.Fatalf("type stored as %q, want %q", created.Type, promotion.Percentage)
	}
}
