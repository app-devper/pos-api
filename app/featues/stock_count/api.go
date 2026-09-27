package stock_count

import (
	"pos/app/domain"
	"pos/app/featues/stock_count/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyStockCountAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	scRoute := route.Group("stock-counts")

	scRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateStockCount(repository.StockCount, repository.StockAdjustment, repository.ProductStock, repository.Product, repository.Order, repository.Sequence),
	)

	scRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockCounts(repository.StockCount),
	)

	scRoute.GET("/:id",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockCountById(repository.StockCount),
	)
}
