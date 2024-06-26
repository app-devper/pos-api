package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"sort"
	"strings"
	"time"
)

func GenerateSerialNumber(sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := sequenceEntity.NextSequence(constant.PRODUCT)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"serialNumber": result.GenerateCode()})
	}
}

func CreateProduct(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CreateProduct{}
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
				Description: req.Description,
				Category:    req.Category,
				Name:        req.Name,
				NameEn:      req.NameEn,
				Status:      req.Status,
				UpdatedBy:   userId,
			}
			product, err = productEntity.UpdateProductById(product.Id.Hex(), updateProduct)

			// Add product history
			_, _ = productEntity.CreateProductHistory(request.UpdateProductHistory(product.Id.Hex(), updateProduct))
		} else {
			createProduct := request.Product{
				SerialNumber: req.SerialNumber,
				CostPrice:    req.CostPrice,
				Price:        req.Price,
				Description:  req.Description,
				Status:       req.Status,
				Quantity:     0,
				Category:     req.Category,
				Name:         req.Name,
				NameEn:       req.NameEn,
				Unit:         req.Unit,
				CreatedBy:    userId,
			}
			product, err = productEntity.CreateProduct(createProduct)
			if product != nil {

				// Add product history
				_, _ = productEntity.CreateProductHistory(request.AddProductHistory(product.Id.Hex(), createProduct))
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

		ctx.JSON(http.StatusOK, product)
	}
}

func CreateProductReceive(productEntity repositories.IProduct, receiveEntity repositories.IReceive) gin.HandlerFunc {
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
				Description: req.Description,
				Category:    req.Category,
				Name:        req.Name,
				NameEn:      req.NameEn,
				UpdatedBy:   userId,
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

		if req.ReceiveId != "" {
			receive, _ := receiveEntity.GetReceiveById(req.ReceiveId)
			req.ReceiveCode = receive.Code
			_, _ = receiveEntity.CreateReceiveItem(req.ReceiveId, "", product.Id.Hex(), req)
		}

		// Create product stock
		unit, _ = productEntity.GetProductUnitByUnit(product.Id.Hex(), req.Unit)
		if req.Quantity > 0 {
			productStock := request.ProductStock{
				ProductId:   product.Id.Hex(),
				UnitId:      unit.Id.Hex(),
				ReceiveCode: req.ReceiveCode,
				Quantity:    req.Quantity,
				Price:       0,
				CostPrice:   0,
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

func GetProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProduct{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results, err := productEntity.GetProductAll(req)
		userId := ctx.GetString("UserId")

		for _, product := range results {
			if len(product.ProductUnits) == 0 {
				productUnit := request.ProductUnit{
					ProductId: product.Id.Hex(),
					Unit:      product.Unit,
					Size:      1,
					CostPrice: product.CostPrice,
					Barcode:   product.SerialNumber,
					UpdatedBy: userId,
				}
				unit, _ := productEntity.CreateProductUnit(productUnit)
				productPrice := request.ProductPrice{
					ProductId:    product.Id.Hex(),
					UnitId:       unit.Id.Hex(),
					Price:        product.Price,
					CustomerType: constant.CustomerTypeGeneral,
					UpdatedBy:    userId,
				}
				_, _ = productEntity.CreateProductPrice(productPrice)

				if product.Quantity > 0 {
					lots, _ := productEntity.GetProductLotsByProductId(product.Id.Hex())
					if len(lots) > 0 {
						sort.Slice(lots, func(i, j int) bool {
							return lots[i].ExpireDate.After(lots[j].ExpireDate)
						})
						lot := lots[0]
						productStock := request.ProductStock{
							ProductId:   product.Id.Hex(),
							UnitId:      unit.Id.Hex(),
							ReceiveCode: "",
							Quantity:    product.Quantity,
							Price:       0,
							CostPrice:   0,
							ExpireDate:  lot.ExpireDate,
							LotNumber:   lot.LotNumber,
							ImportDate:  lot.CreatedDate,
							UpdatedBy:   userId,
						}
						_, _ = productEntity.CreateProductStock(productStock)
					}
					if len(lots) == 0 {
						productStock := request.ProductStock{
							ProductId:   product.Id.Hex(),
							UnitId:      unit.Id.Hex(),
							ReceiveCode: "",
							Quantity:    product.Quantity,
							Price:       0,
							CostPrice:   0,
							ExpireDate:  time.Now(),
							LotNumber:   "",
							ImportDate:  time.Now(),
							UpdatedBy:   userId,
						}
						_, _ = productEntity.CreateProductStock(productStock)
					}
				}
			}
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func ClearQuantitySoldFirstById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.ClearQuantitySoldFirstById(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
