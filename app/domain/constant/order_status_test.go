package constant

import "testing"

func TestIsConfirmedOrderItemStatusSupportsLegacyAndConfirmedItems(t *testing.T) {
	if !IsConfirmedOrderItemStatus("") {
		t.Fatal("expected empty item status to be treated as legacy confirmed")
	}
	if !IsConfirmedOrderItemStatus(ACTIVE) {
		t.Fatal("expected ACTIVE item status to be treated as confirmed")
	}
	if !IsConfirmedOrderItemStatus(CONFIRMED) {
		t.Fatal("expected CONFIRMED item status to be treated as confirmed")
	}
	if IsConfirmedOrderItemStatus(CANCELLED) {
		t.Fatal("expected CANCELLED item status to be excluded")
	}
	if IsConfirmedOrderItemStatus(VOIDED) {
		t.Fatal("expected VOIDED item status to be excluded")
	}
}
