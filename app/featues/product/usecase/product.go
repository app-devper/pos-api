package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func CreateProductReceive(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Product{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId
		req.BranchId = ctx.GetString("BranchId")
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
		req.BranchId = ctx.GetString("BranchId")
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

func GetProductById(productEntity repositories.IProduct, productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		product, err := productEntity.GetProductById(id)

		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		branchId := ctx.GetString("BranchId")
		units, err := productEntity.GetProductUnitsByProductId(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		prices, err := productEntity.GetProductPricesByProductId(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		stocks, err := productStock.GetProductStocksByProductId(id, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, entities.ProductDetail{
			Id:                product.Id,
			Name:              product.Name,
			NameEn:            product.NameEn,
			Description:       product.Description,
			Price:             product.Price,
			CostPrice:         product.CostPrice,
			Unit:              product.Unit,
			Quantity:          product.Quantity,
			SoldFirst:         product.SoldFirst,
			SerialNumber:      product.SerialNumber,
			Category:          product.Category,
			Status:            product.Status,
			MinStock:          product.MinStock,
			DrugInfo:          product.DrugInfo,
			DrugRegistrations: product.DrugRegistrations,
			DeletedDate:       product.DeletedDate,
			CreatedBy:         product.CreatedBy,
			CreatedDate:       product.CreatedDate,
			UpdatedBy:         product.UpdatedBy,
			UpdatedDate:       product.UpdatedDate,
			ProductUnits:      units,
			ProductPrices:     prices,
			ProductStocks:     stocks,
		})
	}
}

func UpdateProductById(productEntity repositories.IProduct, productStock repositories.IProductStock) gin.HandlerFunc {
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

		appendProductMasterHistory(ctx.GetString("BranchId"), productStock, func(branchId string) request.ProductHistory {
			updProdHistory := request.UpdateProductHistory(id, req)
			updProdHistory.BranchId = branchId
			return updProdHistory
		})

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

func appendProductMasterHistory(branchId string, productStock repositories.IProductStock, buildHistory func(branchId string) request.ProductHistory) {
	history := buildHistory(branchId)
	if _, err := productStock.CreateProductHistory(history); err != nil {
		logrus.WithError(err).WithField("branchId", branchId).Error("failed to create product history")
	}
}
