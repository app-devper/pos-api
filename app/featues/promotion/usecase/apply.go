package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func ApplyPromotion(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ApplyPromotion{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")

		promo, err := entity.GetPromotionByCode(req.PromotionCode, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, "promotion not found or expired")
			return
		}

		if promo.MinPurchase > 0 && req.OrderTotal < promo.MinPurchase {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, "order total below minimum purchase")
			return
		}

		if len(promo.ProductIds) > 0 && len(req.ProductIds) > 0 {
			promoMap := make(map[string]bool)
			for _, pid := range promo.ProductIds {
				promoMap[pid.Hex()] = true
			}
			hasMatch := false
			for _, pid := range req.ProductIds {
				if promoMap[pid] {
					hasMatch = true
					break
				}
			}
			if !hasMatch {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, "no matching products for this promotion")
				return
			}
		}

		var discount float64
		switch promo.Type {
		case "PERCENTAGE":
			discount = req.OrderTotal * promo.Value / 100
			if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
				discount = promo.MaxDiscount
			}
		case "FIXED":
			discount = promo.Value
			if discount > req.OrderTotal {
				discount = req.OrderTotal
			}
		default:
			discount = 0
		}

		result := request.ApplyPromotionResult{
			PromotionId: promo.Id.Hex(),
			Code:        promo.Code,
			Name:        promo.Name,
			Type:        promo.Type,
			Discount:    discount,
		}
		ctx.JSON(http.StatusOK, result)
	}
}
