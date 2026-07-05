package customer

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/customer/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyCustomerAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	customerRoute := route.Group("customers")

	customerRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateCustomer(repository.Customer, repository.Sequence),
	)

	customerRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetCustomers(repository.Customer),
	)

	customerRoute.GET("/:customerId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetCustomerById(repository.Customer),
	)

	customerRoute.PUT("/:customerId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateCustomerById(repository.Customer),
	)

	customerRoute.PATCH("/:customerId/status",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateCustomerStatusById(repository.Customer),
	)

	customerRoute.DELETE("/:customerId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.DeleteCustomerById(repository.Customer),
	)

	customerRoute.GET("/code/:customerCode",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetCustomerByCode(repository.Customer),
	)

}
