package customer

import (
	"pos/app/domain"
	"pos/app/featues/customer/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyCustomerAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	customerRoute := route.Group("customers")
	shopAdmin := policies.ShopAdmin.On(customerRoute)
	signedIn := policies.SignedIn.On(customerRoute)

	shopAdmin.POST("",
		usecase.CreateCustomer(repository.Customer, repository.Sequence),
	)

	signedIn.GET("",
		usecase.GetCustomers(repository.Customer),
	)

	signedIn.GET("/:customerId",
		usecase.GetCustomerById(repository.Customer),
	)

	shopAdmin.PUT("/:customerId",
		usecase.UpdateCustomerById(repository.Customer),
	)

	shopAdmin.PATCH("/:customerId/status",
		usecase.UpdateCustomerStatusById(repository.Customer),
	)

	shopAdmin.DELETE("/:customerId",
		usecase.DeleteCustomerById(repository.Customer),
	)

	signedIn.GET("/code/:customerCode",
		usecase.GetCustomerByCode(repository.Customer),
	)

}
