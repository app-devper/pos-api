package supplier

import (
	"github.com/app-devper/um-api/sessionclient"
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
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateSupplier(repository.Supplier),
	)

	supplierRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		usecase.GetSuppliers(repository.Supplier),
	)

	supplierRoute.PUT("/info",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateSupplierInfo(repository.Supplier),
	)

	supplierRoute.GET("/info",
		middlewares.RequireSession(repository.Auth),
		usecase.GetSupplierInfo(repository.Supplier),
	)

	supplierRoute.GET("/:supplierId",
		middlewares.RequireSession(repository.Auth),
		usecase.GetSupplierById(repository.Supplier),
	)

	supplierRoute.DELETE("/:supplierId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteSupplierById(repository.Supplier),
	)

	supplierRoute.PUT("/:supplierId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateSupplierById(repository.Supplier),
	)

}
