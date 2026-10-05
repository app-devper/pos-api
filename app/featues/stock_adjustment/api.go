package stock_adjustment

import (
	"pos/app/domain"
	"pos/app/featues/stock_adjustment/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyStockAdjustmentAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	ajRoute := route.Group("stock-adjustments")

	ajRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateStockAdjustment(repository.StockAdjustment),
	)

	ajRoute.GET("/product/:productId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetStockAdjustmentsByProductId(repository.StockAdjustment),
	)
}
