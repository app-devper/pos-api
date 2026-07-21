package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	stockadjustment "pos/app/featues/stock_adjustment/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateStockCount(
	stockCountEntity repositories.IStockCount,
	stockAdjustmentEntity repositories.IStockAdjustment,
	productStock repositories.IProductStock,
	productEntity repositories.IProduct,
	orderEntity repositories.IOrder,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockCount{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = utils.GetBranchId(ctx)

		sequence, err := sequenceEntity.NextSequence(constant.STOCK_COUNT)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, "stock count sequence not available")
			return
		}
		countNo := "SC-" + sequence.GenerateCode()

		items := make([]entities.StockCountItem, 0, len(req.Items))
		for _, line := range req.Items {
			stock, err := productStock.GetProductStockById(line.StockId)
			if err != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, fmt.Sprintf("failed to load stock %s: %s", line.StockId, err.Error()))
				return
			}
			if stock == nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, fmt.Sprintf("stock %s not found", line.StockId))
				return
			}

			systemQty := stock.Quantity
			delta := line.Counted - systemQty

			if delta != 0 {
				adjustReq := request.StockAdjustment{
					ProductId: line.ProductId,
					StockId:   line.StockId,
					Reason:    constant.AdjustmentReasonCount,
					Note:      countNo + " " + req.Note,
					Delta:     delta,
					BranchId:  req.BranchId,
					CreatedBy: req.CreatedBy,
				}
				if _, err := stockadjustment.ApplyAdjustment(stockAdjustmentEntity, productStock, productEntity, orderEntity, sequenceEntity, adjustReq); err != nil {
					errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, fmt.Sprintf("failed to apply count adjustment for product %s: %s", line.ProductId, err.Error()))
					return
				}
			}

			productObjId, err := primitive.ObjectIDFromHex(line.ProductId)
			if err != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_001, err.Error())
				return
			}
			stockObjId, err := primitive.ObjectIDFromHex(line.StockId)
			if err != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_001, err.Error())
				return
			}
			items = append(items, entities.StockCountItem{
				ProductId:       productObjId,
				StockId:         stockObjId,
				SystemQuantity:  systemQty,
				CountedQuantity: line.Counted,
				Delta:           delta,
			})
		}

		branchObjId, err := primitive.ObjectIDFromHex(req.BranchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_001, err.Error())
			return
		}

		result, err := stockCountEntity.CreateStockCount(repositories.StockCountInput{
			BranchId:  branchObjId,
			CountNo:   countNo,
			Note:      req.Note,
			Items:     items,
			CreatedBy: req.CreatedBy,
		})
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockCounts(stockCountEntity repositories.IStockCount) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := utils.GetBranchId(ctx)
		result, err := stockCountEntity.GetStockCounts(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockCountById(stockCountEntity repositories.IStockCount) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := stockCountEntity.GetStockCountById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
