package setting

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/setting/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplySettingAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	settingRoute := route.Group("settings")

	settingRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetSetting(repository.Setting),
	)

	settingRoute.PUT("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpsertSetting(repository.Setting),
	)
}
