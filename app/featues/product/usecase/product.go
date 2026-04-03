package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"strings"

	"github.com/gin-gonic/gin"
)

func GenerateSerialNumber(sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := sequenceEntity.NextSequence(constant.PRODUCT)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"serialNumber": result.GenerateCode()})
	}
}

func CreateProduct(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CreateProduct{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId
		product, err := productEntity.CreateProductCatalog(request.Product{
			SerialNumber:      strings.TrimSpace(req.SerialNumber),
			CostPrice:         req.CostPrice,
			Price:             req.Price,
			Description:       req.Description,
			Status:            req.Status,
			Quantity:          0,
			Category:          req.Category,
			Name:              req.Name,
			NameEn:            req.NameEn,
			Unit:              req.Unit,
			DrugInfo:          req.DrugInfo,
			DrugRegistrations: req.DrugRegistrations,
			CreatedBy:         userId,
			BranchId:          ctx.GetString("BranchId"),
			MinStock:          req.MinStock,
		})

		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, product)
	}
}

func CreateProductReceive(productEntity repositories.IProduct, receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Product{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId
		req.BranchId = ctx.GetString("BranchId")
		_ = receiveEntity
		product, err := productEntity.CreateProductReceive(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, product)
	}
}

func GetProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProduct{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		results, err := productEntity.GetProductAll(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, results)
	}
}

func GetProductBySerialNumber(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serialNumber := ctx.Param("serialNumber")
		result, err := productEntity.GetProductBySerialNumber(serialNumber)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteProductById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		product, err := productEntity.RemoveProductById(id)

		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, product)
	}
}

func GetProductById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		product, err := productEntity.GetProductById(id)

		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, product)
	}
}

func UpdateProductById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		req := request.UpdateProduct{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		updProdHistory := request.UpdateProductHistory(id, req)
		updProdHistory.BranchId = ctx.GetString("BranchId")
		_, _ = productEntity.CreateProductHistory(updProdHistory)

		ctx.JSON(http.StatusOK, result)
	}
}

func ClearQuantitySoldFirstById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.ClearQuantitySoldFirstById(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
