package repositories

import (
	"testing"
	"time"

	"pos/app/domain/request"
)

func TestCreateOrderRejectsInvalidBranchId(t *testing.T) {
	entity := &orderEntity{}

	if _, err := entity.createOrderWithContext(nil, request.Order{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetOrderRangeRejectsInvalidBranchId(t *testing.T) {
	entity := &orderEntity{}

	if _, err := entity.GetOrderRange(request.GetOrderRange{
		BranchId:  "invalid-branch-id",
		StartDate: time.Now().Add(-time.Hour),
		EndDate:   time.Now(),
	}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetOrderItemRangeRejectsInvalidBranchId(t *testing.T) {
	entity := &orderEntity{}

	if _, err := entity.GetOrderItemRange(request.GetOrderRange{
		BranchId:  "invalid-branch-id",
		StartDate: time.Now().Add(-time.Hour),
		EndDate:   time.Now(),
	}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateDispensingLogRejectsInvalidBranchId(t *testing.T) {
	entity := &dispensingLogEntity{}

	if _, err := entity.CreateDispensingLog(request.DispensingLog{
		BranchId:  "invalid-branch-id",
		OrderId:   "507f1f77bcf86cd799439011",
		PatientId: "507f1f77bcf86cd799439012",
	}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateDispensingLogRejectsInvalidProductId(t *testing.T) {
	entity := &dispensingLogEntity{}

	if _, err := entity.CreateDispensingLog(request.DispensingLog{
		BranchId:  "507f1f77bcf86cd799439011",
		OrderId:   "507f1f77bcf86cd799439012",
		PatientId: "507f1f77bcf86cd799439013",
		Items: []request.DispensingItem{
			{ProductId: "invalid-product-id"},
		},
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
}

func TestGetDispensingLogsRejectsInvalidBranchId(t *testing.T) {
	entity := &dispensingLogEntity{}

	if _, err := entity.GetDispensingLogs("invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetRefillRemindersRejectsInvalidBranchId(t *testing.T) {
	entity := &dispensingLogEntity{}

	if _, err := entity.GetRefillReminders("invalid-branch-id", 30); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetDispensingLogsByDateRangeRejectsInvalidBranchId(t *testing.T) {
	entity := &dispensingLogEntity{}

	if _, err := entity.GetDispensingLogsByDateRange("invalid-branch-id", time.Now().Add(-time.Hour), time.Now()); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}
