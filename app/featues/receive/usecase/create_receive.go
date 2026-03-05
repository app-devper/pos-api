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

func CreateReceive(receiveEntity repositories.IReceive, sequenceEntity repositories.ISequence, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Receive{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)

		sequence, _ := sequenceEntity.NextSequence(constant.RECEIVE)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}
		req.UpdatedBy = userId
		req.BranchId = branchId

		result, err := receiveEntity.CreateReceive(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		receiveId := result.Id.Hex()
		var totalCost float64

		for _, item := range req.Items {
			if item.ProductId == "" || item.Quantity <= 0 {
				continue
			}

			product, pErr := productEntity.GetProductById(item.ProductId)
			if pErr != nil || product == nil {
				logrus.Warnf("CreateReceive: product %s not found, skipping", item.ProductId)
				continue
			}

			productReq := request.Product{
				Name:         product.Name,
				SerialNumber: product.SerialNumber,
				Price:        product.Price,
				CostPrice:    item.CostPrice,
				Unit:         product.Unit,
				Quantity:     item.Quantity,
				LotNumber:    item.LotNumber,
				ExpireDate:   time.Time{},
				ReceiveId:    receiveId,
				ReceiveCode:  req.Code,
				CreatedBy:    userId,
				BranchId:     branchId,
			}
			if item.ExpireDate != "" {
				if t, tErr := time.Parse(time.RFC3339, item.ExpireDate); tErr == nil {
					productReq.ExpireDate = t
				} else if t, tErr := time.Parse("2006-01-02", item.ExpireDate); tErr == nil {
					productReq.ExpireDate = t
				}
			}

			_, _ = receiveEntity.CreateReceiveItem(receiveId, "", item.ProductId, productReq)

			unit, _ := productEntity.GetProductUnitByUnit(item.ProductId, product.Unit)
			if unit != nil && item.Quantity > 0 {
				stock := request.ProductStock{
					ProductId:   item.ProductId,
					UnitId:      unit.Id.Hex(),
					ReceiveCode: req.Code,
					Quantity:    item.Quantity,
					CostPrice:   item.CostPrice,
					ExpireDate:  productReq.ExpireDate,
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

		ctx.JSON(http.StatusOK, result)
	}
}
