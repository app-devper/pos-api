package receive

import (
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
		usecase.CreateReceive(repository.Receive, repository.Sequence),
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
		usecase.UpdateReceiveById(repository.Receive),
	)

	receiveRoute.DELETE("/:receiveId",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.DeleteReceiveById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/total-cost",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateReceiveTotalCostById(repository.Receive),
	)

	receiveRoute.PATCH("/:receiveId/items",
		middlewares.RequireAuthenticated(),
		middlewares.RequireSession(repository.Session),
		middlewares.RequireBranch(repository.Employee, repository.Branch),
		usecase.UpdateReceiveItemsById(repository.Receive),
	)

}
