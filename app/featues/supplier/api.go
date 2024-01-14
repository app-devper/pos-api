package supplier

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain"
	"pos/app/featues/supplier/usecase"
	"pos/middlewares"
)

func ApplySupplierAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	supplierRoute := route.Group("suppliers")

	supplierRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.CreateSupplier(repository.Supplier),
	)

	supplierRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetSupplier(repository.Supplier),
	)

}
