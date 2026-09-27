package setting

import (
	"pos/app/domain"
	"pos/app/featues/setting/usecase"
	"pos/middlewares"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

func ApplySettingAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	settingRoute := route.Group("settings")

	settingRoute.GET("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetSetting(repository.Setting),
	)

	settingRoute.PUT("",
		middlewares.RequireSession(repository.Auth),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		repository.Auth.AtLeast(sessionclient.RoleAdmin),
		usecase.UpsertSetting(repository.Setting),
	)
}
