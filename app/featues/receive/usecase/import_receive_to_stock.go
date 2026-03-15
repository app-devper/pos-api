package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ImportReceiveToStock(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		receiveId := ctx.Param("receiveId")

		receive, err := receiveEntity.GetReceiveById(receiveId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		if receive.Status == constant.IMPORTED {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, "receive already imported")
			return
		}

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)

		items, err := receiveEntity.GetReceiveItemsByReceiveId(receiveId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		var totalCost float64

		for _, item := range items {
			productId := item.ProductId.Hex()
			product, pErr := productEntity.GetProductById(productId)
			if pErr != nil || product == nil {
				logrus.Warnf("ImportReceiveToStock: product %s not found, skipping", productId)
				continue
			}

			unit, _ := productEntity.GetProductUnitByUnit(productId, product.Unit)
			if unit != nil && item.Quantity > 0 {
				stock := request.ProductStock{
					ProductId:   productId,
					UnitId:      unit.Id.Hex(),
					ReceiveCode: receive.Code,
					Quantity:    item.Quantity,
					CostPrice:   item.CostPrice,
					ExpireDate:  item.ExpireDate,
					LotNumber:   item.LotNumber,
					ImportDate:  time.Now(),
					UpdatedBy:   userId,
					BranchId:    branchId,
				}
				created, _ := productEntity.CreateProductStock(stock)
				if created != nil {
					balance := productEntity.GetProductStockBalance(created.ProductId.Hex(), created.UnitId.Hex())
					hist := request.AddProductStockHistory(created.ProductId.Hex(), product.Unit, stock, balance)
					hist.BranchId = branchId
					_, _ = productEntity.CreateProductHistory(hist)
				}
			}

			totalCost += item.CostPrice * float64(item.Quantity)
		}

		if totalCost > 0 {
			_, _ = receiveEntity.UpdateReceiveTotalCostById(receiveId, totalCost)
		}

		result, _ := receiveEntity.UpdateReceiveStatusById(receiveId, constant.IMPORTED, userId)
		if result == nil {
			result = receive
			result.Status = constant.IMPORTED
		}

		resultItems, _ := receiveEntity.GetReceiveItemsByReceiveId(receiveId)
		if len(resultItems) > 0 {
			result.Items = resultItems
		}

		ctx.JSON(http.StatusOK, result)
	}
}
