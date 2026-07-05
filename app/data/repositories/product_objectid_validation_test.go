package repositories

import (
	"context"
	"testing"
	"time"

	"pos/app/domain/request"
)

func TestCreateProductLotByProductIdRejectsInvalidProductID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.CreateProductLotByProductId("invalid-id", request.Product{}); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestCreateProductLotRejectsInvalidProductID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.CreateProductLot(request.ProductLot{ProductId: "invalid-id"}); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestCreateProductLotRejectsInvalidBranchID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.CreateProductLot(request.ProductLot{ProductId: "507f1f77bcf86cd799439011", BranchId: "invalid-id"}); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestCreateProductUnitByProductIdRejectsInvalidProductID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.CreateProductUnitByProductId("invalid-id", request.Product{}); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestCreateProductPriceWithContextRejectsInvalidIDs(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.createProductPriceWithContext(context.Background(), request.ProductPrice{
		ProductId: "invalid-id",
		UnitId:    "507f1f77bcf86cd799439012",
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}

	if _, err := entity.createProductPriceWithContext(context.Background(), request.ProductPrice{
		ProductId: "507f1f77bcf86cd799439011",
		UnitId:    "invalid-id",
	}); err == nil {
		t.Fatal("expected invalid unit id error")
	}
}

func TestCreateProductUnitWithContextRejectsInvalidProductID(t *testing.T) {
	entity := &productEntity{}

	if _, err := entity.createProductUnitWithContext(context.Background(), request.ProductUnit{ProductId: "invalid-id"}); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestCreateProductStockWithContextRejectsInvalidIDs(t *testing.T) {
	entity := &productStockEntity{}
	param := request.ProductStock{
		BranchId:   "507f1f77bcf86cd799439011",
		ProductId:  "507f1f77bcf86cd799439012",
		UnitId:     "507f1f77bcf86cd799439013",
		ExpireDate: request.NewFlexibleTime(time.Now()),
		ImportDate: request.NewFlexibleTime(time.Now()),
	}

	param.BranchId = "invalid-id"
	if _, err := entity.createProductStockWithContext(context.Background(), param); err == nil {
		t.Fatal("expected invalid branch id error")
	}

	param.BranchId = "507f1f77bcf86cd799439011"
	param.ProductId = "invalid-id"
	if _, err := entity.createProductStockWithContext(context.Background(), param); err == nil {
		t.Fatal("expected invalid product id error")
	}

	param.ProductId = "507f1f77bcf86cd799439012"
	param.UnitId = "invalid-id"
	if _, err := entity.createProductStockWithContext(context.Background(), param); err == nil {
		t.Fatal("expected invalid unit id error")
	}
}

func TestGetProductStockByIdRejectsInvalidObjectID(t *testing.T) {
	entity := &productStockEntity{}

	if _, err := entity.GetProductStockById("invalid-id"); err == nil {
		t.Fatal("expected invalid object id error")
	}
}

func TestGetProductStocksByProductIdRejectsInvalidIDs(t *testing.T) {
	entity := &productStockEntity{}

	if _, err := entity.GetProductStocksByProductId("invalid-id", ""); err == nil {
		t.Fatal("expected invalid product id error")
	}

	if _, err := entity.GetProductStocksByProductId("507f1f77bcf86cd799439011", "invalid-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateProductHistoryWithContextRejectsInvalidIDs(t *testing.T) {
	entity := &productStockEntity{}

	if _, err := entity.createProductHistoryWithContext(context.Background(), request.ProductHistory{
		BranchId:  "invalid-id",
		ProductId: "507f1f77bcf86cd799439011",
	}); err == nil {
		t.Fatal("expected invalid branch id error")
	}

	if _, err := entity.createProductHistoryWithContext(context.Background(), request.ProductHistory{
		BranchId:  "507f1f77bcf86cd799439011",
		ProductId: "invalid-id",
	}); err == nil {
		t.Fatal("expected invalid product id error")
	}
}

func TestGetProductStockBalanceWithContextRejectsInvalidIDs(t *testing.T) {
	entity := &productStockEntity{}

	if _, err := entity.getProductStockBalanceWithContext(context.Background(), "invalid-id", "507f1f77bcf86cd799439011", ""); err == nil {
		t.Fatal("expected invalid product id error")
	}

	if _, err := entity.getProductStockBalanceWithContext(context.Background(), "507f1f77bcf86cd799439011", "invalid-id", ""); err == nil {
		t.Fatal("expected invalid unit id error")
	}

	if _, err := entity.getProductStockBalanceWithContext(context.Background(), "507f1f77bcf86cd799439011", "507f1f77bcf86cd799439012", "invalid-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}
