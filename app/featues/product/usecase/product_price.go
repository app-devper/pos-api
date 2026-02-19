package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
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

		result, err := productEntity.CreateProductPrice(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		if unit != nil {
			addPriceHistory := request.AddProductPriceHistory(req.ProductId, unit.Unit, req)
			addPriceHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(addPriceHistory)
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

func UpdateProductPriceById(productEntity repositories.IProduct) gin.HandlerFunc {
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
		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		if unit != nil {
			updPriceHistory := request.UpdateProductPriceHistory(req.ProductId, unit.Unit, req)
			updPriceHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(updPriceHistory)
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func RemoveProductPriceById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("priceId")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductPriceById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		if unit != nil {
			remPriceHistory := request.RemoveProductPriceHistory(result.ProductId.Hex(), unit.Unit, result, userId)
			remPriceHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(remPriceHistory)
		}

		ctx.JSON(http.StatusOK, result)
	}
}
