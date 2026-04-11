package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateProductPrice(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}

		customerTypes := constant.CustomerTypes()
		if customerTypeIsValid := utils.InArrayString(req.CustomerType, customerTypes); !customerTypeIsValid {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, "customer type is not valid")
			return
		}

		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.CreateProductPriceCascade(req, ctx.GetString("BranchId"), userId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductPricesByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductPricesByProductId(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductPriceById(productEntity repositories.IProduct, productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		id := ctx.Param("priceId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}

		customerTypes := constant.CustomerTypes()
		if customerTypeIsValid := utils.InArrayString(req.CustomerType, customerTypes); !customerTypeIsValid {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, "customer type is not valid")
			return
		}

		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.UpdateProductPriceById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		appendProductPriceHistory(ctx.GetString("BranchId"), productEntity, productStock, req.UnitId, func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			updPriceHistory := request.UpdateProductPriceHistory(req.ProductId, unit.Unit, req)
			updPriceHistory.BranchId = branchId
			return updPriceHistory
		})

		ctx.JSON(http.StatusOK, result)
	}
}

func RemoveProductPriceById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("priceId")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductPriceCascade(id, ctx.GetString("BranchId"), userId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func appendProductPriceHistory(branchId string, productEntity repositories.IProduct, productStock repositories.IProductStock, unitId string, buildHistory func(unit *entities.ProductUnit, branchId string) request.ProductHistory) {
	unit, err := productEntity.GetProductUnitById(unitId)
	if err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to load product unit for price history")
		return
	}
	if unit == nil {
		logrus.WithField("unitId", unitId).Warn("product unit missing for price history")
		return
	}

	history := buildHistory(unit, branchId)
	if _, err = productStock.CreateProductHistory(history); err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to create product price history")
	}
}
