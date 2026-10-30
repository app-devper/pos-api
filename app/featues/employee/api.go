package employee

import (
	"pos/app/domain"
	"pos/app/featues/employee/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyEmployeeAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	employeeRoute := route.Group("employees")
	shopAdmin := policies.ShopAdmin.On(employeeRoute)

	shopAdmin.POST("",
		usecase.CreateEmployee(repository.Employee),
	)

	shopAdmin.GET("",
		usecase.GetEmployees(repository.Employee),
	)

	shopAdmin.GET("/:employeeId",
		usecase.GetEmployeeById(repository.Employee),
	)

	shopAdmin.PUT("/:employeeId",
		usecase.UpdateEmployeeById(repository.Employee),
	)

	shopAdmin.DELETE("/:employeeId",
		usecase.DeleteEmployeeById(repository.Employee),
	)

	shopAdmin.GET("/branch/:branchId",
		usecase.GetEmployeesByBranchId(repository.Employee),
	)
}
