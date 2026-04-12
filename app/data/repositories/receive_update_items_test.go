package repositories

import (
	"testing"
	"time"

	"pos/app/domain/request"
)

func TestBuildReceiveItemsMapsPayloadForReceiveItemsCollection(t *testing.T) {
	items, err := buildReceiveItems([]request.ReceiveItem{
		{
			ProductId:    "507f1f77bcf86cd799439011",
			CostPrice:    12.5,
			Quantity:     3,
			LotNumber:    "LOT-01",
			UnitId:       "BOX",
			BaseQuantity: 30,
			ExpireDate:   "2026-12-31T00:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("expected buildReceiveItems to succeed, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ProductId.Hex() != "507f1f77bcf86cd799439011" {
		t.Fatalf("expected product id to be converted, got %s", items[0].ProductId.Hex())
	}
	if items[0].LotNumber != "LOT-01" || items[0].UnitId != "BOX" {
		t.Fatalf("expected lot/unit to be preserved, got %+v", items[0])
	}
	if items[0].ExpireDate.Format(time.RFC3339) != "2026-12-31T00:00:00Z" {
		t.Fatalf("expected expire date to be parsed, got %s", items[0].ExpireDate.Format(time.RFC3339))
	}
}

func TestBuildReceiveItemsReturnsErrorForInvalidProductId(t *testing.T) {
	_, err := buildReceiveItems([]request.ReceiveItem{
		{
			ProductId: "invalid",
		},
	})
	if err == nil {
		t.Fatal("expected invalid product id to return an error")
	}
}
