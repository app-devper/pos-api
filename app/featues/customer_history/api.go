package customer_history

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain"
	"pos/app/featues/customer_history/usecase"
	"pos/middlewares"
)

func ApplyCustomerHistoryAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	chRoute := route.Group("customer-histories")

	chRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateCustomerHistory(repository.CustomerHistory),
	)

	chRoute.GET("/:customerCode",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetCustomerHistories(repository.CustomerHistory),
	)
}
