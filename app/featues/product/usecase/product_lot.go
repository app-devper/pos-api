package usecase

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"
)

func CreateProductLot(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductLot{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.CreateProductLot(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductLotByLotId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lotId := ctx.Param("lotId")

		result, err := productEntity.GetProductLotById(lotId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductLots(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProductLotsExpireRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := productEntity.GetProductLots(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductLotsByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")

		result, err := productEntity.GetProductLotsByProductId(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductLotsExpireNotify(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		location := utils.GetLocation()
		today := time.Now().In(location)
		startDate := utils.Bod(today)
		endDate := startDate.Add(24 * time.Hour)
		req := request.GetProductLotsExpireRange{
			StartDate: startDate.UTC(),
			EndDate:   endDate.UTC(),
		}
		result, err := productEntity.GetProductLotsExpireNotify(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		}

		if len(result) > 0 {
			var message = ""
			var no = 1
			for _, item := range result {
				message += fmt.Sprintf("%d. %s Lot: %s\n", no, item.Product.Name, item.LotNumber)
				no += 1
			}

			date := utils.ToFormatDate(today)
			_, _ = utils.NotifyMassage("สินค้าหมดอายุวันที่ " + date + "\n\n" + message)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success",
		})

	}
}

func GetProductLotsExpired(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := productEntity.GetProductLotsExpired()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductLotByLotId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("lotId")
		req := request.UpdateProductLot{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductLotById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductLotNotifyByLotId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("lotId")
		req := request.UpdateProductLotNotify{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductLotNotifyById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductLotQuantityByLotId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("lotId")
		req := request.UpdateProductLotQuantity{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductLotQuantityById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
