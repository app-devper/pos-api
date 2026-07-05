package usecase

import (
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"

	"github.com/gin-gonic/gin"
)

func abortOrderBranchMismatch(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
}

func ensureOrderBranchAccess(order *entities.Order, branchId string) error {
	if order == nil {
		return errors.New("order not found")
	}
	if order.BranchId.Hex() != branchId {
		return errors.New("order belongs to another branch")
	}
	return nil
}

func ensureOrderItemBranchAccess(item *entities.OrderItem, branchId string) error {
	if item == nil {
		return errors.New("order item not found")
	}
	if item.BranchId.Hex() != branchId {
		return errors.New("order item belongs to another branch")
	}
	return nil
}
