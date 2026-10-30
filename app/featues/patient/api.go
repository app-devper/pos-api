package patient

import (
	"pos/app/domain"
	"pos/app/featues/patient/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyPatientAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	patientRoute := route.Group("patients")
	staff := policies.Staff.On(patientRoute)

	staff.POST("",
		usecase.CreatePatient(repository.Patient),
	)

	staff.GET("",
		usecase.GetPatients(repository.Patient),
	)

	staff.GET("/:id",
		usecase.GetPatientById(repository.Patient),
	)

	staff.GET("/customer/:customerCode",
		usecase.GetPatientByCustomerCode(repository.Patient),
	)

	staff.PUT("/:id",
		usecase.UpdatePatientById(repository.Patient),
	)

	staff.DELETE("/:id",
		usecase.DeletePatientById(repository.Patient),
	)

	staff.POST("/:id/allergy-check",
		usecase.AllergyCheck(repository.Patient, repository.Product),
	)
}
