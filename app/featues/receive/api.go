package receive

import (
	"pos/app/core/constant"
	"pos/app/domain"
	"pos/app/featues/receive/usecase"
	"pos/middlewares"

	"github.com/gin-gonic/gin"
)

func ApplyReceiveAPI(
	route *gin.RouterGroup,
	repository *domain.Repository,
) {
	receiveRoute := route.Group("receives")

	receiveRoute.POST("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.CreateReceive(repository.Receive, repository.Sequence, repository.Product),
	)

	receiveRoute.GET("",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetReceivesRange(repository.Receive),
	)

	receiveRoute.GET("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.GetReceiveById(repository.Receive),
	)

	receiveRoute.PUT("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateReceiveById(repository.Receive),
	)

	receiveRoute.DELETE("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.DeleteReceiveById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateReceiveTotalCostById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/items",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		middlewares.RequireAuthorization(constant.ADMIN, constant.SUPER),
		usecase.UpdateReceiveItemsById(repository.Receive),
	)

}
