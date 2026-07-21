package stock_adjustment

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/stock_adjustment/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyStockAdjustmentAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	ajRoute := route.Group("stock-adjustments")

	ajRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateStockAdjustment(repository.StockAdjustment, repository.ProductStock, repository.Product, repository.Order, repository.Sequence),
	)

	ajRoute.GET("/product/:productId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetStockAdjustmentsByProductId(repository.StockAdjustment),
	)
}
