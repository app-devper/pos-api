package customer_history

import (
	"pos/app/domain"
	"pos/app/featues/customer_history/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyCustomerHistoryAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	chRoute := route.Group("customer-histories")
	staff := policies.Staff.On(chRoute)

	staff.POST("",
		usecase.CreateCustomerHistory(repository.CustomerHistory),
	)

	staff.GET("/:customerCode",
		usecase.GetCustomerHistories(repository.CustomerHistory),
	)
}
