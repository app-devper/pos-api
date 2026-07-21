package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ImportReceiveToStock(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return ImportReceiveToStockWithReconciliation(receiveEntity, productEntity, nil, nil)
}

func ImportReceiveToStockWithReconciliation(
	receiveEntity repositories.IReceive,
	_ repositories.IProduct,
	orderEntity repositories.IOrder,
	productStock repositories.IProductStock,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		receiveId := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)
		receive, err := receiveEntity.GetReceiveById(receiveId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureReceiveBranchAccess(receive, branchId); err != nil {
			abortReceiveBranchMismatch(ctx)
			return
		}
		result, err := receiveEntity.ImportReceiveToStock(receiveId, userId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		if orderEntity != nil && productStock != nil {
			reconcileOversoldFromImport(orderEntity, productStock, result, branchId)
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func reconcileOversoldFromImport(
	orderEntity repositories.IOrder,
	productStock repositories.IProductStock,
	receive *entities.Receive,
	branchId string,
) {
	if receive == nil {
		return
	}
	for _, item := range receive.Items {
		if item.Quantity <= 0 {
			continue
		}
		productId := item.ProductId.Hex()
		oversoldItems, err := orderEntity.GetOversoldOrderItemsByProductId(productId, branchId)
		if err != nil {
			logrus.WithError(err).WithField("productId", productId).Warn("failed to look up oversold order items during import reconciliation")
			continue
		}
		if len(oversoldItems) == 0 {
			continue
		}
		lots, err := productStock.GetProductStocksByProductIdAndReceiveCode(productId, receive.Code, branchId)
		if err != nil {
			logrus.WithError(err).WithField("productId", productId).Warn("failed to look up newly imported lots during import reconciliation")
			continue
		}
		drainOversoldAgainstLots(orderEntity, productStock, oversoldItems, lots)
	}
}

func drainOversoldAgainstLots(
	orderEntity repositories.IOrder,
	productStock repositories.IProductStock,
	oversoldItems []entities.OrderItem,
	lots []entities.ProductStock,
) {
	for i := range lots {
		available := lots[i].Quantity
		if available <= 0 {
			continue
		}
		lotId := lots[i].Id.Hex()
		for j := range oversoldItems {
			if available <= 0 {
				break
			}
			remaining := oversoldItems[j].OversoldQty
			if remaining <= 0 {
				continue
			}
			drain := remaining
			if available < drain {
				drain = available
			}
			if _, err := productStock.RemoveProductStockQuantityById(lotId, drain); err != nil {
				logrus.WithError(err).WithField("stockId", lotId).Warn("failed to drain lot during oversold reconciliation")
				continue
			}
			if _, err := orderEntity.DrainOversoldQtyByOrderItemId(oversoldItems[j].Id.Hex(), drain, lotId); err != nil {
				logrus.WithError(err).WithField("orderItemId", oversoldItems[j].Id.Hex()).Warn("failed to record oversold reconciliation on order item")
				_, _ = productStock.AddProductStockQuantityById(lotId, drain)
				continue
			}
			available -= drain
			oversoldItems[j].OversoldQty -= drain
		}
	}
}
