package stock_count

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/stock_count/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockCountAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	scRoute := route.Group("stock-counts")

	scRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateStockCount(repository.StockCount, repository.StockAdjustment, repository.ProductStock, repository.Product, repository.Order, repository.Sequence),
	)

	scRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetStockCounts(repository.StockCount),
	)

	scRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetStockCountById(repository.StockCount),
	)
}
