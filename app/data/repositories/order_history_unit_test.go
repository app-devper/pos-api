package repositories

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestOrderHistoryUnitLookupErrorMapsMissingUnitToBusinessError(t *testing.T) {
	err := orderHistoryUnitLookupError(mongo.ErrNoDocuments)
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "product unit not found for order history" {
		t.Fatalf("expected missing-unit business error, got %q", err.Error())
	}
}

func TestOrderHistoryUnitLookupErrorPreservesOtherErrors(t *testing.T) {
	original := errors.New("db timeout")
	err := orderHistoryUnitLookupError(original)
	if !errors.Is(err, original) {
		t.Fatalf("expected original error to be preserved, got %v", err)
	}
}
