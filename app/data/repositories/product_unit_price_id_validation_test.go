package repositories

import (
	"context"
	"testing"
)

func TestGetProductUnitByIDWithContextRejectsInvalidObjectID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.getProductUnitByIDWithContext(context.Background(), "invalid-id"); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestGetProductUnitsByProductIdRejectsInvalidObjectID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.GetProductUnitsByProductId("invalid-id"); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestGetProductPricesByProductIdRejectsInvalidObjectID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.GetProductPricesByProductId("invalid-id"); err == nil {
		t.Fatal("expected invalid object id error")
	}
}
