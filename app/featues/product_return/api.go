package product_return

import (
	"pos/app/domain"
	"pos/app/featues/product_return/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyProductReturnAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	rtRoute := route.Group("product-returns")

	rtRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateProductReturn(repository.ProductReturn),
	)

	rtRoute.GET("/order/:orderId",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetProductReturnsByOrderId(repository.ProductReturn),
	)
}
