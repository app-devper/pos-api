package usecase

import (
	"net/http"

	"pos/app/core/errcode"
	"pos/app/core/utils"
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
		result, err := entity.GetStockTransferById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		if result.FromBranchId.Hex() != branchId && result.ToBranchId.Hex() != branchId {
			errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func ApproveStockTransfer(entity repositories.IStockTransfer, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateStockTransfer{
			Status:    "APPROVED",
			UpdatedBy: utils.GetUserId(ctx),
		}

		transfer, err := entity.GetStockTransferById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		if transfer.FromBranchId.Hex() != branchId && transfer.ToBranchId.Hex() != branchId {
			errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
			return
		}

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := entity.ApproveStockTransfer(id, req.UpdatedBy)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func RejectStockTransfer(entity repositories.IStockTransfer, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateStockTransfer{
			Status:    "REJECTED",
			UpdatedBy: utils.GetUserId(ctx),
		}

		transfer, err := entity.GetStockTransferById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		if transfer.FromBranchId.Hex() != branchId && transfer.ToBranchId.Hex() != branchId {
			errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
			return
		}

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := entity.RejectStockTransfer(id, req.UpdatedBy)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
