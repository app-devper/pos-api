package repositories

import (
	"testing"

	"pos/app/domain/constant"
)

func TestOrderCancellationOnlyAllowsConfirmedStatuses(t *testing.T) {
	if !constant.IsConfirmedOrderStatus(constant.ACTIVE) {
		t.Fatal("expected ACTIVE legacy orders to be cancellable")
	}
	if !constant.IsConfirmedOrderStatus(constant.CONFIRMED) {
		t.Fatal("expected CONFIRMED orders to be cancellable")
	}
	if constant.IsConfirmedOrderStatus(constant.CANCELLED) {
		t.Fatal("expected CANCELLED orders to be excluded from cancellation")
	}
	if constant.IsConfirmedOrderStatus(constant.VOIDED) {
		t.Fatal("expected VOIDED orders to be excluded from cancellation")
	}
}
