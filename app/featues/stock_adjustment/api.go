package stock_adjustment

import (
	"pos/app/domain"
	"pos/app/featues/stock_adjustment/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockAdjustmentAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	ajRoute := route.Group("stock-adjustments")
	branchAdmin := policies.BranchAdmin.On(ajRoute)

	branchAdmin.POST("",
		usecase.CreateStockAdjustment(repository.Ledger),
	)

	branchAdmin.GET("/product/:productId",
		usecase.GetStockAdjustmentsByProductId(repository.StockAdjustment),
	)
}
