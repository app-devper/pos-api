package branch

import (
	"github.com/gin-gonic/gin"
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/branch/usecase"
	"pos/middlewares"
)

func ApplyBranchAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	branchRoute := route.Group("branches")

	branchRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.CreateBranch(repository.Branch, repository.Sequence),
	)

	branchRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetBranches(repository.Branch),
	)

	branchRoute.GET("/:branchId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		usecase.GetBranchById(repository.Branch),
	)

	branchRoute.PUT("/:branchId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateBranchById(repository.Branch),
	)

	branchRoute.PATCH("/:branchId/status",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.UpdateBranchStatusById(repository.Branch),
	)

	branchRoute.DELETE("/:branchId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireAuthorization(constant.ADMIN),
		usecase.DeleteBranchById(repository.Branch),
	)
}
