package billing

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/billing/usecase"
	"pos/middlewares"
)

func ApplyBillingAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	billingRoute := route.Group("billings")

	billingRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateBilling(repository.Billing, repository.Sequence),
	)

	billingRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetBillings(repository.Billing),
	)

	billingRoute.GET("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetBillingById(repository.Billing),
	)

	billingRoute.PUT("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateBillingById(repository.Billing),
	)

	billingRoute.DELETE("/:id",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteBillingById(repository.Billing),
	)
}
