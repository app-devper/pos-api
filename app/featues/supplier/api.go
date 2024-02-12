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
		usecase.GetSuppliers(repository.Supplier),
	)

	supplierRoute.PUT("/info",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateSupplierInfo(repository.Supplier),
	)

	supplierRoute.GET("/info",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetSupplierInfo(repository.Supplier),
	)

	supplierRoute.GET("/:supplierId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetSupplierById(repository.Supplier),
	)

	supplierRoute.DELETE("/:supplierId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.DeleteSupplierById(repository.Supplier),
	)

	supplierRoute.PUT("/:supplierId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.UpdateSupplierById(repository.Supplier),
	)

}
