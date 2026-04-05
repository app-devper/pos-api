package repositories

import (
	"testing"

	"pos/app/domain/request"
)

func TestGetReceivesRejectsInvalidBranchId(t *testing.T) {
	entity := &receiveEntity{}

	if _, err := entity.GetReceives(request.GetReceiveRange{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateReceiveRejectsInvalidBranchId(t *testing.T) {
	entity := &receiveEntity{}

	if _, err := entity.CreateReceive(request.Receive{SupplierId: "507f1f77bcf86cd799439011", BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateReceiveItemRejectsInvalidReceiveId(t *testing.T) {
	entity := &receiveEntity{}

	if _, err := entity.CreateReceiveItem("invalid-receive-id", "", "507f1f77bcf86cd799439011", request.Product{}); err == nil {
		t.Fatal("expected invalid receive id error")
	}
}

func TestCreatePromotionRejectsInvalidBranchId(t *testing.T) {
	entity := &promotionEntity{}

	if _, err := entity.CreatePromotion(request.Promotion{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreatePromotionRejectsInvalidProductId(t *testing.T) {
	entity := &promotionEntity{}

	if _, err := entity.CreatePromotion(request.Promotion{
		BranchId:   "507f1f77bcf86cd799439011",
		ProductIds: []string{"invalid-product-id"},
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
}

func TestGetPromotionsRejectsInvalidBranchId(t *testing.T) {
	entity := &promotionEntity{}

	if _, err := entity.GetPromotions("invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetPromotionByCodeRejectsInvalidBranchId(t *testing.T) {
	entity := &promotionEntity{}

	if _, err := entity.GetPromotionByCode("PROMO", "invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestUpdatePromotionByIdRejectsInvalidProductId(t *testing.T) {
	entity := &promotionEntity{}

	if _, err := entity.UpdatePromotionById("507f1f77bcf86cd799439011", request.UpdatePromotion{
		ProductIds: []string{"invalid-product-id"},
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
}

func TestCreateStockTransferRejectsInvalidBranchId(t *testing.T) {
	entity := &stockTransferEntity{}

	if _, err := entity.createStockTransferWithContext(nil, request.StockTransfer{
		FromBranchId: "invalid-branch-id",
		ToBranchId:   "507f1f77bcf86cd799439011",
	}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetStockTransfersRejectsInvalidBranchId(t *testing.T) {
	entity := &stockTransferEntity{}

	if _, err := entity.GetStockTransfers("invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateStockTransferRejectsInvalidProductId(t *testing.T) {
	entity := &stockTransferEntity{}

	if _, err := entity.createStockTransferWithContext(nil, request.StockTransfer{
		FromBranchId: "507f1f77bcf86cd799439011",
		ToBranchId:   "507f1f77bcf86cd799439012",
		Items: []request.StockTransferItem{
			{ProductId: "invalid-product-id", StockId: "stock-1", Quantity: 1},
		},
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
}
