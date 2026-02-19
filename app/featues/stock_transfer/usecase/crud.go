package usecase

import (
	"net/http"
	"time"

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

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := entity.UpdateStockTransferStatus(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		// Add stock to destination branch by creating new stock entries
		for _, item := range transfer.Items {
			if item.StockId != "" {
				sourceStock, err := productEntity.GetProductStockById(item.StockId)
				if err == nil && sourceStock != nil {
					_, _ = productEntity.CreateProductStock(request.ProductStock{
						BranchId:   transfer.ToBranchId.Hex(),
						ProductId:  item.ProductId.Hex(),
						UnitId:     sourceStock.UnitId.Hex(),
						LotNumber:  sourceStock.LotNumber,
						CostPrice:  sourceStock.CostPrice,
						Price:      sourceStock.Price,
						Quantity:   item.Quantity,
						ExpireDate: sourceStock.ExpireDate,
						ImportDate: time.Now(),
					})
				}
			}
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

		if transfer.Status != "PENDING" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "transfer is not pending")
			return
		}

		result, err := entity.UpdateStockTransferStatus(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		for _, item := range transfer.Items {
			if item.StockId != "" {
				_, _ = productEntity.AddProductStockQuantityById(item.StockId, item.Quantity)
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
