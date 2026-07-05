package patient

import (
	"github.com/gin-gonic/gin"
	"pos/app/domain"
	"pos/app/featues/patient/usecase"
	"pos/middlewares"
)

func ApplyPatientAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	patientRoute := route.Group("patients")

	patientRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.CreatePatient(repository.Patient),
	)

	patientRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPatients(repository.Patient),
	)

	patientRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPatientById(repository.Patient),
	)

	patientRoute.GET("/customer/:customerCode",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetPatientByCustomerCode(repository.Patient),
	)

	patientRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdatePatientById(repository.Patient),
	)

	patientRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.DeletePatientById(repository.Patient),
	)

	patientRoute.POST("/:id/allergy-check",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.AllergyCheck(repository.Patient, repository.Product),
	)
}
