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
		middlewares.RequireBranch(repository.Employee),
		usecase.CreatePatient(repository.Patient),
	)

	patientRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.GetPatients(repository.Patient),
	)

	patientRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.GetPatientById(repository.Patient),
	)

	patientRoute.GET("/customer/:customerCode",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.GetPatientByCustomerCode(repository.Patient),
	)

	patientRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.UpdatePatientById(repository.Patient),
	)

	patientRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.DeletePatientById(repository.Patient),
	)

	patientRoute.POST("/:id/allergy-check",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee),
		usecase.AllergyCheck(repository.Patient, repository.Product),
	)
}
