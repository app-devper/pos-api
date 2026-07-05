package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type orderAccessRepoStub struct {
	repositories.IOrder
	getOrderByIDFn           func(id string) (*entities.Order, error)
	getOrderDetailByIDFn     func(id string) (*entities.OrderDetail, error)
	getOrderItemByIDFn       func(id string) (*entities.OrderItem, error)
	getOrderItemDetailByIDFn func(id string) (*entities.OrderItemProductDetail, error)
	getItemsByProductFn      func(productId string, branchId string) ([]entities.OrderItem, error)
	getItemDetailsByProdFn   func(productId string, branchId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error)
	updateCustomerCodeFn     func(id string, customerCode string) (*entities.Order, error)
	cancelOrderByIDFn        func(id string, userId string, branchId string, reason string) (*entities.OrderDetail, error)
	cancelItemByIDFn         func(id string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error)
	cancelByOrderProductFn   func(orderId string, productId string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error)
}

func (s *orderAccessRepoStub) GetOrderById(id string) (*entities.Order, error) {
	return s.getOrderByIDFn(id)
}

func (s *orderAccessRepoStub) GetOrderDetailById(id string) (*entities.OrderDetail, error) {
	return s.getOrderDetailByIDFn(id)
}

func (s *orderAccessRepoStub) GetOrderItemById(id string) (*entities.OrderItem, error) {
	return s.getOrderItemByIDFn(id)
}

func (s *orderAccessRepoStub) GetOrderItemDetailById(id string) (*entities.OrderItemProductDetail, error) {
	return s.getOrderItemDetailByIDFn(id)
}

func (s *orderAccessRepoStub) GetOrderItemByProductId(productId string, branchId string) ([]entities.OrderItem, error) {
	return s.getItemsByProductFn(productId, branchId)
}

func (s *orderAccessRepoStub) GetOrderItemOrderDetailsByProductId(productId string, branchId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error) {
	return s.getItemDetailsByProdFn(productId, branchId, form)
}

func (s *orderAccessRepoStub) UpdateCustomerCodeOrderById(id string, customerCode string) (*entities.Order, error) {
	return s.updateCustomerCodeFn(id, customerCode)
}

func (s *orderAccessRepoStub) CancelOrderById(id string, userId string, branchId string, reason string) (*entities.OrderDetail, error) {
	return s.cancelOrderByIDFn(id, userId, branchId, reason)
}

func (s *orderAccessRepoStub) CancelOrderItemById(id string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
	return s.cancelItemByIDFn(id, userId, branchId, reason)
}

func (s *orderAccessRepoStub) CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
	return s.cancelByOrderProductFn(orderId, productId, userId, branchId, reason)
}

func TestGetOrderByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		getOrderDetailByIDFn: func(id string) (*entities.OrderDetail, error) {
			t.Fatal("detail lookup should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "orderId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetOrderById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.SY_FORBIDDEN_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.SY_FORBIDDEN_002, w.Body.String())
	}
}

func TestGetOrderItemByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			return &entities.OrderItem{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		getOrderItemDetailByIDFn: func(id string) (*entities.OrderItemProductDetail, error) {
			t.Fatal("item detail lookup should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/items/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "itemId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetOrderItemById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateCustomerCodeOrderByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateCustomerCodeFn: func(id string, customerCode string) (*entities.Order, error) {
			t.Fatal("customer code update should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/orders/1/customer-code", strings.NewReader(`{"customerCode":"C001"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "orderId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	UpdateCustomerCodeOrderById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteOrderByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		cancelOrderByIDFn: func(id string, userId string, branchId string, reason string) (*entities.OrderDetail, error) {
			t.Fatal("order cancel should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "orderId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	DeleteOrderById(repo, nil)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteOrderItemByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			return &entities.OrderItem{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		cancelItemByIDFn: func(id string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
			t.Fatal("order item cancel should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/items/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "itemId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	DeleteOrderItemById(repo, nil)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteOrderItemByOrderProductIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &orderAccessRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		cancelByOrderProductFn: func(orderId string, productId string, userId string, branchId string, reason string) (*entities.OrderItemProductDetail, error) {
			t.Fatal("order item by product cancel should not run for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/1/products/2", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "orderId", Value: primitive.NewObjectID().Hex()},
		{Key: "productId", Value: primitive.NewObjectID().Hex()},
	}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	DeleteOrderItemByOrderProductId(repo, nil)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestGetOrderItemByProductIdUsesBranchScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	var gotBranchID string

	repo := &orderAccessRepoStub{
		getItemsByProductFn: func(productId string, branchId string) ([]entities.OrderItem, error) {
			gotBranchID = branchId
			return []entities.OrderItem{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/items/products/"+productID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "productId", Value: productID}}
	ctx.Set("BranchId", branchID)

	GetOrderItemByProductId(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch %s, got %s", branchID, gotBranchID)
	}
}

func TestGetOrderItemDetailsByProductIdUsesBranchScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	var gotBranchID string

	repo := &orderAccessRepoStub{
		getItemDetailsByProdFn: func(productId string, branchId string, form request.GetOrderRange) ([]entities.OrderItemOrderDetail, error) {
			gotBranchID = branchId
			return []entities.OrderItemOrderDetail{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/item-details/products/"+productID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "productId", Value: productID}}
	ctx.Set("BranchId", branchID)

	GetOrderItemDetailsByProductId(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch %s, got %s", branchID, gotBranchID)
	}
}
