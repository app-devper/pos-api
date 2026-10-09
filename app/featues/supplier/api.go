package supplier

import (
	"pos/app/domain"
	"pos/app/featues/supplier/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplySupplierAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	supplierRoute := route.Group("suppliers")
	shopAdmin := policies.ShopAdmin.On(supplierRoute)
	signedIn := policies.SignedIn.On(supplierRoute)

	shopAdmin.POST("",
		usecase.CreateSupplier(repository.Supplier),
	)

	signedIn.GET("",
		usecase.GetSuppliers(repository.Supplier),
	)

	shopAdmin.PUT("/info",
		usecase.UpdateSupplierInfo(repository.Supplier),
	)

	signedIn.GET("/info",
		usecase.GetSupplierInfo(repository.Supplier),
	)

	signedIn.GET("/:supplierId",
		usecase.GetSupplierById(repository.Supplier),
	)

	shopAdmin.DELETE("/:supplierId",
		usecase.DeleteSupplierById(repository.Supplier),
	)

	shopAdmin.PUT("/:supplierId",
		usecase.UpdateSupplierById(repository.Supplier),
	)

}
