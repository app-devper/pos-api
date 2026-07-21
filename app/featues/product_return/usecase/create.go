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

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateProductReturn(
	productReturnEntity repositories.IProductReturn,
	orderEntity repositories.IOrder,
	productStock repositories.IProductStock,
	productEntity repositories.IProduct,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductReturn{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = utils.GetBranchId(ctx)
		req.CreatedBy = utils.GetUserId(ctx)

		order, err := orderEntity.GetOrderById(req.OrderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}
		if order == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, "order not found")
			return
		}
		if req.BranchId != "" && order.BranchId.Hex() != req.BranchId {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, "order does not belong to this branch")
			return
		}

		type plannedLine struct {
			item        *entities.OrderItem
			quantity    int
			refund      float64
			allocations []entities.OrderItemStock
		}
		planned := make([]plannedLine, 0, len(req.Items))

		for _, line := range req.Items {
			if line.Quantity <= 0 {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_001, "return quantity must be greater than zero")
				return
			}
			item, err := orderEntity.GetOrderItemById(line.OrderItemId)
			if err != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
				return
			}
			if item == nil || item.OrderId.Hex() != order.Id.Hex() {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, fmt.Sprintf("order item %s not found on order %s", line.OrderItemId, req.OrderId))
				return
			}

			realQty := realLotQuantity(item.Stocks)
			if item.ReturnedQty+line.Quantity > realQty {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, fmt.Sprintf(
					"คืนได้สูงสุด %d (มี %d หน่วยยังไม่ผูกกับล็อตจริง)",
					maxInt(realQty-item.ReturnedQty, 0), item.Quantity-realQty,
				))
				return
			}
			if item.ReturnedQty+line.Quantity > item.Quantity {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, fmt.Sprintf("คืนได้สูงสุด %d", item.Quantity-item.ReturnedQty))
				return
			}

			allocations := allocateReturnAcrossRealLots(item.Stocks, item.ReturnedQty, line.Quantity)
			planned = append(planned, plannedLine{item: item, quantity: line.Quantity, refund: line.Refund, allocations: allocations})
		}

		sequence, err := sequenceEntity.NextSequence(constant.PRODUCT_RETURN)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, "product return sequence not available")
			return
		}
		returnNo := "RT-" + sequence.GenerateCode()

		items := make([]entities.ProductReturnItem, 0, len(planned))
		var totalRefund float64
		for _, line := range planned {
			for _, allocation := range line.allocations {
				if _, err := productStock.AddProductStockQuantityById(allocation.StockId, allocation.Quantity); err != nil {
					errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, fmt.Sprintf("failed to restore stock %s: %s", allocation.StockId, err.Error()))
					return
				}
			}
			if _, err := orderEntity.IncrementOrderItemReturnedQtyById(line.item.Id.Hex(), line.quantity); err != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, fmt.Sprintf("failed to record returned quantity for order item %s: %s", line.item.Id.Hex(), err.Error()))
				return
			}

			unit, err := productEntity.GetProductUnitById(line.item.UnitId.Hex())
			if err == nil && unit != nil {
				balance := productStock.GetProductStockBalance(line.item.ProductId.Hex(), unit.Id.Hex(), req.BranchId)
				history := request.ProductReturnHistory(line.item.ProductId.Hex(), unit.Unit, line.quantity, req.Reason, balance, req.CreatedBy)
				history.BranchId = req.BranchId
				_, _ = productStock.CreateProductHistory(history)
			}

			items = append(items, entities.ProductReturnItem{
				OrderItemId: line.item.Id,
				ProductId:   line.item.ProductId,
				Quantity:    line.quantity,
				Price:       line.item.Price,
				Refund:      line.refund,
			})
			totalRefund += line.refund
		}

		branchObjId, err := primitive.ObjectIDFromHex(req.BranchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_001, err.Error())
			return
		}

		result, err := productReturnEntity.CreateProductReturn(repositories.ProductReturnInput{
			BranchId:     branchObjId,
			ReturnNo:     returnNo,
			OrderId:      order.Id,
			CustomerCode: order.CustomerCode,
			Reason:       req.Reason,
			Note:         req.Note,
			Items:        items,
			TotalRefund:  totalRefund,
			CreatedBy:    req.CreatedBy,
		})
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductReturnsByOrderId(productReturnEntity repositories.IProductReturn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		result, err := productReturnEntity.GetProductReturnsByOrderId(orderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
