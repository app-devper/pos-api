package branch

import (
	"pos/app/domain"
	"pos/app/featues/branch/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyBranchAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	branchRoute := route.Group("branches")
	shopAdmin := policies.ShopAdmin.On(branchRoute)

	shopAdmin.POST("",
		usecase.CreateBranch(repository.Branch, repository.Sequence),
	)

	shopAdmin.GET("",
		usecase.GetBranches(repository.Branch),
	)

	shopAdmin.GET("/:branchId",
		usecase.GetBranchById(repository.Branch),
	)

	shopAdmin.PUT("/:branchId",
		usecase.UpdateBranchById(repository.Branch),
	)

	shopAdmin.PATCH("/:branchId/status",
		usecase.UpdateBranchStatusById(repository.Branch),
	)

	shopAdmin.DELETE("/:branchId",
		usecase.DeleteBranchById(repository.Branch),
	)
}
