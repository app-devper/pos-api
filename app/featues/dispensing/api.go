package dispensing

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain"
	"pos/app/featues/dispensing/usecase"
	"pos/middlewares"
)

func ApplyDispensingAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	dispRoute := route.Group("dispensing-logs")

	dispRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreateDispensingLog(repository.DispensingLog),
	)

	dispRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDispensingLogs(repository.DispensingLog),
	)

	dispRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDispensingLogById(repository.DispensingLog),
	)

	dispRoute.GET("/patient/:patientId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetDispensingLogsByPatientId(repository.DispensingLog),
	)
}
