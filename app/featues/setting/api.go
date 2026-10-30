package setting

import (
	"pos/app/domain"
	"pos/app/featues/setting/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplySettingAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	policies := middlewares.NewPolicies(repository.Auth, repository.Employee, repository.Branch)
	settingRoute := route.Group("settings")
	branchAdmin := policies.BranchAdmin.On(settingRoute)
	staff := policies.Staff.On(settingRoute)

	staff.GET("",
		usecase.GetSetting(repository.Setting),
	)

	branchAdmin.PUT("",
		usecase.UpsertSetting(repository.Setting),
	)
}
