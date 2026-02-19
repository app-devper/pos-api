package employee

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/employee/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyEmployeeAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	employeeRoute := route.Group("employees")

	employeeRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateEmployee(repository.Employee),
	)

	employeeRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetEmployees(repository.Employee),
	)

	employeeRoute.GET("/:employeeId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetEmployeeById(repository.Employee),
	)

	employeeRoute.PUT("/:employeeId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateEmployeeById(repository.Employee),
	)

	employeeRoute.DELETE("/:employeeId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteEmployeeById(repository.Employee),
	)

	employeeRoute.GET("/branch/:branchId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetEmployeesByBranchId(repository.Employee),
	)
}
