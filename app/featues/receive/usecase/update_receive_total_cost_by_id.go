package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func UpdateReceiveTotalCostById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveTotalCode{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id := ctx.Param("receiveId")

		result, err := receiveEntity.UpdateReceiveTotalCostById(id, req.TotalCost)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
