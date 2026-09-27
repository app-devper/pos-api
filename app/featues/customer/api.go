package customer

import (
	"pos/app/domain"
	"pos/app/featues/customer/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyCustomerAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	customerRoute := route.Group("customers")

	customerRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateCustomer(repository.Customer, repository.Sequence),
	)

	customerRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		usecase.GetCustomers(repository.Customer),
	)

	customerRoute.GET("/:customerId",
		middlewares.RequireSession(repository.Auth),
		usecase.GetCustomerById(repository.Customer),
	)

	customerRoute.PUT("/:customerId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateCustomerById(repository.Customer),
	)

	customerRoute.PATCH("/:customerId/status",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateCustomerStatusById(repository.Customer),
	)

	customerRoute.DELETE("/:customerId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteCustomerById(repository.Customer),
	)

	customerRoute.GET("/code/:customerCode",
		middlewares.RequireSession(repository.Auth),
		usecase.GetCustomerByCode(repository.Customer),
	)

}
