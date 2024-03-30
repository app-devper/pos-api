package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func CreateOrder(
	orderEntity repositories.IOrder,
	productEntity repositories.IProduct,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Order{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, _ := sequenceEntity.NextSequence(constant.ORDER)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}

		result, err := orderEntity.CreateOrder(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update product stock
		var stocks []entities.ProductStock
		for _, item := range req.Items {
			if len(item.Stocks) > 0 {
				// Update stock quantity
				for _, itemStock := range item.Stocks {
					if itemStock.StockId != "" {
						stock, _ := productEntity.RemoveProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
						stocks = append(stocks, *stock)
					} else {
						_, _ = productEntity.RemoveQuantitySoldFirstById(item.ProductId, itemStock.Quantity)
					}
				}

				// Add product history
				unit, _ := productEntity.GetProductUnitById(item.UnitId)
				balance := productEntity.GetProductStockBalance(item.ProductId, unit.Id.Hex())
				_, _ = productEntity.CreateProductHistory(request.AddOrderItemProductHistory(item.ProductId, unit.Unit, item, balance, req.CreatedBy))
			}
		}

		if req.Message != "" {
			date := utils.ToFormat(result.CreatedDate)
			_, _ = utils.NotifyMassage("รายการวันที่ " + date + "\n\n" + req.Message)
		}

		ctx.JSON(http.StatusOK, gin.H{"data": result, "stocks": stocks})
	}
}
