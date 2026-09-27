package branch

import (
	"pos/app/domain"
	"pos/app/featues/branch/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplyBranchAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	branchRoute := route.Group("branches")

	branchRoute.POST("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.CreateBranch(repository.Branch, repository.Sequence),
	)

	branchRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetBranches(repository.Branch),
	)

	branchRoute.GET("/:branchId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.GetBranchById(repository.Branch),
	)

	branchRoute.PUT("/:branchId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateBranchById(repository.Branch),
	)

	branchRoute.PATCH("/:branchId/status",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpdateBranchStatusById(repository.Branch),
	)

	branchRoute.DELETE("/:branchId",
		middlewares.RequireSession(repository.Auth),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.DeleteBranchById(repository.Branch),
	)
}
