package usecase

import (
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"

	"github.com/gin-gonic/gin"
)

func abortProductBranchMismatch(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
}

func ensureProductStockBranchAccess(stock *entities.ProductStock, branchId string) error {
	if stock == nil {
		return errors.New("product stock not found")
	}
	if stock.BranchId.Hex() != branchId {
		return errors.New("product stock belongs to another branch")
	}
	return nil
}
