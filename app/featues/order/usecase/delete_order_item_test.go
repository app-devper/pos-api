package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cancelOrderRepoStub struct {
	repositories.IOrder
	getOrderByIDFn           func(id string) (*entities.Order, error)
	getOrderItemByIDFn       func(id string) (*entities.OrderItem, error)
	cancelOrderByIDFn        func(id string, userId string, branchId string) (*entities.OrderDetail, error)
	cancelByIDFn             func(id string, userId string, branchId string) (*entities.OrderItemProductDetail, error)
	cancelByOrderProductIDFn func(orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error)
}

func (s *cancelOrderRepoStub) GetOrderById(id string) (*entities.Order, error) {
	return s.getOrderByIDFn(id)
}

func (s *cancelOrderRepoStub) GetOrderItemById(id string) (*entities.OrderItem, error) {
	return s.getOrderItemByIDFn(id)
}

func (s *cancelOrderRepoStub) CancelOrderById(id string, userId string, branchId string) (*entities.OrderDetail, error) {
	return s.cancelOrderByIDFn(id, userId, branchId)
}

func (s *cancelOrderRepoStub) CancelOrderItemById(id string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	return s.cancelByIDFn(id, userId, branchId)
}

func (s *cancelOrderRepoStub) CancelOrderItemByOrderProductId(orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
	return s.cancelByOrderProductIDFn(orderId, productId, userId, branchId)
}

func TestDeleteOrderItemByIdUsesTransactionalCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	itemID := primitive.NewObjectID().Hex()
	orderID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	var gotID, gotUser, gotBranch string

	repo := &cancelOrderRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			t.Fatal("order lookup should not run for item delete by id")
			return nil, nil
		},
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			return &entities.OrderItem{Id: primitive.NewObjectID(), BranchId: branchID, OrderId: orderID, ProductId: productID}, nil
		},
		cancelByIDFn: func(id string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
			gotID, gotUser, gotBranch = id, userId, branchId
			return &entities.OrderItemProductDetail{OrderId: orderID, ProductId: productID}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/items/"+itemID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "itemId", Value: itemID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	DeleteOrderItemById(repo, nil)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != itemID || gotUser != "user-1" || gotBranch == "" {
		t.Fatalf("expected transactional cancel to receive item/user/branch, got id=%s user=%s branch=%s", gotID, gotUser, gotBranch)
	}
}

func TestDeleteOrderByIdUsesTransactionalCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()
	var gotID, gotUser, gotBranch string

	repo := &cancelOrderRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: branchID}, nil
		},
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			t.Fatal("item lookup should not run for order delete")
			return nil, nil
		},
		cancelOrderByIDFn: func(id string, userId string, branchId string) (*entities.OrderDetail, error) {
			gotID, gotUser, gotBranch = id, userId, branchId
			return &entities.OrderDetail{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/"+orderID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "orderId", Value: orderID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	DeleteOrderById(repo, nil)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != orderID || gotUser != "user-1" || gotBranch == "" {
		t.Fatalf("expected transactional order cancel to receive order/user/branch, got id=%s user=%s branch=%s", gotID, gotUser, gotBranch)
	}
}

func TestDeleteOrderItemByOrderProductIdReturnsTransactionalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()

	repo := &cancelOrderRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: branchID}, nil
		},
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			t.Fatal("item lookup should not run for order/product delete")
			return nil, nil
		},
		cancelByOrderProductIDFn: func(orderId string, productId string, userId string, branchId string) (*entities.OrderItemProductDetail, error) {
			return nil, errors.New("cancel failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/"+orderID+"/products/"+productID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "orderId", Value: orderID},
		{Key: "productId", Value: productID},
	}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	DeleteOrderItemByOrderProductId(repo, nil)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.OR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.OR_BAD_REQUEST_002, w.Body.String())
	}
}

func TestDeleteOrderByIdReturnsTransactionalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()

	repo := &cancelOrderRepoStub{
		getOrderByIDFn: func(id string) (*entities.Order, error) {
			return &entities.Order{Id: primitive.NewObjectID(), BranchId: branchID}, nil
		},
		getOrderItemByIDFn: func(id string) (*entities.OrderItem, error) {
			t.Fatal("item lookup should not run for order delete")
			return nil, nil
		},
		cancelOrderByIDFn: func(id string, userId string, branchId string) (*entities.OrderDetail, error) {
			return nil, errors.New("cancel failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/orders/"+orderID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "orderId", Value: orderID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	DeleteOrderById(repo, nil)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.OR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.OR_BAD_REQUEST_002, w.Body.String())
	}
}
