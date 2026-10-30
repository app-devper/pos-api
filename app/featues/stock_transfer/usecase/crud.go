package usecase

import (
	"context"
	"net/http"

	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetStockTransfers(entity repositories.IStockTransfer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetStockTransfers(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockTransferById(entity repositories.IStockTransfer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetStockTransferById(id, ctx.GetString("BranchId"))
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.TR_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// transferDecider approves or rejects a Transfer through the Stock ledger.
type transferDecider interface {
	ApproveTransfer(ctx context.Context, transferID, by string) (*entities.StockTransfer, error)
	RejectTransfer(ctx context.Context, transferID, by string) (*entities.StockTransfer, error)
}

func ApproveStockTransfer(entity repositories.IStockTransfer, ledger transferDecider) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateStockTransfer{
			Status:    "APPROVED",
			UpdatedBy: utils.GetUserId(ctx),
		}

		branchId := ctx.GetString("BranchId")
		transfer, err := entity.GetStockTransferById(id, branchId)
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.TR_BAD_REQUEST_002)
			return
		}
		// The branch that receives the Stock approves it; the branch that
		// asked cannot approve its own request.
		if transfer.ToBranchId.Hex() != branchId {
			errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "only the receiving branch can approve a transfer")
			return
		}

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := ledger.ApproveTransfer(ctx.Request.Context(), id, req.UpdatedBy)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.TR_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func RejectStockTransfer(entity repositories.IStockTransfer, ledger transferDecider) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateStockTransfer{
			Status:    "REJECTED",
			UpdatedBy: utils.GetUserId(ctx),
		}

		transfer, err := entity.GetStockTransferById(id, ctx.GetString("BranchId"))
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.TR_BAD_REQUEST_002)
			return
		}

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := ledger.RejectTransfer(ctx.Request.Context(), id, req.UpdatedBy)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.TR_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
