package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/constant"
	"pos/app/domain/repository"
	"pos/app/domain/request"
	"strings"
	"time"
)

func CreateProduct(productEntity repository.IProduct, receiveEntity repository.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Product{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId

		serialNumber := strings.TrimSpace(req.SerialNumber)
		product, err := productEntity.GetProductBySerialNumber(serialNumber)

		if product != nil {
			updateProduct := request.UpdateProduct{
				SerialNumber: req.SerialNumber,
				CostPrice:    req.CostPrice,
				Price:        req.Price,
				Description:  req.Description,
				Quantity:     req.Quantity + product.Quantity,
				Category:     req.Category,
				Name:         req.Name,
				NameEn:       req.NameEn,
				Unit:         req.Unit,
				UpdatedBy:    userId,
			}
			product, err = productEntity.UpdateProductById(product.Id.Hex(), updateProduct)
		} else {
			product, err = productEntity.CreateProduct(req)
			if product != nil {
				// Add product history
				_, _ = productEntity.CreateProductHistory(request.AddProductHistory(product.Id.Hex(), req))
			}
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create product unit default
		unit, _ := productEntity.GetProductUnitByDefault(product.Id.Hex(), req.Unit)
		if unit == nil {
			productUnit := request.ProductUnit{
				ProductId: product.Id.Hex(),
				Unit:      req.Unit,
				Size:      1,
				CostPrice: req.CostPrice,
				Barcode:   req.SerialNumber,
				UpdatedBy: userId,
			}
			unit, _ = productEntity.CreateProductUnit(productUnit)
			productPrice := request.ProductPrice{
				ProductId:    product.Id.Hex(),
				UnitId:       unit.Id.Hex(),
				Price:        req.Price,
				CustomerType: constant.CustomerTypeGeneral,
				UpdatedBy:    userId,
			}
			_, _ = productEntity.CreateProductPrice(productPrice)
		}

		lot, _ := productEntity.CreateProductLotByProductId(product.Id.Hex(), req)

		if req.ReceiveId != "" {
			receive, _ := receiveEntity.GetReceiveById(req.ReceiveId)
			req.ReceiveCode = receive.Code
			_, _ = receiveEntity.CreateReceiveItem(req.ReceiveId, lot.Id.Hex(), product.Id.Hex(), req)
		}

		// Create product stock
		unit, _ = productEntity.GetProductUnitByUnit(product.Id.Hex(), req.Unit)
		if req.Quantity > 0 {
			productStock := request.ProductStock{
				ProductId:   product.Id.Hex(),
				UnitId:      unit.Id.Hex(),
				ReceiveCode: req.ReceiveCode,
				Quantity:    req.Quantity,
				Price:       req.Price,
				CostPrice:   req.CostPrice,
				ExpireDate:  req.ExpireDate,
				LotNumber:   req.LotNumber,
				ImportDate:  time.Now(),
				UpdatedBy:   userId,
			}
			stock, _ := productEntity.CreateProductStock(productStock)

			if stock != nil {
				// Add product history
				balance := productEntity.GetProductStockBalance(stock.ProductId.Hex(), stock.UnitId.Hex())
				_, _ = productEntity.CreateProductHistory(request.AddProductStockHistory(stock.ProductId.Hex(), req.Unit, productStock, balance))
			}
		}

		ctx.JSON(http.StatusOK, product)
	}
}

func GetProducts(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProduct{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results, err := productEntity.GetProductAll(req)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, results)
	}
}

func GetProductBySerialNumber(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serialNumber := ctx.Param("serialNumber")
		result, err := productEntity.GetProductBySerialNumber(serialNumber)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteProductById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		result, err := productEntity.RemoveProductById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		result, err := productEntity.GetProductById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		req := request.UpdateProduct{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductById(id, req)

		_, _ = productEntity.CreateProductHistory(request.UpdateProductHistory(id, req))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
