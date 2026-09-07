package product_return

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/product_return/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyProductReturnAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	rtRoute := route.Group("product-returns")

	rtRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateProductReturn(repository.ProductReturn, repository.Order, repository.ProductStock, repository.Product, repository.Sequence),
	)

	rtRoute.GET("/order/:orderId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.GetProductReturnsByOrderId(repository.ProductReturn),
	)
}
