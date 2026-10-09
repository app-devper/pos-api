package stock_count

import (
	"pos/app/domain"
	"pos/app/featues/stock_count/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockCountAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	scRoute := route.Group("stock-counts")
	branchAdmin := policies.BranchAdmin.On(scRoute)

	branchAdmin.POST("",
		usecase.CreateStockCount(repository.StockCount),
	)

	branchAdmin.GET("",
		usecase.GetStockCounts(repository.StockCount),
	)

	branchAdmin.GET("/:id",
		usecase.GetStockCountById(repository.StockCount),
	)
}
