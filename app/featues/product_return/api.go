package product_return

import (
	"pos/app/domain"
	"pos/app/featues/product_return/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyProductReturnAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	rtRoute := route.Group("product-returns")
	branchAdmin := policies.BranchAdmin.On(rtRoute)

	branchAdmin.POST("",
		usecase.CreateProductReturn(repository.Ledger),
	)

	branchAdmin.GET("/order/:orderId",
		usecase.GetProductReturnsByOrderId(repository.ProductReturn),
	)
}
