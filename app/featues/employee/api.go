package employee

import (
	"pos/app/domain"
	"pos/app/featues/employee/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyEmployeeAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	employeeRoute := route.Group("employees")

	employeeRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateEmployee(repository.Employee),
	)

	employeeRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetEmployees(repository.Employee),
	)

	employeeRoute.GET("/:employeeId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetEmployeeById(repository.Employee),
	)

	employeeRoute.PUT("/:employeeId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateEmployeeById(repository.Employee),
	)

	employeeRoute.DELETE("/:employeeId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteEmployeeById(repository.Employee),
	)

	employeeRoute.GET("/branch/:branchId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetEmployeesByBranchId(repository.Employee),
	)
}
